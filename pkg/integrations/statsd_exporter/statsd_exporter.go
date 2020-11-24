// Package statsd_exporter embeds https://github.com/prometheus/statsd_exporter
package statsd_exporter //nolint:golint

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/gorilla/mux"
	"github.com/grafana/agent/pkg/integrations/common"
	"github.com/grafana/agent/pkg/integrations/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	"github.com/prometheus/statsd_exporter/pkg/address"
	"github.com/prometheus/statsd_exporter/pkg/event"
	"github.com/prometheus/statsd_exporter/pkg/exporter"
	"github.com/prometheus/statsd_exporter/pkg/line"
	"github.com/prometheus/statsd_exporter/pkg/listener"
	"github.com/prometheus/statsd_exporter/pkg/mapper"
	"gopkg.in/yaml.v2"
)

var DefaultConfig = Config{
	ListenUDP:      ":9125",
	ListenTCP:      ":9125",
	UnixSocketMode: "755",

	CacheSize:           1000,
	CacheType:           "lru",
	EventQueueSize:      10000,
	EventFlushThreshold: 1000,
	EventFlushInterval:  200 * time.Millisecond,

	ParseDogStatsd: true,
	ParseInfluxDB:  true,
	ParseLibrato:   true,
	ParseSignalFX:  true,
}

// Config controls the statsd_exporter integration.
type Config struct {
	// Enabled enables the integration.
	Enabled bool `yaml:"enabled"`

	CommonConfig config.Common `yaml:",inline"`

	ListenUDP      string               `yaml:"listen_udp"`
	ListenTCP      string               `yaml:"listen_tcp"`
	ListenUnixgram string               `yaml:"listen_unixgram"`
	UnixSocketMode string               `yaml:"unix_socket_mode"`
	MappingConfig  *mapper.MetricMapper `yaml:"mapping_config"`

	ReadBuffer          int           `yaml:"read_buffer"`
	CacheSize           int           `yaml:"cache_size"`
	CacheType           string        `yaml:"cache_type"`
	EventQueueSize      int           `yaml:"event_queue_size"`
	EventFlushThreshold int           `yaml:"event_flush_threshold"`
	EventFlushInterval  time.Duration `yaml:"event_flush_interval"`

	ParseDogStatsd bool `yaml:"parse_dogstatsd_tags"`
	ParseInfluxDB  bool `yaml:"parse_influxdb_tags"`
	ParseLibrato   bool `yaml:"parse_librato_tags"`
	ParseSignalFX  bool `yaml:"parse_signalfx_tags"`
}

// UnmarshalYAML implements yaml.Unmarshaler for Config.
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultConfig

	type plain Config
	return unmarshal((*plain)(c))
}

// Exporters defines the statsd_exporter integration.
type Exporter struct {
	cfg      Config
	reg      *prometheus.Registry
	metrics  *Metrics
	exporter *exporter.Exporter
	log      log.Logger
}

// New creates a new statsd_exporter integration. The integration scrapes
// metrics from a statsd process.
func New(log log.Logger, c Config) (common.Integration, error) {
	reg := prometheus.NewRegistry()

	m, err := NewMetrics(reg)
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics for network listeners: %w", err)
	}

	cacheOption := mapper.WithCacheType(c.CacheType)

	if c.ListenUDP == "" && c.ListenTCP == "" && c.ListenUnixgram == "" {
		return nil, fmt.Errorf("at least one of UDP/TCP/Unixgram listeners must be used")
	}

	mapper := &mapper.MetricMapper{MappingsCount: m.MappingsCount}
	if c.MappingConfig != nil {
		cfgBytes, err := yaml.Marshal(c.MappingConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize mapping config: %w", err)
		}

		err = mapper.InitFromYAMLString(string(cfgBytes), c.CacheSize, cacheOption)
		if err != nil {
			return nil, fmt.Errorf("failed to load mapping config: %w", err)
		}
	} else {
		mapper.InitCache(c.CacheSize, cacheOption)
	}

	e := exporter.NewExporter(reg, mapper, log, m.EventsActions, m.EventsUnmapped, m.ErrorEventStats, m.EventStats, m.ConflictingEventStats, m.MetricsCount)

	if err := reg.Register(version.NewCollector("statsd_exporter")); err != nil {
		return nil, fmt.Errorf("couldn't register version metrics: %w", err)
	}

	return &Exporter{
		cfg:      c,
		metrics:  m,
		exporter: e,
		reg:      reg,
		log:      log,
	}, nil
}

// Name satisfies common.Integration.
func (e *Exporter) Name() string { return "statsd_exporter" }

// CommonConfig satisfies common.Integration.
func (e *Exporter) CommonConfig() config.Common { return e.cfg.CommonConfig }

// RegisterRoutes satisfies common.Integration. The mux.Router provided
// here is expected to be a subrouter, where all registered paths will be
// registered within that subroute.
func (e *Exporter) RegisterRoutes(r *mux.Router) error {
	handler := promhttp.HandlerFor(e.reg, promhttp.HandlerOpts{
		ErrorHandling: promhttp.ContinueOnError,
	})

	r.Handle("/metrics", handler)
	return nil
}

// ScrapeConfigs satisfies Integration.ScrapeConfigs.
func (e *Exporter) ScrapeConfigs() []config.ScrapeConfig {
	return []config.ScrapeConfig{{JobName: e.Name(), MetricsPath: "/metrics"}}
}

