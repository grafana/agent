package build

import (
	"fmt"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/prometheus/exporter/azure"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/converter/internal/prometheusconvert"
	"github.com/grafana/agent/pkg/integrations/azure_exporter"
)

func (b *IntegrationsV1ConfigBuilder) appendAzureExporter(config *azure_exporter.Config) discovery.Exports {
	args := toAzureExporter(config)
	compLabel := common.LabelForParts(b.globalCtx.LabelPrefix, config.Name())
	b.f.Body().AppendBlock(common.NewBlockWithOverride(
		[]string{"prometheus", "exporter", "azure"},
		compLabel,
		args,
	))

	return prometheusconvert.NewDiscoveryExports(fmt.Sprintf("prometheus.exporter.azure.%s.targets", compLabel))
}

func toAzureExporter(config *azure_exporter.Config) *azure.Arguments {
	return &azure.Arguments{
		Subscriptions:            config.Subscriptions,
		ResourceGraphQueryFilter: config.ResourceGraphQueryFilter,
		ResourceType:             config.ResourceType,
		Metrics:                  config.Metrics,
		MetricAggregations:       config.MetricAggregations,
		Timespan:                 config.Timespan,
		IncludedDimensions:       config.IncludedDimensions,
		IncludedResourceTags:     config.IncludedResourceTags,
		MetricNamespace:          config.MetricNamespace,
		MetricNameTemplate:       config.MetricNameTemplate,
		MetricHelpTemplate:       config.MetricHelpTemplate,
		AzureCloudEnvironment:    config.AzureCloudEnvironment,
	}
}
