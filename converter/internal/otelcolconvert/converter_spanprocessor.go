package otelcolconvert

import (
	"fmt"

	"github.com/grafana/agent/component/otelcol"
	"github.com/grafana/agent/component/otelcol/processor/span"
	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/spanprocessor"
	"go.opentelemetry.io/collector/component"
)

func init() {
	converters = append(converters, spanProcessorConverter{})
}

type spanProcessorConverter struct{}

func (spanProcessorConverter) Factory() component.Factory { return spanprocessor.NewFactory() }

func (spanProcessorConverter) InputComponentName() string { return "otelcol.processor.span" }

func (spanProcessorConverter) ConvertAndAppend(state *state, id component.InstanceID, cfg component.Config) diag.Diagnostics {
	var diags diag.Diagnostics

	label := state.FlowComponentLabel()

	args := toSpanProcessor(state, id, cfg.(*spanprocessor.Config))
	block := common.NewBlockWithOverride([]string{"otelcol", "processor", "span"}, label, args)

	diags.Add(
		diag.SeverityLevelInfo,
		fmt.Sprintf("Converted %s into %s", stringifyInstanceID(id), stringifyBlock(block)),
	)

	state.Body().AppendBlock(block)
	return diags
}

func toSpanProcessor(state *state, id component.InstanceID, cfg *spanprocessor.Config) *span.Arguments {
	var (
		nextTraces = state.Next(id, component.DataTypeTraces)
	)

	var setStatus *span.Status
	if cfg.SetStatus != nil {
		setStatus := &span.Status{}
		setStatus.Code = cfg.SetStatus.Code
		setStatus.Description = cfg.SetStatus.Description
	}

	var toAttributes *span.ToAttributes
	if cfg.Rename.ToAttributes != nil {
		toAttributes := &span.ToAttributes{}
		toAttributes.Rules = cfg.Rename.ToAttributes.Rules
		toAttributes.BreakAfterMatch = cfg.Rename.ToAttributes.BreakAfterMatch
	}

	return &span.Arguments{
		Match: otelcol.MatchConfig{
			Include: toMatchProperties(encodeMapstruct(cfg.Include)),
			Exclude: toMatchProperties(encodeMapstruct(cfg.Exclude)),
		},
		Name: span.Name{
			FromAttributes: cfg.Rename.FromAttributes,
			Separator:      cfg.Rename.Separator,
			ToAttributes:   toAttributes,
		},
		SetStatus: setStatus,
		Output: &otelcol.ConsumerArguments{
			Traces: toTokenizedConsumers(nextTraces),
		},
	}
}

func toMatchProperties(cfg map[string]any) *otelcol.MatchProperties {
	if cfg == nil {
		return nil
	}

	var regexpConfig *otelcol.RegexpConfig
	if cfg["regexp_config"] != nil {
		regexpConfig.CacheEnabled = cfg["regexp_config"].(map[string]any)["cache_enabled"].(bool)
		regexpConfig.CacheMaxNumEntries = cfg["regexp_config"].(map[string]any)["cache_max_num_entries"].(int)
	}

	var ls *otelcol.LogSeverityNumberMatchProperties
	if cfg["log_severity"] != nil {
		ls.Min = otelcol.SeverityLevel(cfg["log_severity"].(map[string]any)["min"].(string))
		ls.MatchUndefined = cfg["log_severity"].(map[string]any)["match_undefined"].(bool)
	}
	a := cfg["attributes"]
	attributes := toOtelcolAttributes(encodeMapslice(a))
	r := cfg["resources"]
	resources := toOtelcolAttributes(encodeMapslice(r))
	l := cfg["libraries"]
	libraries := toOtelcolInstrumentationLibrary(encodeMapslice(l))

	return &otelcol.MatchProperties{
		MatchType:        encodeString(cfg["match_type"]),
		RegexpConfig:     regexpConfig,
		Services:         cfg["services"].([]string),
		SpanNames:        cfg["span_names"].([]string),
		LogBodies:        cfg["log_bodies"].([]string),
		LogSeverityTexts: cfg["log_severity_texts"].([]string),
		LogSeverity:      ls,
		MetricNames:      cfg["metric_names"].([]string),
		Attributes:       attributes,
		Resources:        resources,
		Libraries:        libraries,
		SpanKinds:        cfg["span_kinds"].([]string),
	}
}

func toOtelcolAttributes(in []map[string]any) []otelcol.Attribute {
	res := make([]otelcol.Attribute, 0, len(in))

	for _, a := range in {
		res = append(res, otelcol.Attribute{
			Key:   a["key"].(string),
			Value: a["value"],
		})
	}

	return res
}

func toOtelcolInstrumentationLibrary(in []map[string]any) []otelcol.InstrumentationLibrary {
	res := make([]otelcol.InstrumentationLibrary, 0, len(in))

	for _, l := range in {
		res = append(res, otelcol.InstrumentationLibrary{
			Name:    l["name"].(string),
			Version: l["version"].(*string),
		})
	}
	return res
}
