package build

import (
	"fmt"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/prometheus/exporter/agent"
	"github.com/grafana/agent/converter/internal/common"
	agent_exporter "github.com/grafana/agent/pkg/integrations/agent"
	agent_exporter_v2 "github.com/grafana/agent/pkg/integrations/v2/agent"
)

func (b *IntegrationsConfigBuilder) appendAgentExporter(config any) discovery.Exports {
	var args *agent.Arguments
	var name string
	switch cfg := config.(type) {
	case *agent_exporter.Config:
		args = toAgentExporter(cfg)
		name = cfg.Name()
	case *agent_exporter_v2.Config:
		args = toAgentExporterV2(cfg)
		name = cfg.Name()
	default:
		panic("invalid config type passed to appendAgentExporter.")
	}

	compLabel := common.LabelForParts(b.globalCtx.LabelPrefix, name)
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

func toAgentExporterV2(config *agent_exporter_v2.Config) *agent.Arguments {
	return &agent.Arguments{}
}
