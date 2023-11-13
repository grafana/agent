package build

import (
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/prometheus/exporter/gcp"
	"github.com/grafana/agent/pkg/integrations/gcp_exporter"
)

func (b *IntegrationsConfigBuilder) appendGcpExporter(config *gcp_exporter.Config, instanceKey *string) discovery.Exports {
	args := toGcpExporter(config)
	return b.appendExporterBlock(args, config.Name(), instanceKey, "gcp")
}

func toGcpExporter(config *gcp_exporter.Config) *gcp.Arguments {
	return &gcp.Arguments{
		ProjectIDs:            config.ProjectIDs,
		MetricPrefixes:        config.MetricPrefixes,
		ExtraFilters:          config.ExtraFilters,
		RequestInterval:       config.RequestInterval,
		RequestOffset:         config.RequestOffset,
		IngestDelay:           config.IngestDelay,
		DropDelegatedProjects: config.DropDelegatedProjects,
		ClientTimeout:         config.ClientTimeout,
	}
}
