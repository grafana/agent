package integrations

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/grafana/agent/pkg/integrations/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
)

// CollectorIntegration is an integration exposing metrics from a Prometheus
// collector.
type CollectorIntegration struct {
	name                   string
	c                      prometheus.Collector
	includeExporterMetrics bool
}

// NewCollectorIntegration creates a basic integration that exposes metrics
// from a prometheus.Collector.
func NewCollectorIntegration(name string, c prometheus.Collector, includeExporterMetrics bool) *CollectorIntegration {
	return &CollectorIntegration{
		name:                   name,
		c:                      c,
		includeExporterMetrics: includeExporterMetrics,
	}
}

// RegisterRoutes satisfies Integration.RegisterRoutes. The mux.Router provided
// here is expected to be a subrouter, where all registered paths will be
// registered within that subroute.
func (i *CollectorIntegration) RegisterRoutes(r *mux.Router) error {
	handler, err := i.handler()
	if err != nil {
		return err
	}

	r.Handle("/metrics", handler)
	return nil
}

func (i *CollectorIntegration) handler() (http.Handler, error) {
	r := prometheus.NewRegistry()
	if err := r.Register(i.c); err != nil {
		return nil, fmt.Errorf("couldn't register %s: %w", i.name, err)
	}

	// Register <integration name>_build_info metrics, generally useful for
	// dashboards that depend on them for discovering targets.
	if err := r.Register(version.NewCollector(i.name)); err != nil {
		return nil, fmt.Errorf("couldn't register %s: %w", i.name, err)
	}

	handler := promhttp.HandlerFor(
		r,
		promhttp.HandlerOpts{
			ErrorHandling: promhttp.ContinueOnError,
		},
	)

	if i.includeExporterMetrics {
		// Note that we have to use reg here to use the same promhttp metrics for
		// all expositions.
		handler = promhttp.InstrumentMetricHandler(r, handler)
	}

	return handler, nil
}

// ScrapeConfigs satisfies Integration.ScrapeConfigs.
func (i *CollectorIntegration) ScrapeConfigs() []config.ScrapeConfig {
	return []config.ScrapeConfig{{
		JobName:     i.name,
		MetricsPath: "/metrics",
	}}
}

// Run satisfies Integration.Run.
func (i *CollectorIntegration) Run(ctx context.Context) error {
	// We don't need to do anything here, so we can just wait for the context to
	// finish.
	<-ctx.Done()
	return ctx.Err()
}
