package main

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/loki"
	"github.com/grafana/agent/pkg/tempo"
	"github.com/grafana/agent/pkg/util"
	"github.com/oklog/run"
	"go.uber.org/atomic"
	"gopkg.in/yaml.v2"

	"github.com/grafana/agent/pkg/config"
	"github.com/grafana/agent/pkg/prom"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/weaveworks/common/server"
	"github.com/weaveworks/common/signals"

	"github.com/go-kit/kit/log/level"
)

// Entrypoint is the entrypoint of the application that starts all subsystems.
type Entrypoint struct {
	mut sync.Mutex

	reloader Reloader

	log *util.Logger
	cfg config.Config

	srv   *server.Server
	srvCh chan *server.Server

	doneCh    chan bool
	closeOnce sync.Once

	promMetrics *prom.Agent
	lokiLogs    *loki.Loki
	tempoTraces *tempo.Tempo
	manager     *integrations.Manager

	reloadListener net.Listener
	reloadServer   *http.Server
	reloading      *atomic.Bool

	unreg *util.Unregisterer
}

// Reloader is any function that returns a new config.
type Reloader = func() (*config.Config, error)

// NewEntrypoint creates a new Entrypoint.
func NewEntrypoint(logger *util.Logger, cfg *config.Config, reloader Reloader) (*Entrypoint, error) {
	var (
		ep = &Entrypoint{
			log:       logger,
			srvCh:     make(chan *server.Server, 1),
			doneCh:    make(chan bool),
			reloading: atomic.NewBool(false),
			reloader:  reloader,
			unreg:     util.WrapWithUnregisterer(prometheus.DefaultRegisterer),
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

	// Unregister any metrics that the server registered in the last apply.
	srv.unreg.UnregisterAll()

	// Override the server's signal handler to be a no-op. We will
	// create our own SIGINT signal handler when running the server.
	cfg.Server.SignalHandler = newNoopSignalHandler()

	cfg.Server.Registerer = srv.unreg
	if cfg.Server.Log == nil {
		cfg.Server.Log = srv.cfg.Server.Log
	}

	var (
		// wireServer indicates a new server and that all API endpoints
		// (HTTP and gRPC) need to be re-created.
		wireServer bool

		failed bool
		err    error
	)

	if err := srv.log.ApplyConfig(&cfg.Server); err != nil {
		level.Error(srv.log).Log("msg", "failed to update logger", "err", err)
		failed = true
	}

	// Server doesn't have an ApplyConfig method so we need to do a full
	// restart of it here.
	if !compareServer(&srv.cfg.Server, &cfg.Server) {
		level.Info(srv.log).Log("msg", "server configurations changed, restarting server")

		if srv.srv != nil {
			srv.reloading.Store(true)
			srv.srv.Shutdown()
		}

		srv.srv, err = server.New(cfg.Server)
		if err != nil {
			level.Error(srv.log).Log("msg", "failed to reload server", "err", err)
			failed = true
		}

		// New server, so everything needs re-wiring.
		wireServer = true
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

	if wireServer {
		srv.promMetrics.WireAPI(srv.srv.HTTP)
		srv.promMetrics.WireGRPC(srv.srv.GRPC)

		srv.manager.WireAPI(srv.srv.HTTP)

		srv.srv.HTTP.HandleFunc("/-/healthy", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "Agent is Healthy.\n")
		})
		srv.srv.HTTP.HandleFunc("/-/ready", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "Agent is Ready.\n")
		})
		srv.srv.HTTP.HandleFunc("/agent/api/v1/config", func(rw http.ResponseWriter, r *http.Request) {
			srv.mut.Lock()
			bb, err := yaml.Marshal(srv.cfg)
			srv.mut.Unlock()

			if err != nil {
				http.Error(rw, fmt.Sprintf("failed to marshal config: %s", err), http.StatusInternalServerError)
			} else {
				_, _ = rw.Write(bb)
			}
		})

		srv.srv.HTTP.HandleFunc("/-/reload", srv.reloadHandler)

		// The server is finished being wired up, we can run it now.
		srv.srvCh <- srv.srv
	}

	srv.cfg = cfg
	if failed {
		return fmt.Errorf("changes did not apply successfully")
	}

	return nil
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

	srv.closeOnce.Do(func() {
		close(srv.doneCh)
	})

	srv.manager.Stop()
	srv.lokiLogs.Stop()
	srv.promMetrics.Stop()
	srv.tempoTraces.Stop()
	srv.srv.Shutdown()

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

	var (
		serverMut     sync.Mutex
		currentServer *server.Server
	)

	// Our server actor is responsible for reading the last created server
	// and running it.
	//
	// During a reload, the current server will be shut down to create the
	// new one. This scenario must be detected by the actor, and the actor
	// must wait for the new server to be available for running.
	//
	// If the server shuts down independently of a reload, we treat this as
	// a fatal issue and stop the actor.
	g.Add(func() error {
	NextServer:
		for {
			select {
			case <-srv.doneCh:
				return fmt.Errorf("agent exiting")
			case s := <-srv.srvCh:
				serverMut.Lock()
				currentServer = s
				serverMut.Unlock()

				// If the reload failed, s will be nil. Skip this loop and wait for the
				// next recv.
				if s == nil {
					continue NextServer
				}

				err := s.Run()

				// If we're reloading, wait for the next server. There's an edge case
				// where the server shuts down from a problem in the middle of a
				// reload, but given a new server will replace it anyway, it's safe to
				// ignore.
				if srv.reloading.CAS(true, false) {
					continue NextServer
				}

				return err
			}
		}
	}, func(e error) {
		serverMut.Lock()
		defer serverMut.Unlock()

		// Notify the loop of a shutdown. This SHOULD be called before shutting
		// down the server in case a reload is ongoing while we terminating.
		srv.notifyShutdown()

		// Shut down any currently running server.
		if currentServer != nil {
			currentServer.Shutdown()
		}
	})

	return g.Run()
}

// notifyShutdown informs any running actor that the server is shutting down.
func (srv *Entrypoint) notifyShutdown() {
	srv.closeOnce.Do(func() {
		close(srv.doneCh)
	})
}

// noopSignalHandler implements the SignalHandler interface used by
// weaveworks/common/server.
type noopSignalHandler struct {
	ctx    context.Context
	cancel context.CancelFunc
}

func newNoopSignalHandler() *noopSignalHandler {
	var sh noopSignalHandler
	sh.ctx, sh.cancel = context.WithCancel(context.Background())
	return &sh
}

// Equal implements the equality checking interface used by cmp.
func (sh *noopSignalHandler) Equal(*noopSignalHandler) bool {
	return true
}

func (sh *noopSignalHandler) Loop() {
	<-sh.ctx.Done()
}

func (sh *noopSignalHandler) Stop() {
	sh.cancel()
}

// compareServer returns whether two server configs are equal through their YAML configuration.
func compareServer(a *server.Config, b *server.Config) bool {
	aBytes, err := yaml.Marshal(a)
	if err != nil {
		return false
	}
	bBytes, err := yaml.Marshal(b)
	if err != nil {
		return false
	}
	return bytes.Equal(aBytes, bBytes)
}
