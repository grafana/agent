package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cortexproject/cortex/pkg/util"
	"github.com/go-kit/kit/log/level"
	prom "github.com/grafana/agent/pkg/prometheus"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	"github.com/prometheus/prometheus/config"
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

	Prometheus prom.Config `yaml:"prometheus,omitempty"`
}

// ApplyDefaults applies default values to the config for fields that
// have not changed to their non-zero value.
func (c *Config) ApplyDefaults() {
	c.Prometheus.ApplyDefaults()
}

// Validate checks if the Config has all required fields filled out.
// Should be called after ApplyDefaults.
func (c *Config) Validate() error {
	return c.Prometheus.Validate()
}

func init() {
	prometheus.MustRegister(version.NewCollector("agent"))
}

// TODO(rfratto):
//   1. WAL as flag, create subdirectories for each agent instance
//   2. Get rid of the silly mainError function

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

	go exposeTestMetric()

	cfg.ApplyDefaults()
	if err := cfg.Validate(); err != nil {
		level.Error(util.Logger).Log("msg", "config validation failed", "err", err)
		return nil
	}

	promMetrics := prom.New(cfg.Prometheus, util.Logger)

	// TODO(rfratto): this is going to block forever, even if promMetrics
	// stops. We need a better solution for this.
	term := make(chan os.Signal, 1)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)
	<-term

	promMetrics.Stop()

	level.Info(util.Logger).Log("msg", "agent exiting")
	return nil
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
