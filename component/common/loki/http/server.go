package http

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/gorilla/mux"
	"github.com/prometheus/common/model"
	"github.com/weaveworks/common/logging"
	"github.com/weaveworks/common/server"

	"github.com/grafana/loki/clients/pkg/promtail/targets/serverutils"
)

// TargetServer is wrapper around WeaveWorks Server that handled some common configuration used in all flow source
// components that expose a network server. It just handles configuration and initialization, the handlers implementation
// are left to the consumer.
type TargetServer struct {
	logger           log.Logger
	config           Config
	metricsNamespace string
	server           *server.Server
}

// NewTargetServer creates a new TargetServer, applying some defaults to the server configuration.
func NewTargetServer(logger log.Logger, metricsNamespace string, reg prometheus.Registerer, config Config) (*TargetServer, error) {
	if !model.IsValidMetricName(model.LabelValue(metricsNamespace)) {
		return nil, fmt.Errorf("metrics namespace is not prometheus compatiible: %s", metricsNamespace)
	}

	t := &TargetServer{
		logger:           logger,
		metricsNamespace: metricsNamespace,
		config:           config,
	}

	mergedServerConfigs, err := serverutils.MergeWithDefaults(config.Server)
	if err != nil {
		return nil, fmt.Errorf("failed to parse server configs and override defaults: %w", err)
	}
	// Set the config to the new combined config.
	config.Server = mergedServerConfigs
	// Avoid logging entire received request on failures
	config.Server.ExcludeRequestInLog = true
	// Configure dedicated metrics registerer
	config.Server.Registerer = reg

	return t, nil
}

// MountAndRun does some final configuration of the WeaveWorks server, before mounting the handlers and starting the server.
func (ts *TargetServer) MountAndRun(mountRoute func(router *mux.Router)) error {
	level.Info(ts.logger).Log("msg", "starting server")

	// To prevent metric collisions because all metrics are going to be registered in the global Prometheus registry.
	ts.config.Server.MetricsNamespace = ts.metricsNamespace

	// We don't want the /debug and /metrics endpoints running, since this is not the main promtail HTTP server.
	// We want this target to expose the least surface area possible, hence disabling WeaveWorks HTTP server metrics
	// and debugging functionality.
	ts.config.Server.RegisterInstrumentation = false

	ts.config.Server.Log = logging.GoKit(ts.logger)
	srv, err := server.New(ts.config.Server)
	if err != nil {
		return err
	}

	ts.server = srv
	mountRoute(ts.server.HTTP)

	go func() {
		err := srv.Run()
		if err != nil {
			level.Error(ts.logger).Log("msg", "server shutdown with error", "err", err)
		}
	}()

	return nil
}

// HTTPListenAddr returns the listen address of the HTTP server, if configured.
func (ts *TargetServer) HTTPListenAddr() string {
	return ts.server.HTTPListenAddr().String()
}

// GRPCListenAddr returns the listen address of the gRPC server, if configured.
func (ts *TargetServer) GRPCListenAddr() string {
	return ts.server.GRPCListenAddr().String()
}

// StopAndShutdown stops and shuts down the underlying server.
func (ts *TargetServer) StopAndShutdown() {
	ts.server.Stop()
	ts.server.Shutdown()
}
