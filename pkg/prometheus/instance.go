package prometheus

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/grafana/agent/pkg/prometheus/wal"
	"github.com/oklog/run"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/discovery"
	sd_config "github.com/prometheus/prometheus/discovery/config"
	"github.com/prometheus/prometheus/pkg/relabel"
	"github.com/prometheus/prometheus/scrape"
	"github.com/prometheus/prometheus/storage"
	"github.com/prometheus/prometheus/storage/remote"
)

var (
	errInstanceStoppedNormally = errors.New("instance shutdown normally")
	remoteWriteMetricName      = "queue_highest_sent_timestamp_seconds"
)

// Default configuration values
var (
	// DefaultRelabelConfigs defines a list of relabel_configs that will
	// be automatically appended to the end of all Prometheus
	// configurations.
	DefaultRelabelConfigs = []*relabel.Config{
		// Add __host__ from Kubernetes node name
		{
			SourceLabels: model.LabelNames{"__meta_kubernetes_pod_node_name"},
			TargetLabel:  "__host__",
			Action:       relabel.Replace,
			Separator:    ";",
			Regex:        relabel.MustNewRegexp("(.*)"),
			Replacement:  "$1",
		},
	}

	DefaultInstanceConfig = InstanceConfig{
		HostFilter:           true,
		WALTruncateFrequency: 1 * time.Minute,
		RemoteFlushDeadline:  1 * time.Minute,
		WriteStaleOnShutdown: true,
	}
)

// InstanceConfig is a specific agent that runs within the overall Prometheus
// agent. It has its own set of scrape_configs and remote_write rules.
type InstanceConfig struct {
	Name          string                      `yaml:"name"`
	HostFilter    bool                        `yaml:"host_filter"`
	ScrapeConfigs []*config.ScrapeConfig      `yaml:"scrape_configs,omitempty"`
	RemoteWrite   []*config.RemoteWriteConfig `yaml:"remote_write,omitempty"`

	// How frequently the WAL should be truncated.
	WALTruncateFrequency time.Duration `yaml:"wal_truncate_frequency,omitempty"`

	RemoteFlushDeadline  time.Duration `yaml:"remote_flush_deadline,omitempty"`
	WriteStaleOnShutdown bool          `yaml:"write_stale_on_shutdown,omitempty"`
}

func (c *InstanceConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultInstanceConfig

	type plain InstanceConfig
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}

	// We need to be able to pull the metrics of the last written timestamp for all the remote_writes.
	// This can currently only be done using the remote_name label Prometheus adds to the
	// prometheus_remote_storage_queue_highest_sent_time_seconds label, so we force a name here
	// that we can lookup later.
	for _, cfg := range c.RemoteWrite {
		if cfg.Name != "" {
			continue
		}

		hash, err := getHash(cfg)
		if err != nil {
			return err
		}

		// We have to add the name of the instance to ensure that generated metrics
		// are unique across multiple agent instances. The remote write queues currently
		// globally register their metrics so we can't inject labels here.
		cfg.Name = c.Name + "-" + hash[:6]
	}

	return nil
}

// ApplyDefaults applies default configurations to the configuration to all
// values that have not been changed to their non-zero value.
func (c *InstanceConfig) ApplyDefaults(global *config.GlobalConfig) {
	for _, sc := range c.ScrapeConfigs {
		if sc.ScrapeInterval == 0 {
			sc.ScrapeInterval = global.ScrapeInterval
		}
		if sc.ScrapeTimeout == 0 {
			if global.ScrapeTimeout > sc.ScrapeInterval {
				sc.ScrapeTimeout = sc.ScrapeInterval
			} else {
				sc.ScrapeTimeout = global.ScrapeTimeout
			}
		}

		sc.RelabelConfigs = append(sc.RelabelConfigs, DefaultRelabelConfigs...)
	}
}

// Validate checks if the InstanceConfig has all required fields filled out.
// This should only be called after ApplyDefaults.
func (c *InstanceConfig) Validate() error {
	if c.Name == "" {
		return errors.New("missing instance name")
	}

	return nil
}

// Instance is an individual metrics collector and remote_writer.
type Instance struct {
	cfg       InstanceConfig
	globalCfg config.GlobalConfig
	logger    log.Logger

	walDir   string
	hostname string

	cancelScrape context.CancelFunc
	vc           *MetricValueCollector

	exited  chan bool
	exitErr error
}

