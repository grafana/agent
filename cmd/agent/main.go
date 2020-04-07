package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	// Adds version information
	"github.com/gorilla/mux"
	_ "github.com/grafana/agent/cmd/agent/build"

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
	c.Server.RegisterFlags(f)
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

	if err := fs.Parse(os.Args[1:]); err != nil {
		log.Fatalf("error parsing flags: %v\n", err)
	}

	if printVersion {
		fmt.Println(version.Print("agent"))
		return
	}

	if configFile == "" {
		log.Fatalln("-config.file flag required")
	} else if err := loadConfig(configFile, &cfg); err != nil {
		log.Fatalf("error loading config file %s: %v\n", configFile, err)
	}

	// Parse the flags again to override any yaml values with command line flags
	if err := fs.Parse(os.Args[1:]); err != nil {
		log.Fatalf("error parsing flags: %v\n", err)
	}

	// After this point we can use util.Logger and stop using the log package
	util.InitLogger(&cfg.Server)

	promMetrics, err := prom.New(cfg.Prometheus, util.Logger)
	if err != nil {
		level.Error(util.Logger).Log("msg", "failed to create prometheus instance", "err", err)
		os.Exit(1)
	}

	srv, err := server.New(cfg.Server)
	if err != nil {
		level.Error(util.Logger).Log("msg", "failed to create server", "err", err)
		os.Exit(1)
	}

	// Hook up API paths to the router
	wireAgentConfigsAPI(srv.HTTP, promMetrics)

	if err := srv.Run(); err != nil {
		level.Error(util.Logger).Log("msg", "error running agent", "err", err)
		// Don't os.Exit here; we want to do cleanup by stopping promMetrics
	}

	promMetrics.Stop()
	level.Info(util.Logger).Log("msg", "agent exiting")
}

func wireAgentConfigsAPI(r *mux.Router, a *prom.Agent) {
	listConfig := a.WrapHandler(a.ListConfigurations)
	getConfig := a.WrapHandler(a.GetConfiguration)
	putConfig := a.WrapHandler(a.PutConfiguration)
	deleteConfig := a.WrapHandler(a.DeleteConfiguration)

	r.HandleFunc("/agent/api/v1/configs", listConfig).Methods("GET")
	r.HandleFunc("/agent/api/v1/configs/{name}", getConfig).Methods("GET")
	r.HandleFunc("/agent/api/v1/config/{name}", putConfig).Methods("PUT", "POST")
	r.HandleFunc("/agent/api/v1/config/{name}", deleteConfig).Methods("DELETE")
}

func loadConfig(filename string, config *Config) error {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return errors.Wrap(err, "Error reading config file")
	}

	return yaml.Unmarshal(buf, config)
}
