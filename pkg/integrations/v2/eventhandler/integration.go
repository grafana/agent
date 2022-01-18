package eventhandler

import (
	"github.com/go-kit/log"
	"github.com/grafana/agent/pkg/integrations/v2"
)

// Config controls the EventHandler integration.
type Config struct {
}

// Name returns the name of the integration that this config represents.
func (c *Config) Name() string { return "eventhandler" }

// ApplyDefaults applies runtime-specific defaults to c.
func (c *Config) ApplyDefaults(globals integrations.Globals) error {
	return nil
}

// Identifier uniquely identifies this instance of Config.
func (c *Config) Identifier(globals integrations.Globals) (string, error) {
	return globals.AgentIdentifier, nil
}

// NewIntegration converts this config into an instance of an integration.
func (c *Config) NewIntegration(l log.Logger, globals integrations.Globals) (integrations.Integration, error) {
	return newEventHandler(l, c, globals)
}

func init() {
	integrations.Register(&Config{}, integrations.TypeSingleton)
}
