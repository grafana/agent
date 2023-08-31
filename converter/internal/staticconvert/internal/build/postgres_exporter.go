package build

import (
	"fmt"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/prometheus/exporter/postgres"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/converter/internal/prometheusconvert"
	"github.com/grafana/agent/pkg/integrations/postgres_exporter"
	"github.com/grafana/river/rivertypes"
)

func (b *IntegrationsV1ConfigBuilder) appendPostgresExporter(config *postgres_exporter.Config) discovery.Exports {
	args := toPostgresExporter(config)
	compLabel := common.LabelForParts(b.globalCtx.LabelPrefix, config.Name())
	b.f.Body().AppendBlock(common.NewBlockWithOverride(
		[]string{"prometheus", "exporter", "postgres"},
		compLabel,
		args,
	))

	return prometheusconvert.NewDiscoveryExports(fmt.Sprintf("prometheus.exporter.postgres.%s.targets", compLabel))
}

func toPostgresExporter(config *postgres_exporter.Config) *postgres.Arguments {
	dataSourceNames := make([]rivertypes.Secret, 0)
	for _, dsn := range config.DataSourceNames {
		dataSourceNames = append(dataSourceNames, rivertypes.Secret(dsn))
	}

	return &postgres.Arguments{
		DataSourceNames:         dataSourceNames,
		DisableSettingsMetrics:  config.DisableSettingsMetrics,
		DisableDefaultMetrics:   config.DisableDefaultMetrics,
		CustomQueriesConfigPath: config.QueryPath,
		AutoDiscovery: postgres.AutoDiscovery{
			Enabled:           config.AutodiscoverDatabases,
			DatabaseAllowlist: config.IncludeDatabases,
			DatabaseDenylist:  config.ExcludeDatabases,
		},
	}
}
