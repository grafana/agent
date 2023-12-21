package build

import (
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/prometheus/exporter/azure"
	"github.com/grafana/agent/pkg/integrations/azure_exporter"
)

func (b *IntegrationsConfigBuilder) appendAzureExporter(config *azure_exporter.Config, instanceKey *string) discovery.Exports {
	args := toAzureExporter(config)
	return b.appendExporterBlock(args, config.Name(), instanceKey, "azure")
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
		ValidateDimensions:       config.ValidateDimensions,
		Regions:                  config.Regions,
	}
}
