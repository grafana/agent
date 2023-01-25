package mssql_exporter

import (
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

// Config is the configuration for the mssql v2 integration
type Config struct {
	Config mssql_common.Config  `yaml:",inline"`
	Common common.MetricsConfig `yaml:",inline"`
}

func (c Config) validate() error {
	return c.Config.Validate()
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

	url, err := url.Parse(c.Config.ConnectionString)
	if err != nil {
		return "", fmt.Errorf("failed to parse connection string URL: %w", err)
	}

	return url.Host, nil
}

// UnmarshalYAML implements yaml.Unmarshaler for Config
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	c.Config = mssql_common.DefaultConfig

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
		c.Config.ConnectionString,
		[]*config.CollectorConfig{
			&mssql_common.CollectorConfig,
		},
		prometheus.Labels{},
		&config.GlobalConfig{
			ScrapeTimeout: model.Duration(c.Config.Timeout),
			TimeoutOffset: model.Duration(500 * time.Millisecond),
			MaxConns:      c.Config.MaxOpenConnections,
			MaxIdleConns:  c.Config.MaxIdleConnections,
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
