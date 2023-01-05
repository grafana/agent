package snowflake_exporter

import (
	"github.com/go-kit/log"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/snowflake-prometheus-exporter/collector"
)

// DefaultConfig is the default config for the snowflake integration
var DefaultConfig = Config{
	Role: "ACCOUNTADMIN",
}

// Config is the configuration for the snowflake integration
type Config struct {
	AccountName string `yaml:"account_name,omitempty"`
	Username    string `yaml:"username,omitempty"`
	Password    string `yaml:"password,omitempty"`
	Role        string `yaml:"role,omitempty"`
	Warehouse   string `yaml:"warehouse,omitempty"`
}

func (c *Config) exporterConfig() *collector.Config {
	return &collector.Config{
		AccountName: c.AccountName,
		Username:    c.Username,
		Password:    c.Password,
		Role:        c.Role,
		Warehouse:   c.Warehouse,
	}
}

// Identifier returns a string that identifies the integration.
func (c *Config) InstanceKey(agentKey string) (string, error) {
	return c.AccountName, nil
}

// UnmarshalYAML implements yaml.Unmarshaler for Config
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultConfig

	type plain Config
	return unmarshal((*plain)(c))
}

// Name returns the name of the integration this config is for.
func (c *Config) Name() string {
	return "snowflake"
}

func init() {
	integrations.RegisterIntegration(&Config{})
}

// NewIntegration creates a new integration from the config.
func (c *Config) NewIntegration(l log.Logger) (integrations.Integration, error) {
	exporterConfig := c.exporterConfig()

	if err := exporterConfig.Validate(); err != nil {
		return nil, err
	}

	col := collector.NewCollector(l, exporterConfig)
	return integrations.NewCollectorIntegration(
		c.Name(),
		integrations.WithCollectors(col),
	), nil
}
