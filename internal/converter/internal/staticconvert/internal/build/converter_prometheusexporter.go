package build

import (
	"fmt"

	"github.com/grafana/agent/internal/component/otelcol/exporter/prometheus"
	"github.com/grafana/agent/internal/converter/diag"
	"github.com/grafana/agent/internal/converter/internal/common"
	"github.com/grafana/agent/internal/converter/internal/otelcolconvert"
	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/prometheusexporter"
	"go.opentelemetry.io/collector/component"
)

func init() {
	converters = append(converters, prometheusExporterConverter{})
}

type prometheusExporterConverter struct{}

func (prometheusExporterConverter) Factory() component.Factory {
	return prometheusexporter.NewFactory()
}

func (prometheusExporterConverter) InputComponentName() string {
	return "otelcol.exporter.prometheus"
}

func (prometheusExporterConverter) ConvertAndAppend(state *otelcolconvert.State, id component.InstanceID, cfg component.Config) diag.Diagnostics {
	label := state.FlowComponentLabel()

	args := toPrometheusExporterConfig(state, id, cfg.(*prometheusexporter.Config), label)
	block := common.NewBlockWithOverride([]string{"otelcol", "exporter", "prometheus"}, label, args)

	var diags diag.Diagnostics
	diags.Add(
		diag.SeverityLevelInfo,
		fmt.Sprintf("Converted %s into %s", otelcolconvert.StringifyInstanceID(id), otelcolconvert.StringifyBlock(block)),
	)

	state.Body().AppendBlock(block)
	return diags
}

func toPrometheusExporterConfig(state *otelcolconvert.State, id component.InstanceID, cfg *prometheusexporter.Config, label string) *prometheus.Arguments {
	args := &prometheus.Arguments{}
	args.SetToDefault()
	args.GCFrequency = cfg.MetricExpiration
	args.AddMetricSuffixes = cfg.AddMetricSuffixes
	args.ResourceToTelemetryConversion = cfg.ResourceToTelemetrySettings.Enabled
	// TODO args.ForwardTo
	// IncludeTargetInfo:  false,
	// IncludeScopeInfo:   false,
	// IncludeScopeLabels: false,

	return args
}
