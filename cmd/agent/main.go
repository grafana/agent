package main

import (
	"flag"
	"log"
	"os"

	util_log "github.com/cortexproject/cortex/pkg/util/log"
	"github.com/go-kit/kit/log/level"
	"github.com/grafana/agent/pkg/config"

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
	// If this is a windows service then run it until if finishes
	if IsWindowService() {
		RunService()
		return
	}

	// This is not a windows service so proceed normally
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	cfg, err := config.Load(fs, os.Args[1:])
	if err != nil {
		log.Fatalln(err)
	}

	// After this point we can use util_log.Logger and stop using the log package
	util_log.InitLogger(&cfg.Server)
	logger := util_log.Logger
	srv := NewAgentServer(logger, cfg)
	if err := srv.srv.Run(); err != nil {
		level.Error(logger).Log("msg", "error running agent", "err", err)
		// Don't os.Exit here; we want to do cleanup by stopping promMetrics
	}

	srv.Stop()
	level.Info(logger).Log("msg", "agent exiting")
}
