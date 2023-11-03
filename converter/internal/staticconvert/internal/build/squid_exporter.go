package build

import (
	"fmt"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/prometheus/exporter/squid"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/pkg/integrations/squid_exporter"
	"github.com/grafana/river/rivertypes"
)

func (b *IntegrationsConfigBuilder) appendSquidExporter(config *squid_exporter.Config) discovery.Exports {
	args := toSquidExporter(config)
	compLabel := common.LabelForParts(b.globalCtx.LabelPrefix, config.Name())
	b.f.Body().AppendBlock(common.NewBlockWithOverride(
		[]string{"prometheus", "exporter", "squid"},
		compLabel,
		args,
	))

	return common.NewDiscoveryExports(fmt.Sprintf("prometheus.exporter.squid.%s.targets", compLabel))
}

func toSquidExporter(config *squid_exporter.Config) *squid.Arguments {
	return &squid.Arguments{
		SquidAddr:     config.Address,
		SquidUser:     config.Username,
		SquidPassword: rivertypes.Secret(config.Password),
	}
}
