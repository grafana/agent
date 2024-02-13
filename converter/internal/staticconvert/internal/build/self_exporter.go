package build

import (
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/prometheus/exporter/self"
	agent_exporter "github.com/grafana/agent/pkg/integrations/agent"
	agent_exporter_v2 "github.com/grafana/agent/pkg/integrations/v2/agent"
)

func (b *IntegrationsConfigBuilder) appendAgentExporter(config *agent_exporter.Config) discovery.Exports {
	args := toAgentExporter(config)
	return b.appendExporterBlock(args, config.Name(), nil, "self")
}

func toAgentExporter(config *agent_exporter.Config) *self.Arguments {
	return &self.Arguments{}
}

func (b *IntegrationsConfigBuilder) appendAgentExporterV2(config *agent_exporter_v2.Config) discovery.Exports {
	args := toAgentExporterV2(config)
	return b.appendExporterBlock(args, config.Name(), config.Common.InstanceKey, "self")
}

func toAgentExporterV2(config *agent_exporter_v2.Config) *self.Arguments {
	return &self.Arguments{}
}
