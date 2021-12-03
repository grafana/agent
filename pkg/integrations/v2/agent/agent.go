// Package agent is an "example" integration that has very little functionality,
// but is still useful in practice. The Agent integration re-exposes the Agent's
// own metrics endpoint and allows the Agent to scrape itself.
package agent

import (
	"github.com/go-kit/log"
	"github.com/grafana/agent/pkg/integrations/v2"
	"github.com/grafana/agent/pkg/integrations/v2/metricsutils"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Config controls the Agent integration.
type Config struct {
	metricsutils.CommonConfig `yaml:",inline"`
}

// Name returns the name of the integration that this config represents.
func (c *Config) Name() string { return "agent" }

// Identifier uniquely identifies this instance of Config.
func (c *Config) Identifier(globals integrations.Globals) (string, error) {
	if c.InstanceKey != nil {
		return *c.InstanceKey, nil
	}
	return globals.AgentIdentifier, nil
}

// MetricsConfig implements metricsutils.MetricsConfig.
func (c *Config) MetricsConfig() metricsutils.CommonConfig { return c.CommonConfig }

// NewIntegration converts this config into an instance of an integration.
func (c *Config) NewIntegration(l log.Logger, globals integrations.Globals) (integrations.Integration, error) {
	return metricsutils.NewMetricsHandlerIntegration(l, c, globals, promhttp.Handler())
}

func init() {
	integrations.Register(&Config{}, integrations.TypeSingleton)
}
