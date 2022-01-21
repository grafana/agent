// Package agent is an "example" integration that has very little functionality,
// but is still useful in practice. The Agent integration re-exposes the Agent's
// own metrics endpoint and allows the Agent to scrape itself.
package agent

import (
	"github.com/grafana/agent/pkg/integrations/shared"
	"github.com/grafana/agent/pkg/integrations/v2/common"
)

// Config controls the Agent integration.
type Config struct {
	Common common.MetricsConfig `yaml:",inline"`
}

// Name returns the name of the integration that this config represents.
func (c *Config) Name() string { return "agent" }

// ApplyDefaults applies runtime-specific defaults to c.
func (c *Config) ApplyDefaults(globals shared.Globals) error {
	c.Common.ApplyDefaults(globals.Autoscrape)
	if id, err := c.Identifier(globals); err == nil {
		c.Common.InstanceKey = &id
	}
	return nil
}

// Identifier uniquely identifies this instance of Config.
func (c *Config) Identifier(globals shared.Globals) (string, error) {
	if c.Common.InstanceKey != nil {
		return *c.Common.InstanceKey, nil
	}
	return globals.AgentIdentifier, nil
}
