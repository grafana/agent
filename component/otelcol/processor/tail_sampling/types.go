package tail_sampling

import (
	"github.com/mitchellh/mapstructure"
	tsp "github.com/open-telemetry/opentelemetry-collector-contrib/processor/tailsamplingprocessor"
)

type PolicyCfg struct {
	SharedPolicyCfg SharedPolicyCfg `river:",squash"`

	// Configs for defining composite policy
	CompositeCfg CompositeCfg `river:"composite,block,optional"`

	// Configs for defining and policy
	AndCfg AndCfg `river:"and,block,optional"`
}

func (policyCfg PolicyCfg) Convert() tsp.PolicyCfg {
	var otelCfg tsp.PolicyCfg

	mustDecodeMapStructure(map[string]interface{}{
		"name":              policyCfg.SharedPolicyCfg.Name,
		"type":              policyCfg.SharedPolicyCfg.Type,
		"latency":           policyCfg.SharedPolicyCfg.LatencyCfg.Convert(),
		"numeric_attribute": policyCfg.SharedPolicyCfg.NumericAttributeCfg.Convert(),
		"probabilistic":     policyCfg.SharedPolicyCfg.ProbabilisticCfg.Convert(),
		"status_code":       policyCfg.SharedPolicyCfg.StatusCodeCfg.Convert(),
		"string_attribute":  policyCfg.SharedPolicyCfg.StringAttributeCfg.Convert(),
		"rate_limiting":     policyCfg.SharedPolicyCfg.RateLimitingCfg.Convert(),
		"span_count":        policyCfg.SharedPolicyCfg.SpanCountCfg.Convert(),
		"trace_state":       policyCfg.SharedPolicyCfg.TraceStateCfg.Convert(),
		"composite":         policyCfg.CompositeCfg.Convert(),
		"and":               policyCfg.AndCfg.Convert(),
	}, &otelCfg)

	return otelCfg
}

// This cannot currently have a Convert() because tsp.sharedPolicyCfg isn't public
type SharedPolicyCfg struct {
	Name                string              `river:"name,attr"`
	Type                string              `river:"type,attr"`
	LatencyCfg          LatencyCfg          `river:"latency,block,optional"`
	NumericAttributeCfg NumericAttributeCfg `river:"numeric_attribute,block,optional"`
	ProbabilisticCfg    ProbabilisticCfg    `river:"probabilistic,block,optional"`
	StatusCodeCfg       StatusCodeCfg       `river:"status_code,block,optional"`
	StringAttributeCfg  StringAttributeCfg  `river:"string_attribute,block,optional"`
	RateLimitingCfg     RateLimitingCfg     `river:"rate_limiting,block,optional"`
	SpanCountCfg        SpanCountCfg        `river:"span_count,block,optional"`
	TraceStateCfg       TraceStateCfg       `river:"trace_state,block,optional"`
}

// LatencyCfg holds the configurable settings to create a latency filter sampling policy
// evaluator
type LatencyCfg struct {
	// ThresholdMs in milliseconds.
	ThresholdMs int64 `river:"threshold_ms,attr"`
}

func (latencyCfg LatencyCfg) Convert() tsp.LatencyCfg {
	otelCfg := tsp.LatencyCfg{}

	mustDecodeMapStructure(map[string]interface{}{
		"threshold_ms": latencyCfg.ThresholdMs,
	}, &otelCfg)

	return otelCfg
}

// NumericAttributeCfg holds the configurable settings to create a numeric attribute filter
// sampling policy evaluator.
type NumericAttributeCfg struct {
	// Tag that the filter is going to be matching against.
	Key string `river:"key,attr"`
	// MinValue is the minimum value of the attribute to be considered a match.
	MinValue int64 `river:"min_value,attr"`
	// MaxValue is the maximum value of the attribute to be considered a match.
	MaxValue int64 `river:"max_value,attr"`
}

