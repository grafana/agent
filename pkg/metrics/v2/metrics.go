// Package metrics implements the metrics subsystem of Grafana Agent. The
// metrics subsystem has two primary capabilities:
//
// 1. Discover and collect Prometheus metrics
// 2. Send collected Prometheus metrics to a compatible remote_write endpoint.
//
// The metrics subsystem is cluster aware. If an agent is participating in a
// cluster, the responsibility of collecting Prometheus metrics will be
// distributed throughout all known agents.
package metrics

import (
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
)

// Metrics runs the metrics subsystem.
type Metrics struct {
}

// New creates a new Metrics subsystem. ApplyConfig must be invoked after
// calling New to apply settings.
func New(l log.Logger, reg prometheus.Registerer, opts Options) (*Metrics, error) {
	level.Warn(l).Log("msg", "using metrics-next, which is experimental and subject to change")

	// TODO(rfratto): implement
	return &Metrics{}, nil
}

// ApplyConfig applies cfg to Metrics. May be called multiple times through the
// lifecycle of m.
func (m *Metrics) ApplyConfig(cfg Config) error {
	return nil
}

// WireAPI wires up HTTP API routes for the metrics subsystem.
func (m *Metrics) WireAPI(r *mux.Router) {

}

// WireAPI wires up gRPC handlers for the metrics subsystem.
func (m *Metrics) WireGRPC(srv *grpc.Server) {

}

// Stop stops the metrics subsystem.
func (m *Metrics) Stop() {

}
