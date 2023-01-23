package logs

import (
	"reflect"

	"github.com/grafana/loki/clients/pkg/promtail/client"
)

// DefaultGlobalConfig holds default global settings to be used across all instances.
var DefaultGlobalConfig = GlobalConfig{
	ClientConfigs: []client.Config{},
}

// GlobalConfig holds global settings that apply to all instances by default.
type GlobalConfig struct {
	ClientConfigs []client.Config `yaml:"clients,omitempty"`
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (c *GlobalConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultGlobalConfig

	type plain GlobalConfig
	return unmarshal((*plain)(c))
}

func (c GlobalConfig) IsZero() bool {
	return reflect.DeepEqual(c, GlobalConfig{}) || reflect.DeepEqual(c, DefaultGlobalConfig)
}
