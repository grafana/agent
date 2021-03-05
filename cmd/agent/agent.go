package main

import (
	"fmt"
	"net/http"

	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/loki"
	"github.com/grafana/agent/pkg/tempo"

	"github.com/grafana/agent/pkg/config"
	"github.com/grafana/agent/pkg/prom"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/weaveworks/common/server"

	"github.com/go-kit/kit/log"
)

type EntryPoint struct {
	promMetrics *prom.Agent
	lokiLogs    *loki.Loki
	tempoTraces *tempo.Tempo
	manager     *integrations.Manager
	srv         *server.Server
}

func NewEntryPoint(logger log.Logger, cfg *config.Config) (*EntryPoint, error) {
	var (
		promMetrics *prom.Agent
		lokiLogs    *loki.Loki
		tempoTraces *tempo.Tempo
		manager     *integrations.Manager
	)

	srv, err := server.New(cfg.Server)
	if err != nil {
		return nil, err
	}

	if cfg.Prometheus.Enabled {
		promMetrics, err = prom.New(prometheus.DefaultRegisterer, cfg.Prometheus, logger)
		if err != nil {
			return nil, err
		}

		// Hook up API paths to the router
		promMetrics.WireAPI(srv.HTTP)
		promMetrics.WireGRPC(srv.GRPC)
	}

	if cfg.Loki.Enabled {
		lokiLogs, err = loki.New(prometheus.DefaultRegisterer, cfg.Loki, logger)
		if err != nil {
			return nil, err
		}
	}

	if cfg.Tempo.Enabled {
		tempoTraces, err = tempo.New(prometheus.DefaultRegisterer, cfg.Tempo, cfg.Server.LogLevel)
		if err != nil {
			return nil, err
		}
	}

	if cfg.Integrations.Enabled {
		manager, err = integrations.NewManager(cfg.Integrations, logger, promMetrics.InstanceManager())
		if err != nil {
			return nil, err

		}

		if err := manager.WireAPI(srv.HTTP); err != nil {
			return nil, err

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

	return &EntryPoint{
		promMetrics: promMetrics,
		lokiLogs:    lokiLogs,
		tempoTraces: tempoTraces,
		manager:     manager,
		srv:         srv,
	}, nil

}

func (srv *EntryPoint) Stop() {
	// Stop enabled subsystems
	if srv.manager != nil {
		srv.manager.Stop()
	}
	if srv.lokiLogs != nil {
		srv.lokiLogs.Stop()
	}
	if srv.promMetrics != nil {
		srv.promMetrics.Stop()
	}
	if srv.tempoTraces != nil {
		srv.tempoTraces.Stop()
	}
}

func (srv *EntryPoint) Start() error {
	return srv.srv.Run()
}
