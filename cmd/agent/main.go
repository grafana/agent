package main

import (
	"flag"
	"log"
	"os"

	util_log "github.com/cortexproject/cortex/pkg/util/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/pkg/config"
	"github.com/grafana/agent/pkg/util"
	"github.com/weaveworks/common/logging"

	// Adds version information
	_ "github.com/grafana/agent/pkg/build"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/version"

	// Register Prometheus SD components
	_ "github.com/prometheus/prometheus/discovery/install"

	// Register integrations
	_ "github.com/grafana/agent/pkg/integrations/install"
)

func init() {
	prometheus.MustRegister(version.NewCollector("agent"))
}

func main() {
	// If Windows is trying to run us as a service, go through that
	// path instead.
	if IsWindowsService() {
		err := RunService()
		if err != nil {
			log.Fatalln(err)
		}
		return
	}

	var cfgLogger logging.Interface

	reloader := func() (*config.Config, error) {
		fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
		cfg, err := config.Load(fs, os.Args[1:])
		if cfg != nil {
			cfg.Server.Log = cfgLogger
		}
		return cfg, err
	}
	cfg, err := reloader()
	if err != nil {
		log.Fatalln(err)
	}

	// After this point we can start using go-kit logging.
	logger := util.NewLogger(&cfg.Server)
	util_log.Logger = logger

	// We need to manually set the logger for the first call to reload.
	// Subsequent reloads will use cfgLogger.
	cfgLogger = util.GoKitLogger(logger)
	cfg.Server.Log = cfgLogger

	ep, err := NewEntrypoint(logger, cfg, reloader)
	if err != nil {
		level.Error(logger).Log("msg", "error creating the agent server entrypoint", "err", err)
		os.Exit(1)
	}

	if err = ep.Start(); err != nil {
		level.Error(logger).Log("msg", "error running agent", "err", err)
		// Don't os.Exit here; we want to do cleanup by stopping promMetrics
	}

	ep.Stop()
	level.Info(logger).Log("msg", "agent exiting")
}