// NewInstance creates and starts a new Instance. NewInstance creates a WAL in
// a folder with the same name as the instance's name in a subdirectory of the
// walDir parameter.
func NewInstance(globalCfg config.GlobalConfig, cfg InstanceConfig, walDir string, logger log.Logger) (*Instance, error) {
	logger = log.With(logger, "instance", cfg.Name)

	instWALDir := filepath.Join(walDir, cfg.Name)

	reg := prometheus.WrapRegistererWith(prometheus.Labels{
		"instance": cfg.Name,
	}, prometheus.DefaultRegisterer)

	wstore, err := wal.NewStorage(logger, reg, instWALDir)
	if err != nil {
		return nil, err
	}

	return newInstance(globalCfg, cfg, reg, instWALDir, logger, wstore)
}

func newInstance(globalCfg config.GlobalConfig, cfg InstanceConfig, reg prometheus.Registerer, walDir string, logger log.Logger, wstore walStorage) (*Instance, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	hostname, err := hostname()
	if err != nil {
		return nil, err
	}

	vc := NewMetricValueCollector(prometheus.DefaultGatherer, remoteWriteMetricName)

	i := &Instance{
		cfg:       cfg,
		globalCfg: globalCfg,
		logger:    logger,
		walDir:    walDir,
		hostname:  hostname,
		vc:        vc,
		exited:    make(chan bool),
	}

	ctx, cancel := context.WithCancel(context.Background())
	i.cancelScrape = cancel

	level.Debug(i.logger).Log("msg", "creating instance", "hostname", hostname)
	go i.run(ctx, reg, wstore)
	return i, nil
}

func (i *Instance) run(ctx context.Context, reg prometheus.Registerer, wstore walStorage) {
	discoveryManagerScrape := discovery.NewManager(ctx, log.With(i.logger, "component", "discovery manager scrape"), discovery.Name("scrape"))
	{
		// TODO(rfratto): refactor this to a function?
		// TODO(rfratto): ensure job name name is unique
		c := map[string]sd_config.ServiceDiscoveryConfig{}
		for _, v := range i.cfg.ScrapeConfigs {
			c[v.JobName] = v.ServiceDiscoveryConfig
		}
		i.exitErr = discoveryManagerScrape.ApplyConfig(c)
		if i.exitErr != nil {
			level.Error(i.logger).Log("msg", "failed applying config to discovery manager", "err", i.exitErr)
			return
		}
	}

	remoteStore := remote.NewStorage(log.With(i.logger, "component", "remote"), reg, wstore.StartTime, i.walDir, i.cfg.RemoteFlushDeadline)
	i.exitErr = remoteStore.ApplyConfig(&config.Config{
		GlobalConfig:       i.globalCfg,
		RemoteWriteConfigs: i.cfg.RemoteWrite,
	})
	if i.exitErr != nil {
		level.Error(i.logger).Log("msg", "failed applying config to remote storage", "err", i.exitErr)
		return
	}

	fanoutStorage := storage.NewFanout(i.logger, wstore, remoteStore)

	scrapeManager := scrape.NewManager(log.With(i.logger, "component", "scrape manager"), fanoutStorage)
	i.exitErr = scrapeManager.ApplyConfig(&config.Config{
		GlobalConfig:  i.globalCfg,
		ScrapeConfigs: i.cfg.ScrapeConfigs,
	})
	if i.exitErr != nil {
		level.Error(i.logger).Log("msg", "failed applying config to scrape manager", "err", i.exitErr)
		return
	}

	level.Debug(i.logger).Log("msg", "creating host filterer", "for_host", i.hostname)
	filterer := NewHostFilter(i.hostname)

	var g run.Group

	// The actors defined here are defined in the order we want them to shut down.
	// Primarily, we want to ensure that the following shutdown order is
	// maintained:
	//		1. The scrape manager stops
	//    2. WAL storage is closed
	//    3. Remote write storage is closed
	// This is done to allow the instance to write stale markers for all active
	// series.
	{
		// Scrape discovery manager
		g.Add(
			func() error {
				err := discoveryManagerScrape.Run()
				level.Info(i.logger).Log("msg", "service discovery manager stopped")
				return err
			},
			func(err error) {
				level.Info(i.logger).Log("msg", "stopping scrape discovery manager...")
				i.cancelScrape()
			},
		)
	}
	if i.cfg.HostFilter {
		// Target host filterer
		g.Add(
			func() error {
				filterer.Run(discoveryManagerScrape.SyncCh())
				level.Info(i.logger).Log("msg", "host filterer stopped")
				return nil
			},
			func(err error) {
				level.Info(i.logger).Log("msg", "stopping host filterer...")
				filterer.Stop()
			},
		)
	}
	{
		// Truncation loop
		ctx, contextCancel := context.WithCancel(context.Background())
		defer contextCancel()
		g.Add(
			func() error {
				i.truncateLoop(ctx, wstore)
				level.Info(i.logger).Log("msg", "truncation loop stopped")
				return nil
			},
			func(err error) {
				level.Info(i.logger).Log("msg", "stopping truncation loop...")
				contextCancel()
			},
		)
	}
	{
		// Scrape manager
		g.Add(
			func() error {
				var readCh GroupChannel
				if i.cfg.HostFilter {
					readCh = filterer.SyncCh()
				} else {
					readCh = discoveryManagerScrape.SyncCh()
				}

				err := scrapeManager.Run(readCh)
				level.Info(i.logger).Log("msg", "scrape manager stopped")
				return err
			},
			func(err error) {
				// The scrape manager is closed first to allow us to write staleness
				// markers without receiving new samples from scraping in the meantime.
				level.Info(i.logger).Log("msg", "stopping scrape manager...")
				scrapeManager.Stop()

				// On a graceful shutdown, write staleness markers. If something went
				// wrong, then the instance will be relaunched.
				if err == nil && i.cfg.WriteStaleOnShutdown {
					level.Info(i.logger).Log("msg", "writing staleness markers...")
					err := wstore.WriteStalenessMarkers(i.getRemoteWriteTimestamp)
					if err != nil {
						level.Error(i.logger).Log("msg", "error writing staleness markers", "err", err)
					}
				}

				level.Info(i.logger).Log("msg", "closing storage...")
				if err := fanoutStorage.Close(); err != nil {
					level.Error(i.logger).Log("msg", "error stopping storage", "err", err)
				}
			},
		)
	}

	err := g.Run()
	if err != nil {
		level.Error(i.logger).Log("msg", "agent instance stopped with error", "err", err)
	}
	if i.exitErr == nil {
		i.exitErr = err
	}

	close(i.exited)
}

