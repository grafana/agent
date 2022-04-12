// Package snmp_exporter embeds https://github.com/prometheus/snmp_exporter
package snmp_exporter

import (
	_ "embed"
	"fmt"

	"github.com/go-kit/log"
	snmp_common "github.com/grafana/agent/pkg/integrations/snmp_exporter/common"
	integrations_v2 "github.com/grafana/agent/pkg/integrations/v2"
	"github.com/grafana/agent/pkg/integrations/v2/common"
	snmp_config "github.com/prometheus/snmp_exporter/config"
)

// DefaultConfig holds the default settings for the snmp_exporter integration.
var DefaultConfig = Config{
	WalkParams:     make(map[string]snmp_config.WalkParams),
	SnmpConfigFile: "",
}

// Config configures the SNMP integration.
type Config struct {
	WalkParams     map[string]snmp_config.WalkParams `yaml:"walk_params,omitempty"`
	SnmpConfigFile string                            `yaml:"config_file,omitempty"`
	SnmpTargets    []SNMPTarget                      `yaml:"snmp_targets"`
	Common         common.MetricsConfig              `yaml:",inline"`

	globals integrations_v2.Globals
}

// SNMPTarget defines a target device to be used by the integration.
type SNMPTarget struct {
	Name       string `yaml:"name"`
	Target     string `yaml:"address"`
	Module     string `yaml:"module"`
	WalkParams string `yaml:"walk_params,omitempty"`
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

// NewIntegration creates a new SNMP integration.
func (c *Config) NewIntegration(log log.Logger, globals integrations_v2.Globals) (integrations_v2.Integration, error) {
	var modules *snmp_config.Config
	var err error
	if c.SnmpConfigFile != "" {
		modules, err = snmp_config.LoadFile(c.SnmpConfigFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load snmp config from file %v: %w", c.SnmpConfigFile, err)
		}
	} else {
		modules, err = snmp_common.LoadEmbeddedConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to load embedded snmp config: %w", err)
		}
	}
	c.globals = globals
	sh := &snmpHandler{
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
	return "snmp"
}

// InstanceKey returns the hostname:port of the agent.
func (c *Config) InstanceKey(agentKey string) (string, error) {
	return agentKey, nil
}

func init() {
	integrations_v2.Register(&Config{}, integrations_v2.TypeSingleton)
}
