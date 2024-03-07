package otelcolconvert

import (
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/grafana/agent/internal/component/otelcol"
	"github.com/grafana/agent/internal/component/otelcol/processor/tail_sampling"
	"github.com/grafana/agent/internal/converter/diag"
	"github.com/grafana/agent/internal/converter/internal/common"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/tailsamplingprocessor"
	"go.opentelemetry.io/collector/component"
)

func init() {
	converters = append(converters, tailSamplingProcessorConverter{})
}

type tailSamplingProcessorConverter struct{}

func (tailSamplingProcessorConverter) Factory() component.Factory {
	return tailsamplingprocessor.NewFactory()
}

func (tailSamplingProcessorConverter) InputComponentName() string {
	return "otelcol.processor.tail_sampling"
}

func (tailSamplingProcessorConverter) ConvertAndAppend(state *state, id component.InstanceID, cfg component.Config) diag.Diagnostics {
	var diags diag.Diagnostics

	label := state.FlowComponentLabel()

	args := toTailSamplingProcessor(state, id, cfg.(*tailsamplingprocessor.Config))
	block := common.NewBlockWithOverride([]string{"otelcol", "processor", "tail_sampling"}, label, args)

	diags.Add(
		diag.SeverityLevelInfo,
		fmt.Sprintf("Converted %s into %s", stringifyInstanceID(id), stringifyBlock(block)),
	)

	state.Body().AppendBlock(block)
	return diags
}

func toTailSamplingProcessor(state *state, id component.InstanceID, cfg *tailsamplingprocessor.Config) *tail_sampling.Arguments {
	var (
		nextTraces = state.Next(id, component.DataTypeTraces)
	)

	testEncode := encodeMapstruct(cfg.PolicyCfgs[0])
	spew.Dump(testEncode)

	return &tail_sampling.Arguments{
		PolicyCfgs:              toPolicyCfgs(cfg.PolicyCfgs),
		DecisionWait:            cfg.DecisionWait,
		NumTraces:               cfg.NumTraces,
		ExpectedNewTracesPerSec: cfg.ExpectedNewTracesPerSec,
		Output: &otelcol.ConsumerArguments{
			Traces: toTokenizedConsumers(nextTraces),
		},
	}
}

func toPolicyCfgs(cfgs []tailsamplingprocessor.PolicyCfg) []tail_sampling.PolicyConfig {
	var out []tail_sampling.PolicyConfig
	for _, cfg := range cfgs {
		out = append(out, tail_sampling.PolicyConfig{
			SharedPolicyConfig: toSharedPolicyConfig(cfg),
			CompositeConfig:    toCompositeConfig(cfg.CompositeCfg),
			AndConfig:          toAndConfig(cfg.AndCfg),
		})
	}
	return out
}

func toSharedPolicyConfig(cfg tailsamplingprocessor.PolicyCfg) tail_sampling.SharedPolicyConfig {
	return tail_sampling.SharedPolicyConfig{
		Name:                   cfg.Name,
		Type:                   string(cfg.Type),
		LatencyConfig:          toLatencyConfig(cfg.LatencyCfg),
		NumericAttributeConfig: toNumericAttributeConfig(cfg.NumericAttributeCfg),
		ProbabilisticConfig:    toProbabilisticConfig(cfg.ProbabilisticCfg),
		StatusCodeConfig:       toStatusCodeConfig(cfg.StatusCodeCfg),
		StringAttributeConfig:  toStringAttributeConfig(cfg.StringAttributeCfg),
		RateLimitingConfig:     toRateLimitingConfig(cfg.RateLimitingCfg),
		SpanCountConfig:        toSpanCountConfig(cfg.SpanCountCfg),
		BooleanAttributeConfig: toBooleanAttributeConfig(cfg.BooleanAttributeCfg),
		OttlConditionConfig:    toOttlConditionConfig(cfg.OTTLConditionCfg),
		TraceStateConfig:       toTraceStateConfig(cfg.TraceStateCfg),
	}
}

func toCompositeConfig(cfg tailsamplingprocessor.CompositeCfg) tail_sampling.CompositeConfig {
	return tail_sampling.CompositeConfig{
		MaxTotalSpansPerSecond: cfg.MaxTotalSpansPerSecond,
		PolicyOrder:            cfg.PolicyOrder,
		SubPolicyCfg:           toSubPolicyConfig(cfg.SubPolicyCfg),
		RateAllocation:         toRateAllocationConfig(cfg.RateAllocation),
	}
}

func toSubPolicyConfig(cfgs []tailsamplingprocessor.CompositeSubPolicyCfg) []tail_sampling.CompositeSubPolicyConfig {
	var out []tail_sampling.CompositeSubPolicyConfig
	for _, cfg := range cfgs {
		out = append(out, tail_sampling.CompositeSubPolicyConfig{
			AndConfig: toAndConfig(cfg.AndCfg),
			SharedPolicyConfig: tail_sampling.SharedPolicyConfig{
				Name:                   cfg.Name,
				Type:                   string(cfg.Type),
				LatencyConfig:          toLatencyConfig(cfg.LatencyCfg),
				NumericAttributeConfig: toNumericAttributeConfig(cfg.NumericAttributeCfg),
				ProbabilisticConfig:    toProbabilisticConfig(cfg.ProbabilisticCfg),
				StatusCodeConfig:       toStatusCodeConfig(cfg.StatusCodeCfg),
				StringAttributeConfig:  toStringAttributeConfig(cfg.StringAttributeCfg),
				RateLimitingConfig:     toRateLimitingConfig(cfg.RateLimitingCfg),
				SpanCountConfig:        toSpanCountConfig(cfg.SpanCountCfg),
				BooleanAttributeConfig: toBooleanAttributeConfig(cfg.BooleanAttributeCfg),
				OttlConditionConfig:    toOttlConditionConfig(cfg.OTTLConditionCfg),
				TraceStateConfig:       toTraceStateConfig(cfg.TraceStateCfg),
			},
		})
	}
	return out
}

