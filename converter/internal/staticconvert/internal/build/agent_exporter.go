package build

import (
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/prometheus/exporter/agent"
	agent_exporter "github.com/grafana/agent/pkg/integrations/agent"
	agent_exporter_v2 "github.com/grafana/agent/pkg/integrations/v2/agent"
)

func (b *IntegrationsConfigBuilder) appendAgentExporter(config *agent_exporter.Config) discovery.Exports {
	args := toAgentExporter(config)
	return b.appendExporterBlock(args, config.Name(), nil, "agent")
}

func toAgentExporter(config *agent_exporter.Config) *agent.Arguments {
	return &agent.Arguments{}
}

func (b *IntegrationsConfigBuilder) appendAgentExporterV2(config *agent_exporter_v2.Config) discovery.Exports {
	args := toAgentExporterV2(config)
	return b.appendExporterBlock(args, config.Name(), config.Common.InstanceKey, "agent")
}

func toAgentExporterV2(config *agent_exporter_v2.Config) *agent.Arguments {
	return &agent.Arguments{}
}
