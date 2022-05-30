package ebpf

import (
	ebpf_config "github.com/cloudflare/ebpf_exporter/config"
	"github.com/grafana/agent/pkg/integrations/v2"
)

func init() {
	integrations.Register(&Config{}, integrations.TypeSingleton)
}

var defaultConfig = Config{
	Programs: []ebpf_config.Program{},
}

// Config controls the ebpf integration
type Config struct {
	Programs []ebpf_config.Program `yaml:"programs,omitempty"`
}

func (c *Config) ApplyDefaults(globals integrations.Globals) error        { return nil }
func (c *Config) Identifier(globals integrations.Globals) (string, error) { return c.Name(), nil }
func (c *Config) Name() string                                            { return "ebpf" }

// UnmarshalYAML implements yaml.Unmarshaler for Config.
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = defaultConfig
	type plain Config

	return unmarshal((*plain)(c))
}