func (i *Instance) truncateLoop(ctx context.Context, wal walStorage) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(i.cfg.WALTruncateFrequency):
			ts := i.getRemoteWriteTimestamp()
			if ts == 0 {
				level.Debug(i.logger).Log("msg", "can't truncate the WAL yet")
				continue
			}

			level.Debug(i.logger).Log("msg", "truncating the WAL", "ts", ts)
			err := wal.Truncate(ts)
			if err != nil {
				// The only issue here is larger disk usage and a greater replay time,
				// so we'll only log this as a warning.
				level.Warn(i.logger).Log("msg", "could not truncate WAL", "err", err)
			}
		}
	}
}

// getRemoteWriteTimestamp looks up the last successful remote write timestamp.
// This is passed to wal.Storage for its truncation.
func (i *Instance) getRemoteWriteTimestamp() int64 {
	lbls := make([]string, len(i.cfg.RemoteWrite))
	for idx := 0; idx < len(lbls); idx++ {
		lbls[idx] = i.cfg.RemoteWrite[idx].Name
	}

	vals, err := i.vc.GetValues("remote_name", lbls...)
	if err != nil {
		level.Error(i.logger).Log("msg", "could not get remote write timestamps", "err", err)
		return 0
	}
	if len(vals) == 0 {
		return 0
	}

	// We use the lowest value from the metric since we don't want to delete any
	// segments from the WAL until they've been written by all of the remote_write
	// configurations.
	ts := int64(math.MaxInt64)
	for _, val := range vals {
		ival := int64(val)
		if ival < ts {
			ts = ival
		}
	}

	// Convert to the millisecond precision which is used by the WAL
	return ts * 1000
}

// Config returns the instance's config.
func (i *Instance) Config() InstanceConfig {
	return i.cfg
}

// Wait blocks until the instance exits, returning its error, if any.
func (i *Instance) Wait() error {
	<-i.exited
	return i.exitErr
}

// Stop stops the instance.
func (i *Instance) Stop() {
	i.exitErr = errInstanceStoppedNormally

	i.cancelScrape()
	<-i.exited
}

// Err returns the error generated by instance when shut down. If the shutdown
// was intentional (i.e., the user called stop), then Err returns
// errInstanceStoppedNormally.
func (i *Instance) Err() error {
	return i.exitErr
}

// walStorage is an interface satisfied by wal.Storage, and created for testing.
type walStorage interface {
	// walStorage implements Queryable for compatibility, but is unused.
	storage.Queryable

	StartTime() (int64, error)
	WriteStalenessMarkers(remoteTsFunc func() int64) error
	Appender() (storage.Appender, error)
	Truncate(mint int64) error

	Close() error
}

func hostname() (string, error) {
	hostname := os.Getenv("HOSTNAME")
	if hostname != "" {
		return hostname, nil
	}

	return os.Hostname()
}

func getHash(data interface{}) (string, error) {
	bytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	hash := md5.Sum(bytes)
	return hex.EncodeToString(hash[:]), nil
}
