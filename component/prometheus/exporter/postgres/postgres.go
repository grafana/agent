package postgres

import (
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/prometheus/exporter"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/integrations/postgres_exporter"
	"github.com/grafana/agent/pkg/river/rivertypes"
	config_util "github.com/prometheus/common/config"
)

func init() {
	component.Register(component.Registration{
		Name:    "prometheus.exporter.postgres",
		Args:    Arguments{},
		Exports: exporter.Exports{},
		Build:   exporter.New(createExporter, "postgres"),
	})
}

func createExporter(opts component.Options, args component.Arguments) (integrations.Integration, error) {
	a := args.(Arguments)
	return a.Convert().NewIntegration(opts.Logger)
}

// DefaultArguments holds the default arguments for the prometheus.exporter.postgres
// component.
var DefaultArguments = Arguments{
	DisableSettingsMetrics: false,
	AutoDiscovery: AutoDiscovery{
		Enabled: false,
	},
	DisableDefaultMetrics:   false,
	CustomQueriesConfigPath: "",
}

// Arguments configures the prometheus.exporter.postgres component
type Arguments struct {
	// DataSourceNames to use to connect to Postgres. This is marked optional because it
	// may also be supplied by the POSTGRES_EXPORTER_DATA_SOURCE_NAME env var,
	// though it is not recommended to do so.
	DataSourceNames []rivertypes.Secret `river:"data_source_names,attr,optional"`

	// Attributes
	DisableSettingsMetrics  bool   `river:"disable_settings_metrics,attr,optional"`
	DisableDefaultMetrics   bool   `river:"disable_default_metrics,attr,optional"`
	CustomQueriesConfigPath string `river:"custom_queries_config_path,attr,optional"`

	// Blocks
	AutoDiscovery AutoDiscovery `river:"autodiscovery,block,optional"`
}

// Autodiscovery controls discovery of databases outside any specified in DataSourceNames.
type AutoDiscovery struct {
	Enabled           bool     `river:"enabled,attr,optional"`
	DatabaseAllowlist []string `river:"database_allowlist,attr,optional"`
	DatabaseDenylist  []string `river:"database_denylist,attr,optional"`
}

// UnmarshalRiver implements River unmarshalling for Arguments.
func (a *Arguments) UnmarshalRiver(f func(interface{}) error) error {
	*a = DefaultArguments

	type args Arguments
	return f((*args)(a))
}

func (a *Arguments) Convert() *postgres_exporter.Config {
	return &postgres_exporter.Config{
		DataSourceNames:        a.convertDataSourceNames(),
		DisableSettingsMetrics: a.DisableSettingsMetrics,
		AutodiscoverDatabases:  a.AutoDiscovery.Enabled,
		ExcludeDatabases:       a.AutoDiscovery.DatabaseDenylist,
		IncludeDatabases:       a.AutoDiscovery.DatabaseAllowlist,
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
