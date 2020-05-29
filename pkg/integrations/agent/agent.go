// Package agent is a "example" integration that does very little functionally,
// but is still useful in practice. The Agent integration re-exposes the Agent's
// own metrics endpoint and allows the Agent to scrape itself.
package agent

import (
	"context"
	"flag"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Config controls the Agent integration.
type Config struct {
	// Enabled enables the Agent integration.
	Enabled bool
}

func (c *Config) RegisterFlagsWithPrefix(prefix string, f *flag.FlagSet) {
	f.BoolVar(&c.Enabled, prefix+"agent.enabled", false, "enable the Agent integration to scrape the Agent's own metrics")
}

// Integration is the Agent integration. The Agent integration scrapes the
// Agent's own metrics.
type Integration struct{}

func New() *Integration {
	return &Integration{}
}

// Name satisfies Integration.Name.
func (i *Integration) Name() string { return "agent" }

// RegisterRoutes satisfies Integration.RegisterRoutes.
func (i *Integration) RegisterRoutes(r *mux.Router) error {
	// Note that if the weaveworks common server is set to not register
	// instrumentation endpoints, this lets the agent integration still be able
	// to scrape itself, just at /integrations/agent/metrics.
	r.Handle("/metrics", promhttp.Handler())

	return nil
}

// MetricsEndpoints satisfies Integration.MetricsEndpoints.
func (i *Integration) MetricsEndpoints() []string {
	return []string{"/metrics"}
}

// Run satisfies Integration.Run.
func (i *Integration) Run(ctx context.Context) error {
	// We don't need to do anything here, so we can just wait for the context to
	// finish.
	<-ctx.Done()
	return ctx.Err()
}