func (numericAttributeCfg NumericAttributeCfg) Convert() tsp.NumericAttributeCfg {
	var otelCfg tsp.NumericAttributeCfg

	mustDecodeMapStructure(map[string]interface{}{
		"key":       numericAttributeCfg.Key,
		"min_value": numericAttributeCfg.MinValue,
		"max_value": numericAttributeCfg.MaxValue,
	}, &otelCfg)

	return otelCfg
}

// ProbabilisticCfg holds the configurable settings to create a probabilistic
// sampling policy evaluator.
type ProbabilisticCfg struct {
	// HashSalt allows one to configure the hashing salts. This is important in scenarios where multiple layers of collectors
	// have different sampling rates: if they use the same salt all passing one layer may pass the other even if they have
	// different sampling rates, configuring different salts avoids that.
	HashSalt string `river:"hash_salt,attr,optional"`
	// SamplingPercentage is the percentage rate at which traces are going to be sampled. Defaults to zero, i.e.: no sample.
	// Values greater or equal 100 are treated as "sample all traces".
	SamplingPercentage float64 `river:"sampling_percentage,attr"`
}

func (probabilisticCfg ProbabilisticCfg) Convert() tsp.ProbabilisticCfg {
	var otelCfg tsp.ProbabilisticCfg

	mustDecodeMapStructure(map[string]interface{}{
		"hash_salt":           probabilisticCfg.HashSalt,
		"sampling_percentage": probabilisticCfg.SamplingPercentage,
	}, &otelCfg)

	return otelCfg
}

// StatusCodeCfg holds the configurable settings to create a status code filter sampling
// policy evaluator.
type StatusCodeCfg struct {
	StatusCodes []string `river:"status_codes,attr"`
}

func (statusCodeCfg StatusCodeCfg) Convert() tsp.StatusCodeCfg {
	var otelCfg tsp.StatusCodeCfg

	mustDecodeMapStructure(map[string]interface{}{
		"status_codes": statusCodeCfg.StatusCodes,
	}, &otelCfg)

	return otelCfg
}

// StringAttributeCfg holds the configurable settings to create a string attribute filter
// sampling policy evaluator.
type StringAttributeCfg struct {
	// Tag that the filter is going to be matching against.
	Key string `river:"key,attr"`
	// Values indicate the set of values or regular expressions to use when matching against attribute values.
	// StringAttribute Policy will apply exact value match on Values unless EnabledRegexMatching is true.
	Values []string `river:"values,attr"`
	// EnabledRegexMatching determines whether match attribute values by regexp string.
	EnabledRegexMatching bool `river:"enabled_regex_matching,attr,optional"`
	// CacheMaxSize is the maximum number of attribute entries of LRU Cache that stores the matched result
	// from the regular expressions defined in Values.
	// CacheMaxSize will not be used if EnabledRegexMatching is set to false.
	CacheMaxSize int `river:"cache_max_size,attr,optional"`
	// InvertMatch indicates that values or regular expressions must not match against attribute values.
	// If InvertMatch is true and Values is equal to 'acme', all other values will be sampled except 'acme'.
	// Also, if the specified Key does not match on any resource or span attributes, data will be sampled.
	InvertMatch bool `river:"invert_match,attr,optional"`
}

func (stringAttributeCfg StringAttributeCfg) Convert() tsp.StringAttributeCfg {
	var otelCfg tsp.StringAttributeCfg

	mustDecodeMapStructure(map[string]interface{}{
		"key":                    stringAttributeCfg.Key,
		"values":                 stringAttributeCfg.Values,
		"enabled_regex_matching": stringAttributeCfg.EnabledRegexMatching,
		"cache_max_size":         stringAttributeCfg.CacheMaxSize,
		"invert_match":           stringAttributeCfg.InvertMatch,
	}, &otelCfg)

	return otelCfg
}

// RateLimitingCfg holds the configurable settings to create a rate limiting
// sampling policy evaluator.
type RateLimitingCfg struct {
	// SpansPerSecond sets the limit on the maximum nuber of spans that can be processed each second.
	SpansPerSecond int64 `river:"spans_per_second,attr"`
}

