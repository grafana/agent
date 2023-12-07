package build

import (
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/prometheus/exporter/mssql"
	mssql_exporter "github.com/grafana/agent/pkg/integrations/mssql"
	"github.com/grafana/river/rivertypes"
)

func (b *IntegrationsConfigBuilder) appendMssqlExporter(config *mssql_exporter.Config, instanceKey *string) discovery.Exports {
	args := toMssqlExporter(config)
	return b.appendExporterBlock(args, config.Name(), instanceKey, "mssql")
}

func toMssqlExporter(config *mssql_exporter.Config) *mssql.Arguments {
	return &mssql.Arguments{
		ConnectionString:   rivertypes.Secret(config.ConnectionString),
		MaxIdleConnections: config.MaxIdleConnections,
		MaxOpenConnections: config.MaxOpenConnections,
		Timeout:            config.Timeout,
	}
}
