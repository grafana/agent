// Package postgres_exporter embeds https://github.com/prometheus/postgres_exporter
package postgres_exporter //nolint:golint

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/integrations/config"
	"github.com/wrouesnel/postgres_exporter/exporter"
)

var DefaultConfig = Config{}

// Config controls the postgres_exporter integration.
type Config struct {
	Common config.Common `yaml:",inline"`

	// DataSourceNames to use to connect to Postgres.
	DataSourceNames []string `yaml:"data_source_names"`

	DisableSettingsMetrics bool     `yaml:"disable_settings_metrics"`
	AutodiscoverDatabases  bool     `yaml:"autodiscover_databases"`
	ExcludeDatabases       []string `yaml:"exclude_databases"`
}

// UnmarshalYAML implements yaml.Unmarshaler for Config.
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultConfig

	type plain Config
	return unmarshal((*plain)(c))
}

func (c *Config) Name() string {
	return "postgres_exporter"
}

func (c *Config) CommonConfig() config.Common {
	return c.Common
}

func (c *Config) NewIntegration(l log.Logger) (integrations.Integration, error) {
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
		exporter.DisableDefaultMetrics(false),
		exporter.DisableSettingsMetrics(c.DisableSettingsMetrics),
		exporter.AutoDiscoverDatabases(c.AutodiscoverDatabases),
		exporter.ExcludeDatabases(strings.Join(c.ExcludeDatabases, ",")),
	)

	return integrations.NewCollectorIntegration(c.Name(), e, false), nil
}
