package prometheus

import (
	"context"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/grafana/agent/pkg/wal"
	"github.com/oklog/run"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/discovery"
	sd_config "github.com/prometheus/prometheus/discovery/config"
	"github.com/prometheus/prometheus/scrape"
	"github.com/prometheus/prometheus/storage"
	"github.com/prometheus/prometheus/storage/remote"
)

// InstanceConfig is a specific agent that runs within the overall Prometheus
// agent. It has its own set of scrape_configs, remote_write rules, and
// its own WAL directory.
type InstanceConfig struct {
	Name          string                      `yaml:"name"`
	WALDir        string                      `yaml:"wal_directory"`
	ScrapeConfigs []*config.ScrapeConfig      `yaml:"scrape_configs,omitempty"`
	RemoteWrite   []*config.RemoteWriteConfig `yaml:"remote_write,omitempty"`
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
	}
}

// Validate checks if the InstanceConfig has all required fields filled out.
// This should only be called after ApplyDefaults.
func (c *InstanceConfig) Validate() error {
	// TODO(rfratto): validation
	return nil
}

// instance is an individual metrics collector and remote_writer.
type instance struct {
	cfg       InstanceConfig
	globalCfg config.GlobalConfig
	logger    log.Logger

	cancelScrape context.CancelFunc
}

// newInstance creates and starts a new instance.
func newInstance(globalCfg config.GlobalConfig, cfg InstanceConfig, logger log.Logger) *instance {
	i := &instance{
		cfg:       cfg,
		globalCfg: globalCfg,
		logger:    log.With(logger, "instance", cfg.Name),
	}
	go i.run()
	return i
}

func (i *instance) run() {
	ctxScrape, cancelScrape := context.WithCancel(context.Background())
	i.cancelScrape = cancelScrape

	discoveryManagerScrape := discovery.NewManager(ctxScrape, log.With(i.logger, "component", "discovery manager scrape"), discovery.Name("scrape"))
	{
		// TODO(rfratto): refactor this to a function?
		c := map[string]sd_config.ServiceDiscoveryConfig{}
		for _, v := range i.cfg.ScrapeConfigs {
			c[v.JobName] = v.ServiceDiscoveryConfig
		}
		discoveryManagerScrape.ApplyConfig(c)
	}

	wstore, err := wal.NewStorage(i.logger, prometheus.DefaultRegisterer, i.cfg.WALDir)
	if err != nil {
		// TODO(rfratto): what do we want to do here? maybe anything that fails
		// should be brought out to newInstance
		panic(err)
	}

	// TODO(rfratto): tunable flush deadline?
	remoteStore := remote.NewStorage(log.With(i.logger, "component", "remote"), prometheus.DefaultRegisterer, wstore.StartTime, i.cfg.WALDir, time.Duration(1*time.Minute))
	remoteStore.ApplyConfig(&config.Config{
		GlobalConfig:       i.globalCfg,
		RemoteWriteConfigs: i.cfg.RemoteWrite,
	})

	fanoutStorage := storage.NewFanout(i.logger, wstore, remoteStore)

	scrapeManager := scrape.NewManager(log.With(i.logger, "component", "scrape manager"), fanoutStorage)
	scrapeManager.ApplyConfig(&config.Config{
		GlobalConfig:  i.globalCfg,
		ScrapeConfigs: i.cfg.ScrapeConfigs,
	})

	var g run.Group
	// Prometheus generally runs a Termination handler here, but termination handling
	// is done outside of the instance.
	// TODO: anything else we need to do here?
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
		// Scrape manager
		g.Add(
			func() error {
				err := scrapeManager.Run(discoveryManagerScrape.SyncCh())
				level.Info(i.logger).Log("msg", "scrape manager stopped")
				return err
			},
			func(err error) {
				// TODO(rfratto): is this correct? do we want to stop the fanoutStorage
				// later?
				if err := fanoutStorage.Close(); err != nil {
					level.Error(i.logger).Log("msg", "error stopping storage", "err", err)
				}
				level.Info(i.logger).Log("msg", "stopping scrape manager...")
				scrapeManager.Stop()
			},
		)
	}

	err = g.Run()
	if err != nil {
		level.Error(i.logger).Log("msg", "agent instance stopped with error", "err", err)
	}

	// TODO(rfratto): what if this function exits enexpectedly? how should that be
	// handled? should it be restarted?
}

// Stop stops the instance.
func (i *instance) Stop() {
	// TODO(rfratto): anything else we need to stop here?
	i.cancelScrape()
}
