// Package dnsmasq_exporter embeds https://github.com/google/dnsmasq_exporter
package dnsmasq_exporter //nolint:golint

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-kit/kit/log"
	"github.com/google/dnsmasq_exporter/collector"
	"github.com/gorilla/mux"
	"github.com/grafana/agent/pkg/integrations/config"
	"github.com/miekg/dns"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Integration is the dnsmasq_exporter integration. The integration scrapes metrics
// from a dnsmasq server.
type Integration struct {
	c        Config
	logger   log.Logger
	exporter *collector.Collector
}

// New creates a new dnsmasq_exporter integration.
func New(log log.Logger, c Config) (*Integration, error) {
	exporter := collector.New(&dns.Client{
		SingleInflight: true,
	}, c.DnsmasqAddress, c.LeasesPath)

	return &Integration{
		c:        c,
		logger:   log,
		exporter: exporter,
	}, nil
}

// CommonConfig satisfies Integration.CommonConfig.
func (i *Integration) CommonConfig() config.Common { return i.c.CommonConfig }

// Name satisfies Integration.Name.
func (i *Integration) Name() string { return "dnsmasq_exporter" }

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
	if err := r.Register(i.exporter); err != nil {
		return nil, fmt.Errorf("couldn't register dnsmasq_exporter collector: %w", err)
	}
	handler := promhttp.HandlerFor(r, promhttp.HandlerOpts{})
	return handler, nil
}

// ScrapeConfigs satisfies Integration.ScrapeConfigs.
func (i *Integration) ScrapeConfigs() []config.ScrapeConfig {
	return []config.ScrapeConfig{{
		JobName:     i.Name(),
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