func toRateAllocationConfig(cfgs []tailsamplingprocessor.RateAllocationCfg) []tail_sampling.RateAllocationConfig {
	var out []tail_sampling.RateAllocationConfig
	for _, cfg := range cfgs {
		out = append(out, tail_sampling.RateAllocationConfig{
			Policy:  cfg.Policy,
			Percent: cfg.Percent,
		})
	}
	return out
}

func toAndConfig(cfg tailsamplingprocessor.AndCfg) tail_sampling.AndConfig {
	return tail_sampling.AndConfig{
		SubPolicyConfig: toAndSubPolicyCfg(cfg.SubPolicyCfg),
	}
}

func toAndSubPolicyCfg(cfgs []tailsamplingprocessor.AndSubPolicyCfg) []tail_sampling.AndSubPolicyConfig {
	var out []tail_sampling.AndSubPolicyConfig
	for _, cfg := range cfgs {
		out = append(out, tail_sampling.AndSubPolicyConfig{
			SharedPolicyConfig: tail_sampling.SharedPolicyConfig{
				Name:                   cfg.Name,
				Type:                   string(cfg.Type),
				LatencyConfig:          toLatencyConfig(cfg.LatencyCfg),
				NumericAttributeConfig: toNumericAttributeConfig(cfg.NumericAttributeCfg),
				ProbabilisticConfig:    toProbabilisticConfig(cfg.ProbabilisticCfg),
				StatusCodeConfig:       toStatusCodeConfig(cfg.StatusCodeCfg),
				StringAttributeConfig:  toStringAttributeConfig(cfg.StringAttributeCfg),
				RateLimitingConfig:     toRateLimitingConfig(cfg.RateLimitingCfg),
				SpanCountConfig:        toSpanCountConfig(cfg.SpanCountCfg),
				BooleanAttributeConfig: toBooleanAttributeConfig(cfg.BooleanAttributeCfg),
				OttlConditionConfig:    toOttlConditionConfig(cfg.OTTLConditionCfg),
				TraceStateConfig:       toTraceStateConfig(cfg.TraceStateCfg),
			},
		})
	}
	return out
}

func toLatencyConfig(cfg tailsamplingprocessor.LatencyCfg) tail_sampling.LatencyConfig {
	return tail_sampling.LatencyConfig{
		ThresholdMs: cfg.ThresholdMs,
	}
}

func toNumericAttributeConfig(cfg tailsamplingprocessor.NumericAttributeCfg) tail_sampling.NumericAttributeConfig {
	return tail_sampling.NumericAttributeConfig{
		Key:         cfg.Key,
		MinValue:    cfg.MinValue,
		MaxValue:    cfg.MaxValue,
		InvertMatch: cfg.InvertMatch,
	}
}

func toProbabilisticConfig(cfg tailsamplingprocessor.ProbabilisticCfg) tail_sampling.ProbabilisticConfig {
	return tail_sampling.ProbabilisticConfig{
		HashSalt:           cfg.HashSalt,
		SamplingPercentage: cfg.SamplingPercentage,
	}
}

func toStatusCodeConfig(cfg tailsamplingprocessor.StatusCodeCfg) tail_sampling.StatusCodeConfig {
	return tail_sampling.StatusCodeConfig{
		StatusCodes: cfg.StatusCodes,
	}
}

func toStringAttributeConfig(cfg tailsamplingprocessor.StringAttributeCfg) tail_sampling.StringAttributeConfig {
	return tail_sampling.StringAttributeConfig{
		Key:                  cfg.Key,
		Values:               cfg.Values,
		EnabledRegexMatching: cfg.EnabledRegexMatching,
		CacheMaxSize:         cfg.CacheMaxSize,
		InvertMatch:          cfg.InvertMatch,
	}
}

func toRateLimitingConfig(cfg tailsamplingprocessor.RateLimitingCfg) tail_sampling.RateLimitingConfig {
	return tail_sampling.RateLimitingConfig{
		SpansPerSecond: cfg.SpansPerSecond,
	}
}

func toSpanCountConfig(cfg tailsamplingprocessor.SpanCountCfg) tail_sampling.SpanCountConfig {
	return tail_sampling.SpanCountConfig{
		MinSpans: cfg.MinSpans,
		MaxSpans: cfg.MaxSpans,
	}
}

func toBooleanAttributeConfig(cfg tailsamplingprocessor.BooleanAttributeCfg) tail_sampling.BooleanAttributeConfig {
	return tail_sampling.BooleanAttributeConfig{
		Key:   cfg.Key,
		Value: cfg.Value,
	}
}

func toOttlConditionConfig(cfg tailsamplingprocessor.OTTLConditionCfg) tail_sampling.OttlConditionConfig {
	return tail_sampling.OttlConditionConfig{
		ErrorMode:           tail_sampling.ErrorMode(cfg.ErrorMode),
		SpanConditions:      cfg.SpanConditions,
		SpanEventConditions: cfg.SpanEventConditions,
	}
}

func toTraceStateConfig(cfg tailsamplingprocessor.TraceStateCfg) tail_sampling.TraceStateConfig {
	return tail_sampling.TraceStateConfig{
		Key:    cfg.Key,
		Values: cfg.Values,
	}
}
