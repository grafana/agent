package main

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/loki"
	"github.com/grafana/agent/pkg/tempo"
	"github.com/grafana/agent/pkg/util"

	"github.com/grafana/agent/pkg/config"
	"github.com/grafana/agent/pkg/prom"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/weaveworks/common/logging"
	"github.com/weaveworks/common/server"

	"github.com/go-kit/kit/log/level"
)

// Entrypoint is the entrypoint of the application that starts all subsystems.
type Entrypoint struct {
	mut sync.Mutex

	log *util.Logger
	cfg config.Config

	srv         *server.Server
	promMetrics *prom.Agent
	lokiLogs    *loki.Loki
	tempoTraces *tempo.Tempo
	manager     *integrations.Manager
}

// NewEntrypoint creates a new Entrypoint.
func NewEntrypoint(logger *util.Logger, cfg *config.Config) (*Entrypoint, error) {
	var (
		ep  = Entrypoint{log: logger}
		err error
	)

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
	return &ep, nil
}

// ApplyConfig applies changes to the subsystems of the Agent.
func (srv *Entrypoint) ApplyConfig(cfg config.Config) error {
	srv.mut.Lock()
	defer srv.mut.Unlock()

	// The server config uses some unexported fields which can't be compared by
	// default. Since only exported fields are used by YAML, we'll only compare
	// those here.
	var ignoreUnexported = cmpopts.IgnoreUnexported(logging.Format{}, logging.Level{})

	if cmp.Equal(srv.cfg, cfg, ignoreUnexported) {
		return nil
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
	if !cmp.Equal(srv.cfg.Server, cfg.Server, ignoreUnexported) {
		if srv.srv != nil {
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
	}

	srv.cfg = cfg
	if failed {
		return fmt.Errorf("changes did not apply successfully")
	}
	return nil
}

// Stop stops the Entrypoint and all subsystems.
func (srv *Entrypoint) Stop() {
	srv.manager.Stop()
	srv.lokiLogs.Stop()
	srv.promMetrics.Stop()
	srv.tempoTraces.Stop()
	srv.srv.Shutdown()
}

// Start starts the server used by the Entrypoint, and will block until a
// termination signal is sent to the process.
func (srv *Entrypoint) Start() error {
	return srv.srv.Run()
}
