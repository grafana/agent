package prometheus

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math"
	"os"
	"path"
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
		WALTruncateFrequency: 1 * time.Minute,
		RemoteFlushDeadline:  1 * time.Minute,
	}
)

// InstanceConfig is a specific agent that runs within the overall Prometheus
// agent. It has its own set of scrape_configs and remote_write rules.
type InstanceConfig struct {
	Name          string                      `yaml:"name"`
	ScrapeConfigs []*config.ScrapeConfig      `yaml:"scrape_configs,omitempty"`
	RemoteWrite   []*config.RemoteWriteConfig `yaml:"remote_write,omitempty"`

	// How frequently the WAL should be truncated.
	WALTruncateFrequency time.Duration `yaml:"wal_truncate_frequency"`

	RemoteFlushDeadline time.Duration `yaml:"remote_flush_deadline,omitempty"`
}

func (c *InstanceConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultInstanceConfig

	type plain InstanceConfig
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}

	return nil
}

// ApplyDefaults applies default configurations to the configuration to all
// values that have not been changed to their non-zero value.
func (c *InstanceConfig) ApplyDefaults(global *config.GlobalConfig) {
	// TODO(rfratto): what other defaults need to be applied?
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

// instance is an individual metrics collector and remote_writer.
type instance struct {
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

// newInstance creates and starts a new instance.
func newInstance(globalCfg config.GlobalConfig, cfg InstanceConfig, walDir string, logger log.Logger) (*instance, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	// We need to be able to pull the metrics of the last written timestamp for all the remote_writes.
	// This can currently only be done using the remote_name label Prometheus adds to the
	// prometheus_remote_storage_queue_highest_sent_time_seconds label, so we force a name here
	// that we can lookup later.
	for _, rcfg := range cfg.RemoteWrite {
		if rcfg.Name != "" {
			continue
		}

		hash, err := getHash(rcfg)
		if err != nil {
			return nil, err
		}

		// We have to add the name of the instance to ensure that generated metrics
		// are unique across multiple agent instances. The remote write queues currently
		// globally register their metrics so we can't inject labels here.
		rcfg.Name = cfg.Name + "-" + hash[:6]
	}

	hostname, err := hostname()
	if err != nil {
		return nil, err
	}

	vc := NewMetricValueCollector(prometheus.DefaultGatherer, remoteWriteMetricName)

	i := &instance{
		cfg:       cfg,
		globalCfg: globalCfg,
		logger:    log.With(logger, "instance", cfg.Name),
		walDir:    path.Join(walDir, cfg.Name),
		hostname:  hostname,
		vc:        vc,
		exited:    make(chan bool),
	}

	level.Debug(i.logger).Log("msg", "creating instance", "hostname", hostname)

	wstore, err := wal.NewStorage(i.logger, prometheus.DefaultRegisterer, i.walDir)
	if err != nil {
		return nil, err
	}

	go i.run(wstore)
	return i, nil
}

// Err returns the error generated by instance when shut down. If the shutdown
// was intentional (i.e., the user called stop), then Err returns
// errInstanceStoppedNormally.
func (i *instance) Err() error {
	return i.exitErr
}

func (i *instance) run(wstore *wal.Storage) {
	ctxScrape, cancelScrape := context.WithCancel(context.Background())
	i.cancelScrape = cancelScrape

	discoveryManagerScrape := discovery.NewManager(ctxScrape, log.With(i.logger, "component", "discovery manager scrape"), discovery.Name("scrape"))
	{
		// TODO(rfratto): refactor this to a function?
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

	// TODO(rfratto): tunable flush deadline?
	remoteStore := remote.NewStorage(log.With(i.logger, "component", "remote"), prometheus.DefaultRegisterer, wstore.StartTime, i.walDir, i.cfg.RemoteFlushDeadline)
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
	{
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
				err := scrapeManager.Run(filterer.SyncCh())
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
				if err == nil {
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

func (i *instance) truncateLoop(ctx context.Context, wal *wal.Storage) {
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

func (i *instance) getRemoteWriteTimestamp() int64 {
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

// Stop stops the instance.
func (i *instance) Stop() {
	i.exitErr = errInstanceStoppedNormally

	// TODO(rfratto): anything else we need to stop here?
	i.cancelScrape()
	<-i.exited
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