// Run satisfies Run.
func (e *Exporter) Run(ctx context.Context) error {
	parser := line.NewParser()
	if e.cfg.ParseDogStatsd {
		parser.EnableDogstatsdParsing()
	}
	if e.cfg.ParseInfluxDB {
		parser.EnableInfluxdbParsing()
	}
	if e.cfg.ParseLibrato {
		parser.EnableLibratoParsing()
	}
	if e.cfg.ParseSignalFX {
		parser.EnableSignalFXParsing()
	}

	events := make(chan event.Events, e.cfg.EventQueueSize)
	defer close(events)
	eventQueue := event.NewEventQueue(events, e.cfg.EventFlushThreshold, e.cfg.EventFlushInterval, e.metrics.EventsFlushed)

	if e.cfg.ListenUDP != "" {
		addr, err := address.UDPAddrFromString(e.cfg.ListenUDP)
		if err != nil {
			return fmt.Errorf("invalid UDP listen address %s: %w", e.cfg.ListenUDP, err)
		}
		uconn, err := net.ListenUDP("udp", addr)
		if err != nil {
			return fmt.Errorf("failed to start UDP listener: %w", err)
		}
		defer func() {
			err := uconn.Close()
			if err != nil {
				level.Warn(e.log).Log("msg", "failed to close UDP listener", "err", err)
			}
		}()

		if e.cfg.ReadBuffer != 0 {
			err = uconn.SetReadBuffer(e.cfg.ReadBuffer)
			if err != nil {
				return fmt.Errorf("failed to set UDP read buffer: %w", err)
			}
		}

		ul := &listener.StatsDUDPListener{
			Conn:            uconn,
			EventHandler:    eventQueue,
			Logger:          e.log,
			LineParser:      parser,
			UDPPackets:      e.metrics.UDPPackets,
			LinesReceived:   e.metrics.LinesReceived,
			EventsFlushed:   e.metrics.EventsFlushed,
			SampleErrors:    *e.metrics.SampleErrors,
			SamplesReceived: e.metrics.SamplesReceived,
			TagErrors:       e.metrics.TagErrors,
			TagsReceived:    e.metrics.TagsReceived,
		}

		go ul.Listen()
	}

	if e.cfg.ListenTCP != "" {
		addr, err := address.TCPAddrFromString(e.cfg.ListenTCP)
		if err != nil {
			return fmt.Errorf("invalid TCP listen address %s: %w", e.cfg.ListenTCP, err)
		}
		tconn, err := net.ListenTCP("tcp", addr)
		if err != nil {
			return fmt.Errorf("failed to start TCP listener: %w", err)
		}
		defer func() {
			err := tconn.Close()
			if err != nil {
				level.Warn(e.log).Log("msg", "failed to close TCP listener", "err", err)
			}
		}()

		tl := &listener.StatsDTCPListener{
			Conn:            tconn,
			EventHandler:    eventQueue,
			Logger:          e.log,
			LineParser:      parser,
			LinesReceived:   e.metrics.LinesReceived,
			EventsFlushed:   e.metrics.EventsFlushed,
			SampleErrors:    *e.metrics.SampleErrors,
			SamplesReceived: e.metrics.SamplesReceived,
			TagErrors:       e.metrics.TagErrors,
			TagsReceived:    e.metrics.TagsReceived,
			TCPConnections:  e.metrics.TCPConnections,
			TCPErrors:       e.metrics.TCPErrors,
			TCPLineTooLong:  e.metrics.TCPLineTooLong,
		}

		go tl.Listen()
	}

	if e.cfg.ListenUnixgram != "" {
		var err error
		if _, err = os.Stat(e.cfg.ListenUnixgram); !os.IsNotExist(err) {
			return fmt.Errorf("unixgram socket %s already exists: %w", e.cfg.ListenUnixgram, err)
		}
		uxgconn, err := net.ListenUnixgram("unixgram", &net.UnixAddr{
			Net:  "unixgram",
			Name: e.cfg.ListenUnixgram,
		})
		if err != nil {
			return fmt.Errorf("failed to listen on unixgram socket: %w", err)
		}
		defer func() {
			err := uxgconn.Close()
			if err != nil {
				level.Warn(e.log).Log("msg", "failed to close unixgram listener", "err", err)
			}
		}()

		if e.cfg.ReadBuffer != 0 {
			err = uxgconn.SetReadBuffer(e.cfg.ReadBuffer)
			if err != nil {
				return fmt.Errorf("error setting unixgram read buffer: %w", err)
			}
		}

		ul := &listener.StatsDUnixgramListener{
			Conn:            uxgconn,
			EventHandler:    eventQueue,
			Logger:          e.log,
			LineParser:      parser,
			UnixgramPackets: e.metrics.UnixgramPackets,
			LinesReceived:   e.metrics.LinesReceived,
			EventsFlushed:   e.metrics.EventsFlushed,
			SampleErrors:    *e.metrics.SampleErrors,
			SamplesReceived: e.metrics.SamplesReceived,
			TagErrors:       e.metrics.TagErrors,
			TagsReceived:    e.metrics.TagsReceived,
		}

		go ul.Listen()

		// If it's an abstract unix domain socket, it won't exist on fs so we can't
		// chmod it either.
		if _, err := os.Stat(e.cfg.ListenUnixgram); !os.IsNotExist(err) {
			defer os.Remove(e.cfg.ListenUnixgram)

			// Convert the string to octet
			perm, err := strconv.ParseInt("0"+e.cfg.UnixSocketMode, 8, 32)
			if err != nil {
				level.Warn(e.log).Log("msg", "bad permission on unixgram socket, ignoring", "permission", e.cfg.UnixSocketMode, "socket", e.cfg.ListenUnixgram, "err", err)
			} else {
				err = os.Chmod(e.cfg.ListenUnixgram, os.FileMode(perm))
				if err != nil {
					level.Warn(e.log).Log("msg", "failed to change unixgram socket permission", "socket", e.cfg.ListenUnixgram, "err", err)
				}
			}
		}
	}

	go e.exporter.Listen(events)

	<-ctx.Done()
	return nil
}
