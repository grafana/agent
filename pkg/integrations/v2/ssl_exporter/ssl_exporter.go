// Package ssl_exporter embeds https://github.com/ribbybibby/ssl_exporter/v2
package ssl_exporter //nolint:golint

import (
	"fmt"

	"github.com/go-kit/log"
	integrations_v2 "github.com/grafana/agent/pkg/integrations/v2"
	"github.com/grafana/agent/pkg/integrations/v2/common"
	ssl_config "github.com/ribbybibby/ssl_exporter/v2/config"
)

// DefaultConfig holds the default settings for the ssl_exporter integration.
var DefaultConfig = Config{
	ConfigFile: "",
	SSLTargets: []SSLTarget{},
}

// SSLTarget represents a target to be scraped.
type SSLTarget struct {
	Name   string `yaml:"name"`
	Target string `yaml:"target"`
	Module string `yaml:"module"`
}

// Config controls the ssl_exporter integration.
type Config struct {
	ConfigFile string               `yaml:"key_file,omitempty"`
	SSLTargets []SSLTarget          `yaml:"ssl_targets"`
	Common     common.MetricsConfig `yaml:",inline"`

	globals integrations_v2.Globals
}

// ApplyDefaults applies the integration's default configuration.
func (c *Config) ApplyDefaults(globals integrations_v2.Globals) error {
	c.Common.ApplyDefaults(globals.SubsystemOpts.Metrics.Autoscrape)
	return nil
}

// Identifier returns a string that identifies the integration.
func (c *Config) Identifier(globals integrations_v2.Globals) (string, error) {
	return c.Name(), nil
}

// NewIntegration converts the config into an instance of an integration.
func (c *Config) NewIntegration(log log.Logger, globals integrations_v2.Globals) (integrations_v2.Integration, error) {
	var modules *ssl_config.Config
	var err error

	modules = ssl_config.DefaultConfig
	if c.ConfigFile != "" {
		modules, err = ssl_config.LoadConfig(c.ConfigFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load ssl config from file %v: %w", c.ConfigFile, err)
		}
	}

	// The `name` and `target` fields are mandatory for the ssl targets are mandatory.
	// Enforce this check and fail the creation of the integration if they're missing.
	for _, target := range c.SSLTargets {
		if target.Name == "" || target.Target == "" {
			return nil, fmt.Errorf("failed to load ssl_targets; the `name` and `target` fields are mandatory")
		}
	}

	c.globals = globals
	sh := &sslHandler{
		cfg:     c,
		modules: modules,
		log:     log,
	}

	return sh, nil
}

// UnmarshalYAML implements yaml.Unmarshaler for Config.
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultConfig

	type plain Config
	return unmarshal((*plain)(c))
}

// Name returns the name of the integration.
func (c *Config) Name() string {
	return "ssl"
}

// InstanceKey returns the hostname:port of the agent.
func (c *Config) InstanceKey(agentKey string) (string, error) {
	return agentKey, nil
}

func init() {
	integrations_v2.Register(&Config{}, integrations_v2.TypeSingleton)
}