func (rateLimitingCfg RateLimitingCfg) Convert() tsp.RateLimitingCfg {
	var otelCfg tsp.RateLimitingCfg

	mustDecodeMapStructure(map[string]interface{}{
		"spans_per_second": rateLimitingCfg.SpansPerSecond,
	}, &otelCfg)

	return otelCfg
}

// SpanCountCfg holds the configurable settings to create a Span Count filter sampling policy
// sampling policy evaluator
type SpanCountCfg struct {
	// Minimum number of spans in a Trace
	MinSpans int32 `river:"min_spans,attr"`
}

func (spanCountCfg SpanCountCfg) Convert() tsp.SpanCountCfg {
	var otelCfg tsp.SpanCountCfg

	mustDecodeMapStructure(map[string]interface{}{
		"min_spans": spanCountCfg.MinSpans,
	}, &otelCfg)

	return otelCfg
}

type TraceStateCfg struct {
	// Tag that the filter is going to be matching against.
	Key string `river:"key,attr"`
	// Values indicate the set of values to use when matching against trace_state values.
	Values []string `river:"values,attr"`
}

func (traceStateCfg TraceStateCfg) Convert() tsp.TraceStateCfg {
	var otelCfg tsp.TraceStateCfg

	mustDecodeMapStructure(map[string]interface{}{
		"key":    traceStateCfg.Key,
		"values": traceStateCfg.Values,
	}, &otelCfg)

	return otelCfg
}

// CompositeCfg holds the configurable settings to create a composite
// sampling policy evaluator.
type CompositeCfg struct {
	MaxTotalSpansPerSecond int64                   `river:"max_total_spans_per_second,attr"`
	PolicyOrder            []string                `river:"policy_order,attr"`
	SubPolicyCfg           []CompositeSubPolicyCfg `river:"composite_sub_policy,block,optional"`
	RateAllocation         []RateAllocationCfg     `river:"rate_allocation,block,optional"`
}

func (compositeCfg CompositeCfg) Convert() tsp.CompositeCfg {
	var otelConfig tsp.CompositeCfg

	var otelCompositeSubPolicyCfg []tsp.CompositeSubPolicyCfg
	for _, subPolicyCfg := range compositeCfg.SubPolicyCfg {
		otelCompositeSubPolicyCfg = append(otelCompositeSubPolicyCfg, subPolicyCfg.Convert())
	}

	var otelRateAllocationCfg []tsp.RateAllocationCfg
	for _, rateAllocation := range compositeCfg.RateAllocation {
		otelRateAllocationCfg = append(otelRateAllocationCfg, rateAllocation.Convert())
	}

	mustDecodeMapStructure(map[string]interface{}{
		"max_total_spans_per_second": compositeCfg.MaxTotalSpansPerSecond,
		"policy_order":               compositeCfg.PolicyOrder,
		"composite_sub_policy":       otelCompositeSubPolicyCfg,
		"rate_allocation":            otelRateAllocationCfg,
	}, &otelConfig)

	return otelConfig
}

// CompositeSubPolicyCfg holds the common configuration to all policies under composite policy.
type CompositeSubPolicyCfg struct {
	SharedPolicyCfg SharedPolicyCfg `river:",squash"`

	// Configs for and policy evaluator.
	AndCfg AndCfg `river:"and,block,optional"`
}

