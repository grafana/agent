// +build windows

package windows_exporter //nolint:golint

import (
	"context"
	"fmt"
	"net/http"

	"gopkg.in/yaml.v2"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/gorilla/mux"
	"github.com/grafana/agent/pkg/integrations/config"
	"github.com/prometheus-community/windows_exporter/exporter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Integration struct {
	c      *Config
	logger log.Logger
	wc     *exporter.WindowsCollector

	exporterMetricsRegistry *prometheus.Registry
}

// New creates a new node_exporter integration.
func New(log log.Logger, c *Config) (*Integration, error) {

	bytes, _ := yaml.Marshal(c.WindowsConfig)
	cb := string(bytes)
	wc, _ := exporter.NewWindowsCollector(c.Name(), c.EnabledCollectors, cb)

	level.Info(log).Log("msg", "Enabled windows_exporter collectors")

	return &Integration{
		c:      c,
		logger: log,
		wc:     wc,

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
	if err := r.Register(i.wc); err != nil {
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
