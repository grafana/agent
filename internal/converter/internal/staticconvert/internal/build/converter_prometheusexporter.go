package build

import (
	"fmt"

	flow_relabel "github.com/grafana/agent/internal/component/common/relabel"
	"github.com/grafana/agent/internal/component/otelcol/exporter/prometheus"
	"github.com/grafana/agent/internal/component/prometheus/relabel"
	"github.com/grafana/agent/internal/converter/diag"
	"github.com/grafana/agent/internal/converter/internal/common"
	"github.com/grafana/agent/internal/converter/internal/otelcolconvert"
	"github.com/grafana/agent/internal/converter/internal/prometheusconvert/build"
	prometheus_component "github.com/grafana/agent/internal/converter/internal/prometheusconvert/component"
	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/prometheusexporter"
	prom_relabel "github.com/prometheus/prometheus/model/relabel"
	"github.com/prometheus/prometheus/storage"
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
	var exports *relabel.Exports
	var forwardTo []storage.Appendable
	label := state.FlowComponentLabel()

	if len(cfg.(*prometheusexporter.Config).ConstLabels) > 0 {
		exports = includeRelabelConfig(label, cfg, state)
		forwardTo = []storage.Appendable{exports.Receiver}
	}

	args := toPrometheusExporterConfig(cfg.(*prometheusexporter.Config), forwardTo)
	block := common.NewBlockWithOverride([]string{"otelcol", "exporter", "prometheus"}, label, args)

	var diags diag.Diagnostics
	diags.Add(
		diag.SeverityLevelInfo,
		fmt.Sprintf("Converted %s into %s", otelcolconvert.StringifyInstanceID(id), otelcolconvert.StringifyBlock(block)),
	)

	state.Body().AppendBlock(block)
	return diags
}

func includeRelabelConfig(label string, cfg component.Config, state *otelcolconvert.State) *relabel.Exports {
	pb := build.NewPrometheusBlocks()

	relabelConfigs := []*prom_relabel.Config{}
	for label, replacement := range cfg.(*prometheusexporter.Config).ConstLabels {
		defaultRelabelConfigs := &flow_relabel.Config{}
		defaultRelabelConfigs.SetToDefault()
		relabelConfigs = append(relabelConfigs, &prom_relabel.Config{
			Separator:   defaultRelabelConfigs.Separator,
			Regex:       prom_relabel.Regexp(defaultRelabelConfigs.Regex),
			Modulus:     defaultRelabelConfigs.Modulus,
			TargetLabel: label,
			Replacement: replacement,
			Action:      prom_relabel.Action(defaultRelabelConfigs.Action),
		})
	}

	exports := prometheus_component.AppendPrometheusRelabel(pb, relabelConfigs, nil, label)
	pb.AppendToBody(state.Body())
	return exports
}

func toPrometheusExporterConfig(cfg *prometheusexporter.Config, forwardTo []storage.Appendable) *prometheus.Arguments {
	defaultArgs := &prometheus.Arguments{}
	defaultArgs.SetToDefault()

	return &prometheus.Arguments{
		IncludeTargetInfo:             defaultArgs.IncludeTargetInfo,
		IncludeScopeInfo:              defaultArgs.IncludeScopeInfo,
		IncludeScopeLabels:            defaultArgs.IncludeScopeLabels,
		GCFrequency:                   cfg.MetricExpiration,
		ForwardTo:                     forwardTo,
		AddMetricSuffixes:             cfg.AddMetricSuffixes,
		ResourceToTelemetryConversion: cfg.ResourceToTelemetrySettings.Enabled,
	}
}
