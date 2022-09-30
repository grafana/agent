package blackbox_exporter_v2

import (
	"github.com/go-kit/log"
	"github.com/grafana/agent/pkg/integrations/blackbox_exporter"
	integrations_v2 "github.com/grafana/agent/pkg/integrations/v2"
	"github.com/grafana/agent/pkg/integrations/v2/common"
	blackbox_config "github.com/prometheus/blackbox_exporter/config"
)

// DefaultConfig holds the default settings for the blackbox_exporter integration.
var DefaultConfig = Config{
	// Default value taken from https://github.com/prometheus/blackbox_exporter/blob/master/main.go#L61
	ProbeTimeoutOffset: 0.5,
}

// Config configures the Blackbox integration.
type Config struct {
	BlackboxConfigFile string                             `yaml:"config_file,omitempty"`
	BlackboxTargets    []blackbox_exporter.BlackboxTarget `yaml:"blackbox_targets"`
	BlackboxConfig     blackbox_config.Config             `yaml:"blackbox_config,omitempty"`
	ProbeTimeoutOffset float64                            `yaml:"probe_timeout_offset,omitempty"`

	Common  common.MetricsConfig `yaml:",inline"`
	globals integrations_v2.Globals
}

// UnmarshalYAML implements yaml.Unmarshaler for Config.
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultConfig

	type plain Config
	return unmarshal((*plain)(c))
}

// Name returns the name of the integration.
func (c *Config) Name() string {
	return "blackbox"
}

// ApplyDefaults applies the integration's default configuration.
func (c *Config) ApplyDefaults(globals integrations_v2.Globals) error {
	c.Common.ApplyDefaults(globals.SubsystemOpts.Metrics.Autoscrape)
	return nil
}

// Identifier returns a string that identifies the integration.
func (c *Config) Identifier(globals integrations_v2.Globals) (string, error) {
	if c.Common.InstanceKey != nil {
		return *c.Common.InstanceKey, nil
	}
	return c.Name(), nil
}

func init() {
	integrations_v2.Register(&Config{}, integrations_v2.TypeSingleton)
}

// NewIntegration creates a new blackbox integration.
func (c *Config) NewIntegration(log log.Logger, globals integrations_v2.Globals) (integrations_v2.Integration, error) {
	modules, err := blackbox_exporter.LoadBlackboxConfig(log, c.BlackboxConfigFile, c.BlackboxTargets, &c.BlackboxConfig)
	if err != nil {
		return nil, err
	}

	c.globals = globals
	bbh := &blackboxHandler{
		cfg:     c,
		modules: modules,
		log:     log,
	}
	return bbh, nil
}
