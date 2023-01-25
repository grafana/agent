package mssql

import (
	"fmt"
	"net/url"
	"time"

	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/burningalchemist/sql_exporter"
	"github.com/burningalchemist/sql_exporter/config"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/integrations/mssql/common"
	"github.com/prometheus/common/model"
)

type Config struct {
	common.Config `yaml:",inline"`
}

// Identifier returns a string that identifies the integration.
func (c *Config) InstanceKey(agentKey string) (string, error) {
	url, err := url.Parse(c.ConnectionString)
	if err != nil {
		return "", fmt.Errorf("failed to parse connection string URL: %w", err)
	}

	return url.Host, nil
}

// UnmarshalYAML implements yaml.Unmarshaler for Config
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	c.Config = common.DefaultConfig

	type plain Config
	return unmarshal((*plain)(c))
}

// Name returns the name of the integration this config is for.
func (c *Config) Name() string {
	return "mssql"
}

func init() {
	integrations.RegisterIntegration(&Config{})
}

// NewIntegration creates a new integration from the config.
func (c *Config) NewIntegration(l log.Logger) (integrations.Integration, error) {
	if err := c.Config.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}

	t, err := sql_exporter.NewTarget(
		"mssqlintegration",
		"",
		c.ConnectionString,
		[]*config.CollectorConfig{
			&common.CollectorConfig,
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

	col := common.NewCollector(t, l)

	return integrations.NewCollectorIntegration(
		c.Name(),
		integrations.WithCollectors(col),
	), nil
}