func (compositeSubPolicyCfg CompositeSubPolicyCfg) Convert() tsp.CompositeSubPolicyCfg {
	var otelCfg tsp.CompositeSubPolicyCfg

	mustDecodeMapStructure(map[string]interface{}{
		"name":              compositeSubPolicyCfg.SharedPolicyCfg.Name,
		"type":              compositeSubPolicyCfg.SharedPolicyCfg.Type,
		"latency":           compositeSubPolicyCfg.SharedPolicyCfg.LatencyCfg.Convert(),
		"numeric_attribute": compositeSubPolicyCfg.SharedPolicyCfg.NumericAttributeCfg.Convert(),
		"probabilistic":     compositeSubPolicyCfg.SharedPolicyCfg.ProbabilisticCfg.Convert(),
		"status_code":       compositeSubPolicyCfg.SharedPolicyCfg.StatusCodeCfg.Convert(),
		"string_attribute":  compositeSubPolicyCfg.SharedPolicyCfg.StringAttributeCfg.Convert(),
		"rate_limiting":     compositeSubPolicyCfg.SharedPolicyCfg.RateLimitingCfg.Convert(),
		"span_count":        compositeSubPolicyCfg.SharedPolicyCfg.SpanCountCfg.Convert(),
		"trace_state":       compositeSubPolicyCfg.SharedPolicyCfg.TraceStateCfg.Convert(),
		"and":               compositeSubPolicyCfg.AndCfg.Convert(),
	}, &otelCfg)

	return otelCfg
}

// RateAllocationCfg  used within composite policy
type RateAllocationCfg struct {
	Policy  string `river:"policy,attr"`
	Percent int64  `river:"percent,attr"`
}

func (rateAllocationCfg RateAllocationCfg) Convert() tsp.RateAllocationCfg {
	var otelCfg tsp.RateAllocationCfg

	mustDecodeMapStructure(map[string]interface{}{
		"policy":  rateAllocationCfg.Policy,
		"percent": rateAllocationCfg.Percent,
	}, &otelCfg)

	return otelCfg
}

type AndCfg struct {
	SubPolicyCfg []AndSubPolicyCfg `river:"and_sub_policy,block"`
}

func (andCfg AndCfg) Convert() tsp.AndCfg {
	var otelConfig tsp.AndCfg

	var otelPolicyCfgs []tsp.AndSubPolicyCfg
	for _, subPolicyCfg := range andCfg.SubPolicyCfg {
		otelPolicyCfgs = append(otelPolicyCfgs, subPolicyCfg.Convert())
	}

	mustDecodeMapStructure(map[string]interface{}{
		"and_sub_policy": otelPolicyCfgs,
	}, &otelConfig)

	return otelConfig
}

// AndSubPolicyCfg holds the common configuration to all policies under and policy.
type AndSubPolicyCfg struct {
	SharedPolicyCfg SharedPolicyCfg `river:",squash"`
}

func (andSubPolicyCfg AndSubPolicyCfg) Convert() tsp.AndSubPolicyCfg {
	var otelCfg tsp.AndSubPolicyCfg

	mustDecodeMapStructure(map[string]interface{}{
		"name":              andSubPolicyCfg.SharedPolicyCfg.Name,
		"type":              andSubPolicyCfg.SharedPolicyCfg.Type,
		"latency":           andSubPolicyCfg.SharedPolicyCfg.LatencyCfg.Convert(),
		"numeric_attribute": andSubPolicyCfg.SharedPolicyCfg.NumericAttributeCfg.Convert(),
		"probabilistic":     andSubPolicyCfg.SharedPolicyCfg.ProbabilisticCfg.Convert(),
		"status_code":       andSubPolicyCfg.SharedPolicyCfg.StatusCodeCfg.Convert(),
		"string_attribute":  andSubPolicyCfg.SharedPolicyCfg.StringAttributeCfg.Convert(),
		"rate_limiting":     andSubPolicyCfg.SharedPolicyCfg.RateLimitingCfg.Convert(),
		"span_count":        andSubPolicyCfg.SharedPolicyCfg.SpanCountCfg.Convert(),
		"trace_state":       andSubPolicyCfg.SharedPolicyCfg.TraceStateCfg.Convert(),
	}, &otelCfg)

	return otelCfg
}

func mustDecodeMapStructure(source map[string]interface{}, otelCfg interface{}) {
	err := mapstructure.Decode(source, otelCfg)

	if err != nil {
		panic(err)
	}
}
