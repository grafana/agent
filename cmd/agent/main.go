package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	// Adds version information
	_ "github.com/grafana/agent/pkg/build"

	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/loki"
	"github.com/grafana/agent/pkg/tempo"

	util_log "github.com/cortexproject/cortex/pkg/util/log"
	"github.com/go-kit/kit/log/level"
	"github.com/grafana/agent/pkg/config"
	"github.com/grafana/agent/pkg/prom"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/version"
	"github.com/weaveworks/common/server"

	// Register Prometheus SD components
	_ "github.com/prometheus/prometheus/discovery/install"

	// Register integrations
	_ "github.com/grafana/agent/pkg/integrations/install"
)

func init() {
	prometheus.MustRegister(version.NewCollector("agent"))
}

func main() {
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	cfg, err := config.Load(fs, os.Args[1:])
	if err != nil {
		log.Fatalln(err)
	}

	// After this point we can use util_log.Logger and stop using the log package
	util_log.InitLogger(&cfg.Server)
	logger := util_log.Logger

	var (
		promMetrics *prom.Agent
		lokiLogs    *loki.Loki
		tempoTraces *tempo.Tempo
		manager     *integrations.Manager
	)

	srv, err := server.New(cfg.Server)
	if err != nil {
		level.Error(logger).Log("msg", "failed to create server", "err", err)
		os.Exit(1)
	}

	if cfg.Prometheus.Enabled {
		promMetrics, err = prom.New(prometheus.DefaultRegisterer, cfg.Prometheus, logger)
		if err != nil {
			level.Error(logger).Log("msg", "failed to create prometheus instance", "err", err)
			os.Exit(1)
		}

		// Hook up API paths to the router
		promMetrics.WireAPI(srv.HTTP)
		promMetrics.WireGRPC(srv.GRPC)
	}

	if cfg.Loki.Enabled {
		lokiLogs, err = loki.New(prometheus.DefaultRegisterer, cfg.Loki, logger)
		if err != nil {
			level.Error(logger).Log("msg", "failed to create loki log collection instance", "err", err)
			os.Exit(1)
		}
	}

	if cfg.Tempo.Enabled {
		tempoTraces, err = tempo.New(prometheus.DefaultRegisterer, cfg.Tempo, cfg.Server.LogLevel)
		if err != nil {
			level.Error(logger).Log("msg", "failed to create tempo trace collection instance", "err", err)
			os.Exit(1)
		}
	}

	if cfg.Integrations.Enabled {
		cfg.Integrations.ClientCA = cfg.Server.HTTPTLSConfig.ClientCAs
		cfg.Integrations.ClientAuthType = cfg.Server.HTTPTLSConfig.ClientAuth
		manager, err = integrations.NewManager(cfg.Integrations, logger, promMetrics.InstanceManager())
		if err != nil {
			level.Error(logger).Log("msg", "failed to create integrations manager", "err", err)
			os.Exit(1)
		}

		if err := manager.WireAPI(srv.HTTP); err != nil {
			level.Error(logger).Log("msg", "failed wiring endpoints for integrations", "err", err)
			os.Exit(1)
		}
	}

	srv.HTTP.HandleFunc("/-/healthy", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Agent is Healthy.\n")
	})
	srv.HTTP.HandleFunc("/-/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Agent is Ready.\n")
	})

	if err := srv.Run(); err != nil {
		level.Error(logger).Log("msg", "error running agent", "err", err)
		// Don't os.Exit here; we want to do cleanup by stopping promMetrics
	}

	// Stop enabled subsystems
	if manager != nil {
		manager.Stop()
	}
	if lokiLogs != nil {
		lokiLogs.Stop()
	}
	if promMetrics != nil {
		promMetrics.Stop()
	}
	if tempoTraces != nil {
		tempoTraces.Stop()
	}
	level.Info(logger).Log("msg", "agent exiting")
}
