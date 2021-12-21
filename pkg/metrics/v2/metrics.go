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
	"fmt"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/pkg/metrics/v2/internal/metricspb"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rfratto/ckit/chash"
	"google.golang.org/grpc"
)

// Metrics runs the metrics subsystem.
type Metrics struct {
	log         log.Logger
	discoverers *discovererManager
	senders     *senderManager
	scrapers    *scraperManager
}

// New creates a new Metrics subsystem. ApplyConfig must be invoked after
// calling New to apply settings.
func New(l log.Logger, reg prometheus.Registerer, opts Options) (*Metrics, error) {
	level.Warn(l).Log("msg", "using metrics-next, which is experimental and subject to change")

	senders := newSenderManager(l, reg, opts)
	scrapers := newScraperManager(l, senders)

	hasher := newHasher(chash.Ring(256), opts.Cluster)
	discoverers := newDiscovererManager(l, hasher, scrapers)

	return &Metrics{
		log:         l,
		senders:     senders,
		scrapers:    scrapers,
		discoverers: discoverers,
	}, nil
}

// ApplyConfig applies cfg to Metrics. May be called multiple times through the
// lifecycle of m.
func (m *Metrics) ApplyConfig(cfg Config) error {
	// NOTE(rfratto): We MUST apply things in the following order:
	//
	// 1. Senders
	// 2. Scrapers
	// 3. Discovery
	//
	// If anything fails, we stop early.

	level.Debug(m.log).Log("msg", "applying config to metrics-next senders")
	if err := m.senders.ApplyConfig(&cfg); err != nil {
		return fmt.Errorf("failed to update senders: %w", err)
	}

	level.Debug(m.log).Log("msg", "applying config to metrics-next scrapers")
	if err := m.scrapers.ApplyConfig(&cfg); err != nil {
		return fmt.Errorf("failed to update scrapers: %w", err)
	}

	level.Debug(m.log).Log("msg", "applying config to metrics-next discoverers")
	if err := m.discoverers.ApplyConfig(&cfg); err != nil {
		return fmt.Errorf("failed to update discoverers: %w", err)
	}

	level.Debug(m.log).Log("msg", "finished applying config to metrics-next")
	return nil
}

// WireAPI wires up gRPC handlers for the metrics subsystem.
func (m *Metrics) WireGRPC(srv *grpc.Server) {
	metricspb.RegisterScraperServer(srv, m.scrapers)
}

// Stop stops the metrics subsystem.
func (m *Metrics) Stop() {
	// NOTE(rfratto): We MUST stop things in the following order:
	//
	// 1. Discovery
	// 2. Scrapers
	// 3. Senders

	m.discoverers.Stop()
	m.scrapers.Stop()

	if err := m.senders.Stop(); err != nil {
		level.Warn(m.log).Log("one or more senders failed to stop", "first_err", err)
	}
}
