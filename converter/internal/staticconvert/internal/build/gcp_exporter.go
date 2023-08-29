package build

import (
	"fmt"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/prometheus/exporter/gcp"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/converter/internal/prometheusconvert"
	"github.com/grafana/agent/pkg/integrations/gcp_exporter"
)

func (b *IntegrationsV1ConfigBuilder) appendGcpExporter(config *gcp_exporter.Config) discovery.Exports {
	args := toGcpExporter(config)
	compLabel := common.LabelForParts(b.globalCtx.LabelPrefix, config.Name())
	b.f.Body().AppendBlock(common.NewBlockWithOverride(
		[]string{"prometheus", "exporter", "gcp"},
		compLabel,
		args,
	))

	return prometheusconvert.NewDiscoveryExports(fmt.Sprintf("prometheus.exporter.gcp.%s.targets", compLabel))
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
