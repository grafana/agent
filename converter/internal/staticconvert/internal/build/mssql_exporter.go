package build

import (
	"fmt"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/prometheus/exporter/mssql"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/converter/internal/prometheusconvert"
	mssql_exporter "github.com/grafana/agent/pkg/integrations/mssql"
	"github.com/grafana/river/rivertypes"
)

func (b *IntegrationsV1ConfigBuilder) appendMssqlExporter(config *mssql_exporter.Config) discovery.Exports {
	args := toMssqlExporter(config)
	compLabel := common.LabelForParts(b.globalCtx.LabelPrefix, config.Name())
	b.f.Body().AppendBlock(common.NewBlockWithOverride(
		[]string{"prometheus", "exporter", "mssql"},
		compLabel,
		args,
	))

	return prometheusconvert.NewDiscoveryExports(fmt.Sprintf("prometheus.exporter.mssql.%s.targets", compLabel))
}

func toMssqlExporter(config *mssql_exporter.Config) *mssql.Arguments {
	return &mssql.Arguments{
		ConnectionString:   rivertypes.Secret(config.ConnectionString),
		MaxIdleConnections: config.MaxIdleConnections,
		MaxOpenConnections: config.MaxOpenConnections,
		Timeout:            config.Timeout,
	}
}
