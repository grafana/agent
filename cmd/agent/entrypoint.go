package main

import (
	"fmt"
	"net"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/loki"
	"github.com/grafana/agent/pkg/tempo"
	"github.com/grafana/agent/pkg/util"
	"github.com/grafana/agent/pkg/util/server"
	"github.com/oklog/run"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v2"

	"github.com/grafana/agent/pkg/config"
	"github.com/grafana/agent/pkg/prom"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/weaveworks/common/signals"

	"github.com/go-kit/kit/log/level"
)

// Entrypoint is the entrypoint of the application that starts all subsystems.
type Entrypoint struct {
	mut sync.Mutex

	reloader Reloader

	log *util.Logger
	cfg config.Config

	srv         *server.Server
	promMetrics *prom.Agent
	lokiLogs    *loki.Loki
	tempoTraces *tempo.Tempo
	manager     *integrations.Manager

	reloadListener net.Listener
	reloadServer   *http.Server
}

// Reloader is any function that returns a new config.
type Reloader = func() (*config.Config, error)

// NewEntrypoint creates a new Entrypoint.
func NewEntrypoint(logger *util.Logger, cfg *config.Config, reloader Reloader) (*Entrypoint, error) {
	var (
		ep = &Entrypoint{
			log:      logger,
			reloader: reloader,
		}
		err error
	)

	if cfg.ReloadPort != 0 {
		ep.reloadListener, err = net.Listen("tcp", fmt.Sprintf("%s:%d", cfg.ReloadAddress, cfg.ReloadPort))
		if err != nil {
			return nil, fmt.Errorf("failed to listen on address for secondary /-/reload server: %w", err)
		}

		reloadMux := mux.NewRouter()
		reloadMux.HandleFunc("/-/reload", ep.reloadHandler)
		ep.reloadServer = &http.Server{Handler: reloadMux}
	}

	ep.srv = server.New(prometheus.DefaultRegisterer, logger)

	ep.promMetrics, err = prom.New(prometheus.DefaultRegisterer, cfg.Prometheus, logger)
	if err != nil {
		return nil, err
	}

	ep.lokiLogs, err = loki.New(prometheus.DefaultRegisterer, cfg.Loki, logger)
	if err != nil {
		return nil, err
	}

	ep.tempoTraces, err = tempo.New(prometheus.DefaultRegisterer, cfg.Tempo, cfg.Server.LogLevel.Logrus)
	if err != nil {
		return nil, err
	}

	ep.manager, err = integrations.NewManager(cfg.Integrations, logger, ep.promMetrics.InstanceManager(), ep.promMetrics.Validate)
	if err != nil {
		return nil, err
	}

	// Mostly everything should be up to date except for the server, which hasn't
	// been created yet.
	if err := ep.ApplyConfig(*cfg); err != nil {
		return nil, err
	}
	return ep, nil
}

// ApplyConfig applies changes to the subsystems of the Agent.
func (srv *Entrypoint) ApplyConfig(cfg config.Config) error {
	srv.mut.Lock()
	defer srv.mut.Unlock()

	var failed bool

	if err := srv.log.ApplyConfig(&cfg.Server); err != nil {
		level.Error(srv.log).Log("msg", "failed to update logger", "err", err)
		failed = true
	}

	if err := srv.srv.ApplyConfig(cfg.Server, srv.wire); err != nil {
		level.Error(srv.log).Log("msg", "failed to update server", "err", err)
		failed = true
	}

	// Go through each component and update it.
	if err := srv.promMetrics.ApplyConfig(cfg.Prometheus); err != nil {
		level.Error(srv.log).Log("msg", "failed to update prometheus", "err", err)
		failed = true
	}

	if err := srv.lokiLogs.ApplyConfig(cfg.Loki); err != nil {
		level.Error(srv.log).Log("msg", "failed to update loki", "err", err)
		failed = true
	}

	if err := srv.tempoTraces.ApplyConfig(cfg.Tempo, cfg.Server.LogLevel.Logrus); err != nil {
		level.Error(srv.log).Log("msg", "failed to update tempo", "err", err)
		failed = true
	}

	if err := srv.manager.ApplyConfig(cfg.Integrations); err != nil {
		level.Error(srv.log).Log("msg", "failed to update integrations", "err", err)
		failed = true
	}

	srv.cfg = cfg
	if failed {
		return fmt.Errorf("changes did not apply successfully")
	}

	return nil
}

// wire is used to hook up API endpoints to components, and is called every
// time a new Weaveworks server is creatd.
func (srv *Entrypoint) wire(mux *mux.Router, grpc *grpc.Server) {
	srv.promMetrics.WireAPI(mux)
	srv.promMetrics.WireGRPC(grpc)

	srv.manager.WireAPI(mux)

	mux.HandleFunc("/-/healthy", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Agent is Healthy.\n")
	})

	mux.HandleFunc("/-/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Agent is Ready.\n")
	})

	mux.HandleFunc("/agent/api/v1/config", func(rw http.ResponseWriter, r *http.Request) {
		srv.mut.Lock()
		bb, err := yaml.Marshal(srv.cfg)
		srv.mut.Unlock()

		if err != nil {
			http.Error(rw, fmt.Sprintf("failed to marshal config: %s", err), http.StatusInternalServerError)
		} else {
			_, _ = rw.Write(bb)
		}
	})

	mux.HandleFunc("/-/reload", srv.reloadHandler)
}

func (srv *Entrypoint) reloadHandler(rw http.ResponseWriter, r *http.Request) {
	success := srv.TriggerReload()
	if success {
		rw.WriteHeader(http.StatusOK)
	} else {
		rw.WriteHeader(http.StatusBadRequest)
	}
}

// TriggerReload will cause the Entrypoint to re-request the config file and
// apply the latest config. TriggerReload returns true if the reload was
// successful.
func (srv *Entrypoint) TriggerReload() bool {
	level.Info(srv.log).Log("msg", "reload of config file requested")

	cfg, err := srv.reloader()
	if err != nil {
		level.Error(srv.log).Log("msg", "failed to reload config file", "err", err)
		return false
	}

	err = srv.ApplyConfig(*cfg)
	if err != nil {
		level.Error(srv.log).Log("msg", "failed to reload config file", "err", err)
		return false
	}
	return true
}

// Stop stops the Entrypoint and all subsystems.
func (srv *Entrypoint) Stop() {
	srv.mut.Lock()
	defer srv.mut.Unlock()

	srv.manager.Stop()
	srv.lokiLogs.Stop()
	srv.promMetrics.Stop()
	srv.tempoTraces.Stop()
	srv.srv.Close()

	if srv.reloadServer != nil {
		srv.reloadServer.Close()
	}
}

// Start starts the server used by the Entrypoint, and will block until a
// termination signal is sent to the process.
func (srv *Entrypoint) Start() error {
	var g run.Group

	// Create a signal handler that will stop the Entrypoint once a termination
	// signal is received.
	signalHandler := signals.NewHandler(srv.cfg.Server.Log)

	g.Add(func() error {
		signalHandler.Loop()
		return nil
	}, func(e error) {
		signalHandler.Stop()
	})

	if srv.reloadServer != nil && srv.reloadListener != nil {
		g.Add(func() error {
			return srv.reloadServer.Serve(srv.reloadListener)
		}, func(e error) {
			srv.reloadServer.Close()
		})
	}

	g.Add(func() error {
		return srv.srv.Run()
	}, func(e error) {
		srv.srv.Close()
	})

	return g.Run()
}
