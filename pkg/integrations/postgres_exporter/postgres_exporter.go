// Package postgres_exporter embeds https://github.com/prometheus/postgres_exporter
package postgres_exporter //nolint:golint

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/integrations/config"
	loki "github.com/grafana/agent/pkg/logs"
	"github.com/grafana/agent/pkg/tempo"
	"github.com/prometheus-community/postgres_exporter/exporter"
)

// Config controls the postgres_exporter integration.
type Config struct {
	Common config.Common `yaml:",inline"`

	// DataSourceNames to use to connect to Postgres.
	DataSourceNames []string `yaml:"data_source_names,omitempty"`

	DisableSettingsMetrics bool     `yaml:"disable_settings_metrics,omitempty"`
	AutodiscoverDatabases  bool     `yaml:"autodiscover_databases,omitempty"`
	ExcludeDatabases       []string `yaml:"exclude_databases,omitempty"`
	IncludeDatabases       []string `yaml:"include_databases,omitempty"`
	DisableDefaultMetrics  bool     `yaml:"disable_default_metrics,omitempty"`
	QueryPath              string   `yaml:"query_path,omitempty"`
}

// Name returns the name of the integration this config is for.
func (c *Config) Name() string {
	return "postgres_exporter"
}

// CommonConfig returns the common set of options shared across all configs for
// integrations.
func (c *Config) CommonConfig() config.Common {
	return c.Common
}

// NewIntegration converts this config into an instance of a configuration.
func (c *Config) NewIntegration(l log.Logger, loki *loki.Logs, tempo *tempo.Tempo) (integrations.Integration, error) {
	return New(l, c)
}

func init() {
	integrations.RegisterIntegration(&Config{})
}

// New creates a new postgres_exporter integration. The integration scrapes
// metrics from a postgres process.
func New(log log.Logger, c *Config) (integrations.Integration, error) {
	dsn := c.DataSourceNames
	if len(dsn) == 0 {
		dsn = strings.Split(os.Getenv("POSTGRES_EXPORTER_DATA_SOURCE_NAME"), ",")
	}
	if len(dsn) == 0 {
		return nil, fmt.Errorf("cannot create postgres_exporter; neither postgres_exporter.data_source_name or $POSTGRES_EXPORTER_DATA_SOURCE_NAME is set")
	}

	e := exporter.NewExporter(
		dsn,
		log,
		exporter.DisableDefaultMetrics(c.DisableDefaultMetrics),
		exporter.WithUserQueriesPath(c.QueryPath),
		exporter.DisableSettingsMetrics(c.DisableSettingsMetrics),
		exporter.AutoDiscoverDatabases(c.AutodiscoverDatabases),
		exporter.ExcludeDatabases(strings.Join(c.ExcludeDatabases, ",")),
		exporter.IncludeDatabases(strings.Join(c.IncludeDatabases, ",")),
		exporter.MetricPrefix("pg"),
	)

	return integrations.NewCollectorIntegration(c.Name(), integrations.WithCollectors(e)), nil
}
