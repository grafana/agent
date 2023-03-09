package mssql

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
	config_util "github.com/prometheus/common/config"

	"github.com/burningalchemist/sql_exporter"
	"github.com/burningalchemist/sql_exporter/config"
	"github.com/grafana/agent/pkg/integrations"
	integrations_v2 "github.com/grafana/agent/pkg/integrations/v2"
	"github.com/grafana/agent/pkg/integrations/v2/metricsutils"
	"github.com/prometheus/common/model"
)

// DefaultConfig is the default config for the mssql integration
var DefaultConfig = Config{
	MaxIdleConnections: 3,
	MaxOpenConnections: 3,
	Timeout:            10 * time.Second,
}

// Config is the configuration for the mssql integration
type Config struct {
	ConnectionString   config_util.Secret `yaml:"connection_string,omitempty"`
	MaxIdleConnections int                `yaml:"max_idle_connections,omitempty"`
	MaxOpenConnections int                `yaml:"max_open_connections,omitempty"`
	Timeout            time.Duration      `yaml:"timeout,omitempty"`
}

func (c Config) validate() error {
	if c.ConnectionString == "" {
		return errors.New("the connection_string parameter is required")
	}

	url, err := url.Parse(string(c.ConnectionString))
	if err != nil {
		return fmt.Errorf("failed to parse connection_string: %w", err)
	}

	if url.Scheme != "sqlserver" {
		return errors.New("scheme of provided connection_string URL must be sqlserver")
	}

	if c.MaxOpenConnections < 1 {
		return errors.New("max_connections must be at least 1")
	}

	if c.MaxIdleConnections < 1 {
		return errors.New("max_idle_connection must be at least 1")
	}

	if c.Timeout <= 0 {
		return errors.New("timeout must be positive")
	}

	return nil
}

// Identifier returns a string that identifies the integration.
func (c *Config) InstanceKey(agentKey string) (string, error) {
	url, err := url.Parse(string(c.ConnectionString))
	if err != nil {
		return "", fmt.Errorf("failed to parse connection string URL: %w", err)
	}

	return url.Host, nil
}

// UnmarshalYAML implements yaml.Unmarshaler for Config
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultConfig

	type plain Config
	return unmarshal((*plain)(c))
}

// Name returns the name of the integration this config is for.
func (c *Config) Name() string {
	return "mssql"
}

func init() {
	integrations.RegisterIntegration(&Config{})
	integrations_v2.RegisterLegacy(&Config{}, integrations_v2.TypeMultiplex, metricsutils.NewNamedShim("mssql"))
}

// NewIntegration creates a new integration from the config.
func (c *Config) NewIntegration(l log.Logger) (integrations.Integration, error) {
	if err := c.validate(); err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}

	t, err := sql_exporter.NewTarget(
		"mssqlintegration",
		"",
		string(c.ConnectionString),
		[]*config.CollectorConfig{
			&collectorConfig,
		},
		prometheus.Labels{},
		&config.GlobalConfig{
			ScrapeTimeout: model.Duration(c.Timeout),
			TimeoutOffset: model.Duration(500 * time.Millisecond),
			MaxConns:      c.MaxOpenConnections,
			MaxIdleConns:  c.MaxIdleConnections,
		},
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create mssql target: %w", err)
	}

	col := newTargetCollectorAdapter(t, l)

	return integrations.NewCollectorIntegration(
		c.Name(),
		integrations.WithCollectors(col),
	), nil
}
