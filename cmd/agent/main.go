package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cortexproject/cortex/pkg/util"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/grafana/agent/pkg/wal"
	"github.com/oklog/run"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	"github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/discovery"
	sd_config "github.com/prometheus/prometheus/discovery/config"
	"github.com/prometheus/prometheus/scrape"
	"github.com/prometheus/prometheus/storage"
	"github.com/prometheus/prometheus/storage/remote"
	"github.com/weaveworks/common/logging"
	"github.com/weaveworks/common/server"
	"gopkg.in/yaml.v2"
)

func zeroGlobalConfig(c config.GlobalConfig) bool {
	return c.ExternalLabels == nil &&
		c.ScrapeInterval == 0 &&
		c.ScrapeTimeout == 0 &&
		c.EvaluationInterval == 0
}

type Config struct {
	// TODO(rfratto): move to weaveworks/common sever config
	LogLevel logging.Level `yaml:"log_level"`

	Prometheus struct {
		GlobalConfig config.GlobalConfig `yaml:"global"`
		Configs      []PrometheusConfig  `yaml:"configs,omitempty"`
	} `yaml:"prometheus,omitempty"`
}

type PrometheusConfig struct {
	Name          string                      `yaml:"name,omitempty"`
	ScrapeConfigs []*config.ScrapeConfig      `yaml:"scrape_configs,omitempty"`
	RemoteWrite   []*config.RemoteWriteConfig `yaml:"remote_write,omitempty"`
}

func init() {
	prometheus.MustRegister(version.NewCollector("agent"))
}

func main() {
	err := mainError()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func mainError() error {
	var (
		printVersion bool

		cfg        Config
		configFile string

		// TODO(rfratto): make configurable
		walDirectory string = ".wal"
	)

	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fs.StringVar(&configFile, "config.file", "", "configuration file to load")
	fs.BoolVar(&printVersion, "version", false, "Print this build's version information")
	cfg.LogLevel.RegisterFlags(fs)
	fs.Parse(os.Args[1:])

	if printVersion {
		fmt.Println(version.Print("agent"))
		return nil
	}

	if configFile == "" {
		return errors.New("-config.file flag required")
	} else if err := loadConfig(configFile, &cfg); err != nil {
		return fmt.Errorf("error loading config file %s: %v", configFile, err)
	}

	util.InitLogger(&server.Config{
		LogLevel: cfg.LogLevel,
	})
	level.Debug(util.Logger).Log("msg", "debug logging enabled")

	if zeroGlobalConfig(cfg.Prometheus.GlobalConfig) {
		cfg.Prometheus.GlobalConfig = config.DefaultGlobalConfig
	}

	go exposeTestMetric()

	ctxScrape, cancelScrape := context.WithCancel(context.Background())
	defer cancelScrape()

	discoveryManagerScrape := discovery.NewManager(ctxScrape, log.With(util.Logger, "component", "discovery manager scrape"), discovery.Name("scrape"))
	{
		c := map[string]sd_config.ServiceDiscoveryConfig{}
		// TODO(rfratto): remove hardcoded configs[0]
		for _, v := range cfg.Prometheus.Configs[0].ScrapeConfigs {
			// TODO(rfratto): move this somewhere else
			if v.ScrapeInterval == 0 {
				v.ScrapeInterval = cfg.Prometheus.GlobalConfig.ScrapeInterval
			}
			if v.ScrapeTimeout == 0 {
				if cfg.Prometheus.GlobalConfig.ScrapeTimeout > v.ScrapeInterval {
					v.ScrapeTimeout = v.ScrapeInterval
				} else {
					v.ScrapeTimeout = cfg.Prometheus.GlobalConfig.ScrapeTimeout
				}
			}
			c[v.JobName] = v.ServiceDiscoveryConfig
		}
		discoveryManagerScrape.ApplyConfig(c)
	}

	wstore, err := wal.NewStorage(util.Logger, prometheus.DefaultRegisterer, walDirectory)
	if err != nil {
		panic(err)
	}

	remoteStorage := remote.NewStorage(log.With(util.Logger, "component", "remote"), prometheus.DefaultRegisterer, wstore.StartTime, walDirectory, time.Duration(1*time.Minute))
	// TODO(rfratto): remove hardcoded configs[0]
	remoteStorage.ApplyConfig(&config.Config{
		GlobalConfig:       cfg.Prometheus.GlobalConfig,
		RemoteWriteConfigs: cfg.Prometheus.Configs[0].RemoteWrite,
	})

	fanoutStorage := storage.NewFanout(util.Logger, wstore, remoteStorage)

	scrapeManager := scrape.NewManager(log.With(util.Logger, "component", "scrape manager"), fanoutStorage)
	// TODO(rfratto): remove hardcoded configs[0]
	scrapeManager.ApplyConfig(&config.Config{
		GlobalConfig:  cfg.Prometheus.GlobalConfig,
		ScrapeConfigs: cfg.Prometheus.Configs[0].ScrapeConfigs,
	})

	var g run.Group
	{
		// Temination handler.
		term := make(chan os.Signal, 1)
		signal.Notify(term, os.Interrupt, syscall.SIGTERM)
		cancel := make(chan struct{})
		g.Add(
			func() error {
				select {
				case <-term:
				case <-cancel:
				}
				return nil
			},
			func(err error) {
				close(cancel)
			},
		)
	}
	{
		// Scrape discovery manger.
		g.Add(
			func() error {
				err := discoveryManagerScrape.Run()
				level.Info(util.Logger).Log("msg", "scrape discovery manager stopped")
				return err
			},
			func(err error) {
				level.Info(util.Logger).Log("msg", "stopping scrape discovery manager...")
				cancelScrape()
			},
		)
	}
	{
		// Scrape manager
		g.Add(
			func() error {
				err := scrapeManager.Run(discoveryManagerScrape.SyncCh())
				level.Info(util.Logger).Log("msg", "scrape manager stopped")
				return err
			},
			func(err error) {
				if err := fanoutStorage.Close(); err != nil {
					level.Error(util.Logger).Log("msg", "error stopping storage", "err", err)
				}
				level.Info(util.Logger).Log("msg", "stopping scrape manager...")
				scrapeManager.Stop()
			},
		)
	}

	return g.Run()
}

func exposeTestMetric() {
	testMetric := promauto.NewCounter(prometheus.CounterOpts{
		Name: "agent_test_metric_total",
	})

	go func() {
		for {
			testMetric.Inc()
			time.Sleep(5 * time.Second)
		}
	}()

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":12345", nil)
}

func loadConfig(filename string, config *Config) error {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return errors.Wrap(err, "Error reading config file")
	}

	return yaml.Unmarshal(buf, config)
}
