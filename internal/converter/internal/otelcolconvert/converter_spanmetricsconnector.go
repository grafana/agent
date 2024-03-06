package otelcolconvert

import (
	"fmt"
	"time"

	"github.com/grafana/agent/internal/component/otelcol"
	"github.com/grafana/agent/internal/component/otelcol/connector/spanmetrics"
	"github.com/grafana/agent/internal/converter/diag"
	"github.com/grafana/agent/internal/converter/internal/common"
	"github.com/open-telemetry/opentelemetry-collector-contrib/connector/spanmetricsconnector"
	"go.opentelemetry.io/collector/component"
)

func init() {
	converters = append(converters, spanmetricsConnectorConverter{})
}

type spanmetricsConnectorConverter struct{}

func (spanmetricsConnectorConverter) Factory() component.Factory {
	return spanmetricsconnector.NewFactory()
}

func (spanmetricsConnectorConverter) InputComponentName() string {
	return "otelcol.connector.spanmetrics"
}

func (spanmetricsConnectorConverter) ConvertAndAppend(state *state, id component.InstanceID, cfg component.Config) diag.Diagnostics {
	var diags diag.Diagnostics

	label := state.FlowComponentLabel()

	args := toSpanmetricsConnector(state, id, cfg.(*spanmetricsconnector.Config))
	block := common.NewBlockWithOverride([]string{"otelcol", "connector", "spanmetrics"}, label, args)

	diags.Add(
		diag.SeverityLevelInfo,
		fmt.Sprintf("Converted %s into %s", stringifyInstanceID(id), stringifyBlock(block)),
	)

	state.Body().AppendBlock(block)
	return diags
}

func toSpanmetricsConnector(state *state, id component.InstanceID, cfg *spanmetricsconnector.Config) *spanmetrics.Arguments {
	if cfg == nil {
		return nil
	}
	var (
		nextMetrics = state.Next(id, component.DataTypeMetrics)
	)

	var exponential *spanmetrics.ExponentialHistogramConfig
	if cfg.Histogram.Exponential != nil {
		exponential = &spanmetrics.ExponentialHistogramConfig{
			MaxSize: cfg.Histogram.Exponential.MaxSize,
		}
	}

	var explicit *spanmetrics.ExplicitHistogramConfig
	if cfg.Histogram.Explicit != nil {
		explicit = &spanmetrics.ExplicitHistogramConfig{
			Buckets: cfg.Histogram.Explicit.Buckets,
		}
	}

	// If none have been explicitly set, assign the upstream default.
	if exponential == nil && explicit == nil {
		explicit = &spanmetrics.ExplicitHistogramConfig{Buckets: []time.Duration{}}
		explicit.SetToDefault()
	}

	var dimensions []spanmetrics.Dimension
	for _, d := range cfg.Dimensions {
		dimensions = append(dimensions, spanmetrics.Dimension{
			Name:    d.Name,
			Default: d.Default,
		})
	}

	return &spanmetrics.Arguments{
		Dimensions:             dimensions,
		ExcludeDimensions:      cfg.ExcludeDimensions,
		DimensionsCacheSize:    cfg.DimensionsCacheSize,
		AggregationTemporality: spanmetrics.FromOTelAggregationTemporality(cfg.AggregationTemporality),
		Histogram: spanmetrics.HistogramConfig{
			Disable:     cfg.Histogram.Disable,
			Unit:        cfg.Histogram.Unit.String(),
			Exponential: exponential,
			Explicit:    explicit,
		},
		MetricsFlushInterval: cfg.MetricsFlushInterval,
		Namespace:            cfg.Namespace,
		Exemplars: spanmetrics.ExemplarsConfig{
			Enabled: cfg.Exemplars.Enabled,
		},

		Output: &otelcol.ConsumerArguments{
			Metrics: toTokenizedConsumers(nextMetrics),
		},
	}
}
