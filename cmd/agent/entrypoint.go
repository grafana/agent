package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

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
		reloadMux.HandleFunc("/-/reload", ep.reloadHandler).Methods("GET", "POST")
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

	ep.tempoTraces, err = tempo.New(ep.lokiLogs, ep.promMetrics.InstanceManager(), prometheus.DefaultRegisterer, cfg.Tempo, cfg.Server.LogLevel.Logrus)
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
func (ep *Entrypoint) ApplyConfig(cfg config.Config) error {
	ep.mut.Lock()
	defer ep.mut.Unlock()

	var failed bool

	if err := ep.log.ApplyConfig(&cfg.Server); err != nil {
		level.Error(ep.log).Log("msg", "failed to update logger", "err", err)
		failed = true
	}

	if err := ep.srv.ApplyConfig(cfg.Server, ep.wire); err != nil {
		level.Error(ep.log).Log("msg", "failed to update server", "err", err)
		failed = true
	}

	// Go through each component and update it.
	if err := ep.promMetrics.ApplyConfig(cfg.Prometheus); err != nil {
		level.Error(ep.log).Log("msg", "failed to update prometheus", "err", err)
		failed = true
	}

	if err := ep.lokiLogs.ApplyConfig(cfg.Loki); err != nil {
		level.Error(ep.log).Log("msg", "failed to update loki", "err", err)
		failed = true
	}

	if err := ep.tempoTraces.ApplyConfig(ep.lokiLogs, ep.promMetrics.InstanceManager(), cfg.Tempo, cfg.Server.LogLevel.Logrus); err != nil {
		level.Error(ep.log).Log("msg", "failed to update tempo", "err", err)
		failed = true
	}

	if err := ep.manager.ApplyConfig(cfg.Integrations); err != nil {
		level.Error(ep.log).Log("msg", "failed to update integrations", "err", err)
		failed = true
	}

	ep.cfg = cfg
	if failed {
		return fmt.Errorf("changes did not apply successfully")
	}

	return nil
}

// wire is used to hook up API endpoints to components, and is called every
// time a new Weaveworks server is creatd.
func (ep *Entrypoint) wire(mux *mux.Router, grpc *grpc.Server) {
	ep.promMetrics.WireAPI(mux)
	ep.promMetrics.WireGRPC(grpc)

	ep.manager.WireAPI(mux)

	mux.HandleFunc("/-/healthy", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Agent is Healthy.\n")
	})

	mux.HandleFunc("/-/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Agent is Ready.\n")
	})

	mux.HandleFunc("/-/config", func(rw http.ResponseWriter, r *http.Request) {
		ep.mut.Lock()
		bb, err := yaml.Marshal(ep.cfg)
		ep.mut.Unlock()

		if err != nil {
			http.Error(rw, fmt.Sprintf("failed to marshal config: %s", err), http.StatusInternalServerError)
		} else {
			_, _ = rw.Write(bb)
		}
	})

	mux.HandleFunc("/-/reload", ep.reloadHandler).Methods("GET", "POST")
}

func (ep *Entrypoint) reloadHandler(rw http.ResponseWriter, r *http.Request) {
	success := ep.TriggerReload()
	if success {
		rw.WriteHeader(http.StatusOK)
	} else {
		rw.WriteHeader(http.StatusBadRequest)
	}
}

// TriggerReload will cause the Entrypoint to re-request the config file and
// apply the latest config. TriggerReload returns true if the reload was
// successful.
func (ep *Entrypoint) TriggerReload() bool {
	level.Info(ep.log).Log("msg", "reload of config file requested")

	cfg, err := ep.reloader()
	if err != nil {
		level.Error(ep.log).Log("msg", "failed to reload config file", "err", err)
		return false
	}

	err = ep.ApplyConfig(*cfg)
	if err != nil {
		level.Error(ep.log).Log("msg", "failed to reload config file", "err", err)
		return false
	}
	return true
}

// Stop stops the Entrypoint and all subsystems.
func (ep *Entrypoint) Stop() {
	ep.mut.Lock()
	defer ep.mut.Unlock()

	ep.manager.Stop()
	ep.lokiLogs.Stop()
	ep.promMetrics.Stop()
	ep.tempoTraces.Stop()
	ep.srv.Close()

	if ep.reloadServer != nil {
		ep.reloadServer.Close()
	}
}

// Start starts the server used by the Entrypoint, and will block until a
// termination signal is sent to the process.
func (ep *Entrypoint) Start() error {
	var g run.Group

	// Create a signal handler that will stop the Entrypoint once a termination
	// signal is received.
	signalHandler := signals.NewHandler(ep.cfg.Server.Log)

	notifier := make(chan os.Signal)
	signal.Notify(notifier, syscall.SIGHUP)

	defer func() {
		signal.Stop(notifier)
		close(notifier)
	}()

	g.Add(func() error {
		signalHandler.Loop()
		return nil
	}, func(e error) {
		signalHandler.Stop()
	})

	if ep.reloadServer != nil && ep.reloadListener != nil {
		g.Add(func() error {
			return ep.reloadServer.Serve(ep.reloadListener)
		}, func(e error) {
			ep.reloadServer.Close()
		})
	}

	g.Add(func() error {
		return ep.srv.Run()
	}, func(e error) {
		ep.srv.Close()
	})

	go func() {
		for range notifier {
			ep.TriggerReload()
		}
	}()

	return g.Run()
}
