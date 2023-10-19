package build

import (
	"fmt"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/prometheus/exporter/agent"
	"github.com/grafana/agent/converter/internal/common"
	agent_exporter "github.com/grafana/agent/pkg/integrations/agent"
)

func (b *IntegrationsV1ConfigBuilder) appendAgentExporter(config *agent_exporter.Config) discovery.Exports {
	args := toAgentExporter(config)
	compLabel := common.LabelForParts(b.globalCtx.LabelPrefix, config.Name())
	b.f.Body().AppendBlock(common.NewBlockWithOverride(
		[]string{"prometheus", "exporter", "agent"},
		compLabel,
		args,
	))

	return common.NewDiscoveryExports(fmt.Sprintf("prometheus.exporter.agent.%s.targets", compLabel))
}

func toAgentExporter(config *agent_exporter.Config) *agent.Arguments {
	return &agent.Arguments{}
}
