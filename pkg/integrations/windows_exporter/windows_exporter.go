package windows_exporter //nolint:golint

import (
	"context"
	"fmt"
	"net/http"

	"github.com/prometheus/node_exporter/collector"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/gorilla/mux"
	"github.com/grafana/agent/pkg/integrations/config"
	windows_collector "github.com/prometheus-community/windows_exporter/collector"
	"github.com/prometheus-community/windows_exporter/exporter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Integration is the node_exporter integration. The integration scrapes metrics
// from the host Linux-based system.
type Integration struct {
	c      *Config
	logger log.Logger
	nc     *collector.NodeCollector

	exporterMetricsRegistry *prometheus.Registry
}

// New creates a new node_exporter integration.
func New(log log.Logger, c *Config) (*Integration, error) {

	mappedConfig := make(map[string]*windows_collector.ConfigInstance)
	wc := exporter.NewWindowsCollector(mappedConfig)

	level.Info(log).Log("msg", "Enabled windows_exporter collectors")

	return &Integration{
		c:      c,
		logger: log,
		nc:     wc,

		exporterMetricsRegistry: prometheus.NewRegistry(),
	}, nil
}

// RegisterRoutes satisfies Integration.RegisterRoutes. The mux.Router provided
// here is expected to be a subrouter, where all registered paths will be
// registered within that subroute.
func (i *Integration) RegisterRoutes(r *mux.Router) error {
	handler, err := i.handler()
	if err != nil {
		return err
	}

	r.Handle("/metrics", handler)
	return nil
}

func (i *Integration) handler() (http.Handler, error) {
	r := prometheus.NewRegistry()
	if err := r.Register(i.nc); err != nil {
		return nil, fmt.Errorf("couldn't register windows_exporter collector: %w", err)
	}
	handler := promhttp.HandlerFor(
		prometheus.Gatherers{i.exporterMetricsRegistry, r},
		promhttp.HandlerOpts{
			ErrorHandling:       promhttp.ContinueOnError,
			MaxRequestsInFlight: 0,
			Registry:            i.exporterMetricsRegistry,
		},
	)

	return handler, nil
}

// ScrapeConfigs satisfies Integration.ScrapeConfigs.
func (i *Integration) ScrapeConfigs() []config.ScrapeConfig {
	return []config.ScrapeConfig{{
		JobName:     i.c.Name(),
		MetricsPath: "/metrics",
	}}
}

// Run satisfies Integration.Run.
func (i *Integration) Run(ctx context.Context) error {
	// We don't need to do anything here, so we can just wait for the context to
	// finish.
	<-ctx.Done()
	return ctx.Err()
}
