// Package snmp_exporter embeds https://github.com/prometheus/snmp_exporter
package snmp_exporter

import (
	_ "embed"

	"github.com/go-kit/log"
	"github.com/grafana/agent/pkg/integrations"
	integrations_v2 "github.com/grafana/agent/pkg/integrations/v2"
	"github.com/grafana/agent/pkg/integrations/v2/metricsutils"
	snmp_config "github.com/prometheus/snmp_exporter/config"
	"gopkg.in/yaml.v2"
)

// DefaultConfig holds the default settings for the snmp_exporter integration.
var DefaultConfig = Config{
	WalkParams: make(map[string]snmp_config.WalkParams),
}

//go:embed snmp.yml
var content []byte

type Config struct {
	WalkParams map[string]snmp_config.WalkParams `yaml:"walk_params,omitempty"`
}

// UnmarshalYAML implements yaml.Unmarshaler for Config.
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultConfig

	type plain Config
	return unmarshal((*plain)(c))
}

// Name returns the name of the integration.
func (c *Config) Name() string {
	return "snmp_exporter"
}

// InstanceKey returns the hostname:port of the agent.
func (c *Config) InstanceKey(agentKey string) (string, error) {
	return agentKey, nil
}

// NewIntegration converts the config into an instance of an integration.
func (c *Config) NewIntegration(l log.Logger) (integrations.Integration, error) {
	return New(l, c)
}

func init() {
	integrations.RegisterIntegration(&Config{})
	integrations_v2.RegisterLegacy(&Config{}, integrations_v2.TypeMultiplex, metricsutils.CreateShim)
}

// Load from file via embed
func LoadConfig() (*snmp_config.Config, error) {

	cfg := &snmp_config.Config{}
	err := yaml.UnmarshalStrict(content, cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

// New creates a new snmp_exporter integration
func New(log log.Logger, c *Config) (integrations.Integration, error) {
	modules, err := LoadConfig()

	sh := &snmpHandler{
		cfg:     c,
		modules: modules,
		log:     log,
	}

	if err != nil {
		log.Log("Failed to load config")
	}
	return integrations.NewHandlerIntegration(c.Name(), sh), nil

}
