package snowflake_exporter

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/burningalchemist/sql_exporter"
	"github.com/burningalchemist/sql_exporter/config"
	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/model"

	mssql_common "github.com/grafana/agent/pkg/integrations/mssql/common"
	integrations_v2 "github.com/grafana/agent/pkg/integrations/v2"
	"github.com/grafana/agent/pkg/integrations/v2/common"
	"github.com/grafana/agent/pkg/integrations/v2/metricsutils"
)

// DefaultConfig is the default config for the mssql v2 integration
var DefaultConfig = Config{
	MaxIdleConnections: 3,
	MaxConnections:     3,
	Timeout:            10 * time.Second,
}

// Config is the configuration for the mssql v2 integration
type Config struct {
	ConnectionString   string               `yaml:"connection_string,omitempty"`
	MaxIdleConnections int                  `yaml:"max_idle_connections,omitempty"`
	MaxConnections     int                  `yaml:"max_connections,omitempty"`
	Timeout            time.Duration        `yaml:"timeout,omitempty"`
	Common             common.MetricsConfig `yaml:",inline"`
}

func (c Config) validate() error {
	if c.ConnectionString == "" {
		return errors.New("the connection_string parameter is required")
	}

	url, err := url.Parse(c.ConnectionString)
	if err != nil {
		return fmt.Errorf("failed to parse connection_string: %w", err)
	}

	if url.Scheme != "sqlserver" {
		return errors.New("scheme of provided connection_string URL must be sqlserver")
	}

	if c.MaxConnections < 1 {
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

	url, err := url.Parse(c.ConnectionString)
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
	integrations_v2.Register(&Config{}, integrations_v2.TypeMultiplex)
}

// NewIntegration creates a new v2 integration from the config.
func (c *Config) NewIntegration(l log.Logger, globals integrations_v2.Globals) (integrations_v2.Integration, error) {
	err := c.validate()
	if err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}

	handler, err := createHandler(l, c)
	if err != nil {
		return nil, fmt.Errorf("failed to create http handler: %w", err)
	}

	return metricsutils.NewMetricsHandlerIntegration(l, c, c.Common, globals, handler)
}

func createHandler(logger log.Logger, c *Config) (http.HandlerFunc, error) {
	t, err := sql_exporter.NewTarget(
		"mssqlintegration",
		"",
		c.ConnectionString,
		[]*config.CollectorConfig{
			&mssql_common.CollectorConfig,
		},
		prometheus.Labels{},
		// TODO: Evaluate if these need to be configurable
		&config.GlobalConfig{
			ScrapeTimeout: model.Duration(c.Timeout),
			TimeoutOffset: model.Duration(500 * time.Millisecond),
			MaxConns:      c.MaxConnections,
			MaxIdleConns:  c.MaxIdleConnections,
		},
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create mssql target: %w", err)
	}

	col := mssql_common.NewTargetCollectorAdapter(t, logger)

	registry := prometheus.NewRegistry()
	registry.MustRegister(col)
	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})

	return h.ServeHTTP, nil
}
