package ebpf

import (
	"github.com/grafana/agent/pkg/integrations/v2"
	"github.com/grafana/agent/pkg/integrations/v2/common"

	ebpf_config "github.com/cloudflare/ebpf_exporter/config"
)

func init() {
	integrations.Register(&Config{}, integrations.TypeSingleton)
}

// DefaultConfig for the ebpf_exporter.
var DefaultConfig = Config{
	Programs: []ebpf_config.Program{},
}

// Config controls the eBPF integration.
type Config struct {
	Programs []ebpf_config.Program `yaml:"programs,omitempty"`
	Common   common.MetricsConfig  `yaml:",inline"`
}

// ApplyDefaults passes in configuration from the globals config.
func (c *Config) ApplyDefaults(globals integrations.Globals) error {
	c.Common.ApplyDefaults(globals.SubsystemOpts.Metrics.Autoscrape)
	return nil
}

// Identifier returns a set identifier for the ebpf_exporter integration.
func (c *Config) Identifier(globals integrations.Globals) (string, error) { return c.Name(), nil }

// Name returns the name of the integration that this config represents.
func (c *Config) Name() string { return "ebpf" }

// UnmarshalYAML implements yaml.Unmarshaler for Config.
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultConfig
	type plain Config

	return unmarshal((*plain)(c))
}
