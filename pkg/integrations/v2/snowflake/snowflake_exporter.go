package snowflake_exporter

import (
	"fmt"
	"net/http"

	"github.com/go-kit/log"
	"github.com/grafana/snowflake-prometheus-exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	integrations_v2 "github.com/grafana/agent/pkg/integrations/v2"
	"github.com/grafana/agent/pkg/integrations/v2/common"
	"github.com/grafana/agent/pkg/integrations/v2/metricsutils"
)

// DefaultConfig is the default config for the snowflake v2 integration
var DefaultConfig = Config{
	Role: "ACCOUNTADMIN",
}

// Config is the configuration for the snowflake v2 integration
type Config struct {
	AccountName string               `yaml:"account_name,omitempty"`
	Username    string               `yaml:"username,omitempty"`
	Password    string               `yaml:"password,omitempty"`
	Role        string               `yaml:"role,omitempty"`
	Warehouse   string               `yaml:"warehouse,omitempty"`
	Common      common.MetricsConfig `yaml:",inline"`
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
	integrations_v2.Register(&Config{}, integrations_v2.TypeMultiplex)
}

// NewIntegration creates a new v2 integration from the config.
func (c *Config) NewIntegration(l log.Logger, globals integrations_v2.Globals) (integrations_v2.Integration, error) {
	handler, err := createHandler(l, c)
	if err != nil {
		return nil, fmt.Errorf("failed to create http handler: %w", err)
	}

	return metricsutils.NewMetricsHandlerIntegration(l, c, c.Common, globals, handler)
}

func createHandler(logger log.Logger, c *Config) (http.HandlerFunc, error) {
	exporterConfig := c.exporterConfig()
	if err := exporterConfig.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}

	snowflakeCol := collector.NewCollector(logger, exporterConfig)

	registry := prometheus.NewRegistry()
	registry.MustRegister(snowflakeCol)
	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})

	return h.ServeHTTP, nil
}
