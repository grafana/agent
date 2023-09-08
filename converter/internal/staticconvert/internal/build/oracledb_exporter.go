package build

import (
	"fmt"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/prometheus/exporter/oracledb"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/converter/internal/prometheusconvert"
	"github.com/grafana/agent/pkg/integrations/oracledb_exporter"
	"github.com/grafana/river/rivertypes"
)

func (b *IntegrationsV1ConfigBuilder) appendOracledbExporter(config *oracledb_exporter.Config) discovery.Exports {
	args := toOracledbExporter(config)
	compLabel := common.LabelForParts(b.globalCtx.LabelPrefix, config.Name())
	b.f.Body().AppendBlock(common.NewBlockWithOverride(
		[]string{"prometheus", "exporter", "oracledb"},
		compLabel,
		args,
	))

	return prometheusconvert.NewDiscoveryExports(fmt.Sprintf("prometheus.exporter.oracledb.%s.targets", compLabel))
}

func toOracledbExporter(config *oracledb_exporter.Config) *oracledb.Arguments {
	return &oracledb.Arguments{
		ConnectionString: rivertypes.Secret(config.ConnectionString),
		MaxIdleConns:     config.MaxIdleConns,
		MaxOpenConns:     config.MaxOpenConns,
		QueryTimeout:     config.QueryTimeout,
	}
}
