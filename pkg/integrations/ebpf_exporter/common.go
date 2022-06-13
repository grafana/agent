package ebpf

import (
	"github.com/grafana/agent/pkg/integrations"
	integrations_v2 "github.com/grafana/agent/pkg/integrations/v2"
	"github.com/grafana/agent/pkg/integrations/v2/metricsutils"

	ebpf_config "github.com/cloudflare/ebpf_exporter/config"
)

func init() {
	integrations.RegisterIntegration(&Config{})
	integrations_v2.RegisterLegacy(&Config{}, integrations_v2.TypeSingleton, metricsutils.NewNamedShim("ebpf"))
}

// Config controls the eBPF integration.
type Config struct {
	Programs []ebpf_config.Program `yaml:"programs,omitempty"`
}

// UnmarshalYAML implements yaml.Unmarshaler for Config.
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain Config

	return unmarshal((*plain)(c))
}

// Name returns the name of the integration that this config represents.
func (c *Config) Name() string {
	return "ebpf"
}

// InstanceKey returns a set identifier for the ebpf_exporter integration.
func (c *Config) InstanceKey(_ string) (string, error) {
	return c.Name(), nil
}
