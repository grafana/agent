package postgres

import (
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/prometheus/exporter"
	"github.com/grafana/agent/pkg/flow/rivertypes"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/integrations/postgres_exporter"
	config_util "github.com/prometheus/common/config"
)

func init() {
	component.Register(component.Registration{
		Name:    "prometheus.exporter.postgres",
		Args:    Arguments{},
		Exports: exporter.Exports{},
		Build:   exporter.New(createIntegration, "postgres"),
	})
}

func createIntegration(opts component.Options, args component.Arguments) (integrations.Integration, error) {
	cfg := args.(Arguments)
	return cfg.Convert().NewIntegration(opts.Logger)
}

// Arguments configures the prometheus.exporter.postgres component
type Arguments struct {
	// DataSourceNames to use to connect to Postgres. This is marked optional because it
	// may also be supplied by the POSTGRES_EXPORTER_DATA_SOURCE_NAME env var,
	// though it is not recommended to do so.
	DataSourceNames []rivertypes.Secret `river:"data_source_names,attr,optional"`

	DisableSettingsMetrics  bool     `river:"disable_settings_metrics,attr,optional"`
	AutodiscoverDatabases   bool     `river:"autodiscover_databases,attr,optional"`
	ExcludeDatabases        []string `river:"exclude_databases,attr,optional"`
	IncludeDatabases        []string `river:"include_databases,attr,optional"`
	DisableDefaultMetrics   bool     `river:"disable_default_metrics,attr,optional"`
	CustomQueriesConfigPath string   `river:"custom_queries_config_path,attr,optional"`
}

// UnmarshalRiver implements River unmarshalling for Arguments.
func (c *Arguments) UnmarshalRiver(f func(interface{}) error) error {
	*c = Arguments{}

	type args Arguments
	return f((*args)(c))
}

func (a *Arguments) Convert() *postgres_exporter.Config {
	return &postgres_exporter.Config{
		DataSourceNames:        a.convertDataSourceNames(),
		DisableSettingsMetrics: a.DisableSettingsMetrics,
		AutodiscoverDatabases:  a.AutodiscoverDatabases,
		ExcludeDatabases:       a.ExcludeDatabases,
		IncludeDatabases:       a.IncludeDatabases,
		DisableDefaultMetrics:  a.DisableDefaultMetrics,
		QueryPath:              a.CustomQueriesConfigPath,
	}
}

func (a *Arguments) convertDataSourceNames() []config_util.Secret {
	dataSourceNames := make([]config_util.Secret, len(a.DataSourceNames))
	for i, dataSourceName := range a.DataSourceNames {
		dataSourceNames[i] = config_util.Secret(dataSourceName)
	}
	return dataSourceNames
}
