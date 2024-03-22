package otelcolconvert

import (
	"fmt"

	"github.com/grafana/agent/internal/component/discovery"
	"github.com/grafana/agent/internal/component/otelcol"
	otelcol_discovery "github.com/grafana/agent/internal/component/otelcol/processor/discovery"
	"github.com/grafana/agent/internal/converter/diag"
	"github.com/grafana/agent/internal/converter/internal/common"
	"github.com/grafana/agent/internal/static/traces/promsdprocessor"
	"go.opentelemetry.io/collector/component"
)

func init() {
	// Do not append this to the converter because it is a custom processor
	// from static mode and not a real otel processor.
}

type DiscoveryProcessorConverter struct{}

func (DiscoveryProcessorConverter) Factory() component.Factory {
	return promsdprocessor.NewFactory()
}

func (DiscoveryProcessorConverter) InputComponentName() string {
	return "otelcol.processor.discovery"
}

func (DiscoveryProcessorConverter) ConvertAndAppend(state *state, id component.InstanceID, cfg component.Config) diag.Diagnostics {
	var diags diag.Diagnostics

	label := state.FlowComponentLabel()

	args := toDiscoveryProcessor(state, id, cfg.(*promsdprocessor.Config))
	block := common.NewBlockWithOverride([]string{"otelcol", "processor", "discovery"}, label, args)

	diags.Add(
		diag.SeverityLevelInfo,
		fmt.Sprintf("Converted %s into %s", stringifyInstanceID(id), stringifyBlock(block)),
	)

	state.Body().AppendBlock(block)
	return diags
}

func toDiscoveryProcessor(state *state, id component.InstanceID, cfg *promsdprocessor.Config) *otelcol_discovery.Arguments {
	var (
		nextMetrics = state.Next(id, component.DataTypeMetrics)
		nextLogs    = state.Next(id, component.DataTypeLogs)
		nextTraces  = state.Next(id, component.DataTypeTraces)
	)

	// Since the scrapeConfigs are of type []interface{}, we are able overload the field
	// with []discovery.Target for the purposes of the conversion.
	targets := make([]discovery.Target, len(cfg.ScrapeConfigs))
	for _, scrapeConfig := range cfg.ScrapeConfigs {
		targets = append(targets, scrapeConfig.(discovery.Target))
	}

	return &otelcol_discovery.Arguments{
		Targets:         targets,
		OperationType:   cfg.OperationType,
		PodAssociations: cfg.PodAssociations,
		Output: &otelcol.ConsumerArguments{
			Metrics: toTokenizedConsumers(nextMetrics),
			Logs:    toTokenizedConsumers(nextLogs),
			Traces:  toTokenizedConsumers(nextTraces),
		},
	}
}
