package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/gorilla/mux"
	"github.com/grafana/agent/pkg/logs"
	"github.com/grafana/agent/pkg/metrics"
	"github.com/grafana/agent/pkg/metrics/instance"
	"github.com/grafana/agent/pkg/server"
	"github.com/grafana/agent/pkg/traces"
	"github.com/grafana/agent/pkg/usagestats"
	"github.com/oklog/run"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v2"

	"github.com/grafana/agent/pkg/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/weaveworks/common/signals"

	"github.com/go-kit/log/level"
)

// Entrypoint is the entrypoint of the application that starts all subsystems.
type Entrypoint struct {
	mut sync.Mutex

	reloader Reloader

	log *server.Logger
	cfg config.Config

	srv          *server.Server
	promMetrics  *metrics.Agent
	lokiLogs     *logs.Logs
	tempoTraces  *traces.Traces
	integrations config.Integrations
	reporter     *usagestats.Reporter

	reloadListener net.Listener
	reloadServer   *http.Server
}

// Reloader is any function that returns a new config.
type Reloader = func() (*config.Config, error)

// NewEntrypoint creates a new Entrypoint.
func NewEntrypoint(logger *server.Logger, cfg *config.Config, reloader Reloader) (*Entrypoint, error) {
	var (
		ep = &Entrypoint{
			log:      logger,
			reloader: reloader,
		}
		err error
	)

	ep.srv, err = server.New(logger, prometheus.DefaultRegisterer, prometheus.DefaultGatherer, cfg.Server)
	if err != nil {
		return nil, err
	}

	ep.promMetrics, err = metrics.New(prometheus.DefaultRegisterer, cfg.Metrics, logger)
	if err != nil {
		return nil, err
	}

	ep.lokiLogs, err = logs.New(prometheus.DefaultRegisterer, cfg.Logs, logger)
	if err != nil {
		return nil, err
	}

	ep.tempoTraces, err = traces.New(ep.lokiLogs, ep.promMetrics.InstanceManager(), prometheus.DefaultRegisterer, cfg.Traces, cfg.Server.LogLevel.Logrus, cfg.Server.LogFormat)
	if err != nil {
		return nil, err
	}

	integrationGlobals, err := ep.createIntegrationsGlobals(cfg)
	if err != nil {
		return nil, err
	}

	ep.integrations, err = config.NewIntegrations(logger, &cfg.Integrations, integrationGlobals)
	if err != nil {
		return nil, err
	}

	ep.reporter, err = usagestats.NewReporter(logger, cfg)
	if err != nil {
		return nil, err
	}

	ep.wire(ep.srv.HTTP, ep.srv.GRPC)

	// Mostly everything should be up to date except for the server, which hasn't
	// been created yet.
	if err := ep.ApplyConfig(*cfg); err != nil {
		return nil, err
	}
	return ep, nil
}

func (ep *Entrypoint) createIntegrationsGlobals(cfg *config.Config) (config.IntegrationsGlobals, error) {
	hostname, err := instance.Hostname()
	if err != nil {
		return config.IntegrationsGlobals{}, fmt.Errorf("getting hostname: %w", err)
	}

	var listenPort int
	if ta, ok := ep.srv.HTTPAddress().(*net.TCPAddr); ok {
		listenPort = ta.Port
	}

	return config.IntegrationsGlobals{
		AgentIdentifier: fmt.Sprintf("%s:%d", hostname, listenPort),
		Metrics:         ep.promMetrics,
		Logs:            ep.lokiLogs,
		Tracing:         ep.tempoTraces,
		// TODO(rfratto): set SubsystemOptions here when v1 is removed.

		// Configure integrations to connect to the agent's in-memory server and avoid the network.
		DialContextFunc: ep.srv.DialContext,
		AgentBaseURL: &url.URL{
			Scheme: "http",
			Host:   cfg.Server.Flags.HTTP.InMemoryAddr,
		},
	}, nil
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

	if err := ep.srv.ApplyConfig(cfg.Server); err != nil {
		level.Error(ep.log).Log("msg", "failed to update server", "err", err)
		failed = true
	}

	// Go through each component and update it.
	if err := ep.promMetrics.ApplyConfig(cfg.Metrics); err != nil {
		level.Error(ep.log).Log("msg", "failed to update prometheus", "err", err)
		failed = true
	}

	if err := ep.lokiLogs.ApplyConfig(cfg.Logs); err != nil {
		level.Error(ep.log).Log("msg", "failed to update loki", "err", err)
		failed = true
	}

	if err := ep.tempoTraces.ApplyConfig(ep.lokiLogs, ep.promMetrics.InstanceManager(), cfg.Traces, cfg.Server.LogLevel.Logrus); err != nil {
		level.Error(ep.log).Log("msg", "failed to update traces", "err", err)
		failed = true
	}

	integrationGlobals, err := ep.createIntegrationsGlobals(&cfg)
	if err != nil {
		level.Error(ep.log).Log("msg", "failed to update integrations", "err", err)
		failed = true
	} else if err := ep.integrations.ApplyConfig(&cfg.Integrations, integrationGlobals); err != nil {
		level.Error(ep.log).Log("msg", "failed to update integrations", "err", err)
		failed = true
	}

	ep.cfg = cfg
	if failed {
		return fmt.Errorf("changes did not apply successfully")
	}

	return nil
}

// wire is used to hook up API endpoints to components. It is called once after
// all subsystems are created.
func (ep *Entrypoint) wire(mux *mux.Router, grpc *grpc.Server) {
	ep.promMetrics.WireAPI(mux)
	ep.promMetrics.WireGRPC(grpc)

	ep.integrations.WireAPI(mux)
	ep.lokiLogs.WireAPI(mux)

	mux.HandleFunc("/-/healthy", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Agent is Healthy.\n")
	})

	mux.HandleFunc("/-/ready", func(w http.ResponseWriter, r *http.Request) {
		if !ep.promMetrics.Ready() {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprint(w, "Metrics are not ready yet.\n")

			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Agent is Ready.\n")
	})

	mux.HandleFunc("/-/config", func(rw http.ResponseWriter, r *http.Request) {
		ep.mut.Lock()
		cfg := ep.cfg
		ep.mut.Unlock()

		if cfg.EnableConfigEndpoints {
			bb, err := yaml.Marshal(cfg)
			if err != nil {
				http.Error(rw, fmt.Sprintf("failed to marshal config: %s", err), http.StatusInternalServerError)
			} else {
				_, _ = rw.Write(bb)
			}
		} else {
			rw.WriteHeader(http.StatusNotFound)
			_, _ = rw.Write([]byte("404 - config endpoint is disabled"))
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
	cfg.LogDeprecations(ep.log)

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

	ep.integrations.Stop()
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
	signalHandler := signals.NewHandler(server.GoKitLogger(ep.log))

	notifier := make(chan os.Signal, 1)
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

	srvContext, srvCancel := context.WithCancel(context.Background())
	defer srvCancel()
	defer ep.srv.Close()

	g.Add(func() error {
		return ep.srv.Run(srvContext)
	}, func(e error) {
		srvCancel()
	})

	ep.mut.Lock()
	cfg := ep.cfg
	ep.mut.Unlock()
	if cfg.EnableUsageReport {
		g.Add(func() error {
			return ep.reporter.Start(srvContext)
		}, func(e error) {
			srvCancel()
		})
	}

	go func() {
		for range notifier {
			ep.TriggerReload()
		}
	}()

	return g.Run()
}
