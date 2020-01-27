package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/cortexproject/cortex/pkg/util"
	"github.com/go-kit/kit/log/level"
	prom "github.com/grafana/agent/pkg/prometheus"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/version"
	"github.com/weaveworks/common/server"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Server     server.Config `yaml:"server"`
	Prometheus prom.Config   `yaml:"prometheus,omitempty"`
}

func (c *Config) RegisterFlags(f *flag.FlagSet) {
	c.Server.MetricsNamespace = "agent"
	c.Server.RegisterInstrumentation = true
	c.Prometheus.RegisterFlags(f)
}

// ApplyDefaults applies default values to the config for fields that
// have not changed to their non-zero value.
func (c *Config) ApplyDefaults() {
	c.Prometheus.ApplyDefaults()
}

func init() {
	prometheus.MustRegister(version.NewCollector("agent"))
}

func main() {
	var (
		printVersion bool

		cfg        Config
		configFile string
	)

	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fs.StringVar(&configFile, "config.file", "", "configuration file to load")
	fs.BoolVar(&printVersion, "version", false, "Print this build's version information")
	cfg.RegisterFlags(fs)
	fs.Parse(os.Args[1:])

	if printVersion {
		fmt.Println(version.Print("agent"))
		return
	}

	if configFile == "" {
		exitWithError("-config.file flag required")
	} else if err := loadConfig(configFile, &cfg); err != nil {
		exitWithError("error loading config file %s: %v", configFile, err)
	}

	// Parse the flags again to override any yaml stuff with command line flags
	fs.Parse(os.Args[1:])

	util.InitLogger(&cfg.Server)

	cfg.ApplyDefaults()

	promMetrics, err := prom.New(cfg.Prometheus, util.Logger)
	if err != nil {
		level.Error(util.Logger).Log("msg", "failed to create prometheus instance", "err", err)
	}

	srv, err := server.New(cfg.Server)
	if err != nil {
		level.Error(util.Logger).Log("msg", "failed to create server", "err", err)
	}

	if err := srv.Run(); err != nil {
		level.Error(util.Logger).Log("msg", "error running agent", "err", err)
	}

	promMetrics.Stop()
	level.Info(util.Logger).Log("msg", "agent exiting")
}

func exitWithError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}

func loadConfig(filename string, config *Config) error {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return errors.Wrap(err, "Error reading config file")
	}

	return yaml.Unmarshal(buf, config)
}
