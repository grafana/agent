//go:build !darwin
// +build !darwin

package node_exporter

import (
	"context"
	"net/http"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/integrations/config"
)

func init() {
	// Register macos_exporter
	integrations.RegisterIntegration(&DummyDarwinConfig{})
}

// DummyDarwinConfig extends the Config struct and overrides the name of
// the integration to avoid conflicts with node_exporter integration.
type DummyDarwinConfig struct{}

func (*DummyDarwinConfig) CommonConfig() config.Common {
	return config.Common{}
}

func (*DummyDarwinConfig) InstanceKey(agentKey string) (string, error) {
	return agentKey, nil
}

// Name returns the name of the integration that this config represents.
func (*DummyDarwinConfig) Name() string {
	return "macos_exporter"
}

// NewIntegration converts this config into an instance of an integration.
func (c *DummyDarwinConfig) NewIntegration(l log.Logger) (integrations.Integration, error) {
	level.Warn(l).Log("msg", "the macos_exporter only works on Darwin; enabling it otherwise will do nothing")
	return DarwinIntegration{}, nil
}

// DarwinIntegration is the macos_integration integration. On non-Darwin platforms,
// this integration does nothing and will print a warning if enabled.
type DarwinIntegration struct{}

// MetricsHandler satisfies Integration.RegisterRoutes.
func (DarwinIntegration) MetricsHandler() (http.Handler, error) {
	return http.NotFoundHandler(), nil
}

// ScrapeConfigs satisfies Integration.ScrapeConfigs.
func (DarwinIntegration) ScrapeConfigs() []config.ScrapeConfig {
	// No-op: nothing to scrape.
	return []config.ScrapeConfig{}
}

// Run satisfies Integration.Run.
func (DarwinIntegration) Run(ctx context.Context) error {
	// We don't need to do anything here, so we can just wait for the context to
	// finish.
	<-ctx.Done()
	return ctx.Err()
}
