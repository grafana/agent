// Package agent is an "example" integration that has very little functionality,
// but is still useful in practice. The Agent integration re-exposes the Agent's
// own metrics endpoint and allows the Agent to scrape itself.
package agent

import (
	"github.com/grafana/agent/pkg/integrations/v2"
	"github.com/grafana/agent/pkg/integrations/v2/metricsutils"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Config controls the Agent integration.
type Config struct {
	Common metricsutils.CommonConfig `yaml:",inline"`
}

// Name returns the name of the integration that this config represents.
func (c *Config) Name() string { return "agent" }

// Identifier uniquely identifies this instance of Config.
func (c *Config) Identifier(opts integrations.IntegrationOptions) (string, error) {
	if c.Common.InstanceKey != nil {
		return *c.Common.InstanceKey, nil
	}
	return opts.AgentIdentifier, nil
}

// NewIntegration converts this config into an instance of an integration.
func (c *Config) NewIntegration(opts integrations.IntegrationOptions) (integrations.Integration, error) {
	return metricsutils.NewMetricsHandlerIntegration(c, c.Common, opts, promhttp.Handler())
}

func init() {
	integrations.Register(&Config{}, integrations.TypeSingleton)
}
