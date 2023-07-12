package tail_sampling

import (
	"encoding"
	"fmt"
	"strings"

	"github.com/grafana/agent/pkg/river"
	"github.com/mitchellh/mapstructure"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl"
	tsp "github.com/open-telemetry/opentelemetry-collector-contrib/processor/tailsamplingprocessor"
)

type PolicyConfig struct {
	SharedPolicyConfig SharedPolicyConfig `river:",squash"`

	// Configs for defining composite policy
	CompositeConfig CompositeConfig `river:"composite,block,optional"`

	// Configs for defining and policy
	AndConfig AndConfig `river:"and,block,optional"`
}

func (policyConfig PolicyConfig) Convert() tsp.PolicyCfg {
	var otelConfig tsp.PolicyCfg

	mustDecodeMapStructure(map[string]interface{}{
		"name":              policyConfig.SharedPolicyConfig.Name,
		"type":              policyConfig.SharedPolicyConfig.Type,
		"latency":           policyConfig.SharedPolicyConfig.LatencyConfig.Convert(),
		"numeric_attribute": policyConfig.SharedPolicyConfig.NumericAttributeConfig.Convert(),
		"probabilistic":     policyConfig.SharedPolicyConfig.ProbabilisticConfig.Convert(),
		"status_code":       policyConfig.SharedPolicyConfig.StatusCodeConfig.Convert(),
		"string_attribute":  policyConfig.SharedPolicyConfig.StringAttributeConfig.Convert(),
		"rate_limiting":     policyConfig.SharedPolicyConfig.RateLimitingConfig.Convert(),
		"span_count":        policyConfig.SharedPolicyConfig.SpanCountConfig.Convert(),
		"boolean_attribute": policyConfig.SharedPolicyConfig.BooleanAttributeConfig.Convert(),
		"ottl_condition":    policyConfig.SharedPolicyConfig.OttlConditionConfig.Convert(),
		"trace_state":       policyConfig.SharedPolicyConfig.TraceStateConfig.Convert(),
		"composite":         policyConfig.CompositeConfig.Convert(),
		"and":               policyConfig.AndConfig.Convert(),
	}, &otelConfig)

	return otelConfig
}

// This cannot currently have a Convert() because tsp.sharedPolicyCfg isn't public
type SharedPolicyConfig struct {
	Name                   string                 `river:"name,attr"`
	Type                   string                 `river:"type,attr"`
	LatencyConfig          LatencyConfig          `river:"latency,block,optional"`
	NumericAttributeConfig NumericAttributeConfig `river:"numeric_attribute,block,optional"`
	ProbabilisticConfig    ProbabilisticConfig    `river:"probabilistic,block,optional"`
	StatusCodeConfig       StatusCodeConfig       `river:"status_code,block,optional"`
	StringAttributeConfig  StringAttributeConfig  `river:"string_attribute,block,optional"`
	RateLimitingConfig     RateLimitingConfig     `river:"rate_limiting,block,optional"`
	SpanCountConfig        SpanCountConfig        `river:"span_count,block,optional"`
	BooleanAttributeConfig BooleanAttributeConfig `river:"boolean_attribute,block,optional"`
	OttlConditionConfig    OttlConditionConfig    `river:"ottl_condition,block,optional"`
	TraceStateConfig       TraceStateConfig       `river:"trace_state,block,optional"`
}

// LatencyConfig holds the configurable settings to create a latency filter sampling policy
// evaluator
type LatencyConfig struct {
	// ThresholdMs in milliseconds.
	ThresholdMs int64 `river:"threshold_ms,attr"`
}

func (latencyConfig LatencyConfig) Convert() tsp.LatencyCfg {
	otelConfig := tsp.LatencyCfg{}

	mustDecodeMapStructure(map[string]interface{}{
		"threshold_ms": latencyConfig.ThresholdMs,
	}, &otelConfig)

	return otelConfig
}

// NumericAttributeConfig holds the configurable settings to create a numeric attribute filter
// sampling policy evaluator.
type NumericAttributeConfig struct {
	// Tag that the filter is going to be matching against.
	Key string `river:"key,attr"`
	// MinValue is the minimum value of the attribute to be considered a match.
	MinValue int64 `river:"min_value,attr"`
	// MaxValue is the maximum value of the attribute to be considered a match.
	MaxValue int64 `river:"max_value,attr"`
}

func (numericAttributeConfig NumericAttributeConfig) Convert() tsp.NumericAttributeCfg {
	var otelConfig tsp.NumericAttributeCfg

	mustDecodeMapStructure(map[string]interface{}{
		"key":       numericAttributeConfig.Key,
		"min_value": numericAttributeConfig.MinValue,
		"max_value": numericAttributeConfig.MaxValue,
	}, &otelConfig)

	return otelConfig
}

// ProbabilisticConfig holds the configurable settings to create a probabilistic
// sampling policy evaluator.
type ProbabilisticConfig struct {
	// HashSalt allows one to configure the hashing salts. This is important in scenarios where multiple layers of collectors
	// have different sampling rates: if they use the same salt all passing one layer may pass the other even if they have
	// different sampling rates, configuring different salts avoids that.
	HashSalt string `river:"hash_salt,attr,optional"`
	// SamplingPercentage is the percentage rate at which traces are going to be sampled. Defaults to zero, i.e.: no sample.
	// Values greater or equal 100 are treated as "sample all traces".
	SamplingPercentage float64 `river:"sampling_percentage,attr"`
}

func (probabilisticConfig ProbabilisticConfig) Convert() tsp.ProbabilisticCfg {
	var otelConfig tsp.ProbabilisticCfg

	mustDecodeMapStructure(map[string]interface{}{
		"hash_salt":           probabilisticConfig.HashSalt,
		"sampling_percentage": probabilisticConfig.SamplingPercentage,
	}, &otelConfig)

	return otelConfig
}

// StatusCodeConfig holds the configurable settings to create a status code filter sampling
// policy evaluator.
type StatusCodeConfig struct {
	StatusCodes []string `river:"status_codes,attr"`
}

func (statusCodeConfig StatusCodeConfig) Convert() tsp.StatusCodeCfg {
	var otelConfig tsp.StatusCodeCfg

	mustDecodeMapStructure(map[string]interface{}{
		"status_codes": statusCodeConfig.StatusCodes,
	}, &otelConfig)

	return otelConfig
}

// StringAttributeConfig holds the configurable settings to create a string attribute filter
// sampling policy evaluator.
type StringAttributeConfig struct {
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

func (stringAttributeConfig StringAttributeConfig) Convert() tsp.StringAttributeCfg {
	var otelConfig tsp.StringAttributeCfg

	mustDecodeMapStructure(map[string]interface{}{
		"key":                    stringAttributeConfig.Key,
		"values":                 stringAttributeConfig.Values,
		"enabled_regex_matching": stringAttributeConfig.EnabledRegexMatching,
		"cache_max_size":         stringAttributeConfig.CacheMaxSize,
		"invert_match":           stringAttributeConfig.InvertMatch,
	}, &otelConfig)

	return otelConfig
}

// RateLimitingConfig holds the configurable settings to create a rate limiting
// sampling policy evaluator.
type RateLimitingConfig struct {
	// SpansPerSecond sets the limit on the maximum nuber of spans that can be processed each second.
	SpansPerSecond int64 `river:"spans_per_second,attr"`
}

func (rateLimitingConfig RateLimitingConfig) Convert() tsp.RateLimitingCfg {
	var otelConfig tsp.RateLimitingCfg

	mustDecodeMapStructure(map[string]interface{}{
		"spans_per_second": rateLimitingConfig.SpansPerSecond,
	}, &otelConfig)

	return otelConfig
}

// SpanCountConfig holds the configurable settings to create a Span Count filter sampling policy
// sampling policy evaluator
type SpanCountConfig struct {
	// Minimum number of spans in a Trace
	MinSpans int32 `river:"min_spans,attr"`
}

func (spanCountConfig SpanCountConfig) Convert() tsp.SpanCountCfg {
	var otelConfig tsp.SpanCountCfg

	mustDecodeMapStructure(map[string]interface{}{
		"min_spans": spanCountConfig.MinSpans,
	}, &otelConfig)

	return otelConfig
}

// BooleanAttributeConfig holds the configurable settings to create a boolean attribute filter
// sampling policy evaluator.
type BooleanAttributeConfig struct {
	// Tag that the filter is going to be matching against.
	Key string `river:"key,attr"`
	// Value indicate the bool value, either true or false to use when matching against attribute values.
	// BooleanAttribute Policy will apply exact value match on Value
	Value bool `river:"value,attr"`
}

func (booleanAttributeConfig BooleanAttributeConfig) Convert() tsp.BooleanAttributeCfg {
	var otelConfig tsp.BooleanAttributeCfg

	mustDecodeMapStructure(map[string]interface{}{
		"key":   booleanAttributeConfig.Key,
		"value": booleanAttributeConfig.Value,
	}, &otelConfig)

	return otelConfig
}

// The error mode determines whether to ignore or propagate
// errors with evaluating OTTL conditions.
type ErrorMode string

const (
	// "ignore" causes evaluation to continue to the next statement.
	ErrorModeIgnore ErrorMode = "ignore"
	// "propagate" causes the evaluation to be false and an error is returned.
	ErrorModePropagate ErrorMode = "propagate"
)

var (
	_ river.Validator          = (*ErrorMode)(nil)
	_ encoding.TextUnmarshaler = (*ErrorMode)(nil)
)

// Validate implements river.Validator.
func (e *ErrorMode) Validate() error {
	if e == nil {
		return nil
	}

	var ottlError ottl.ErrorMode
	return ottlError.UnmarshalText([]byte(string(*e)))
}

// Convert the River type to the Otel type
func (e *ErrorMode) Convert() ottl.ErrorMode {
	if e == nil || *e == "" {
		return ottl.ErrorMode("")
	}

	var ottlError ottl.ErrorMode
	err := ottlError.UnmarshalText([]byte(string(*e)))

	//TODO: Rework this to return an error instead of panicking
	if err != nil {
		panic(err)
	}

	return ottlError
}

func (e *ErrorMode) UnmarshalText(text []byte) error {
	if e == nil {
		return nil
	}

	str := ErrorMode(strings.ToLower(string(text)))
	switch str {
	case ErrorModeIgnore, ErrorModePropagate:
		*e = str
		return nil
	default:
		return fmt.Errorf("unknown error mode %v", str)
	}
}

// OttlConditionConfig holds the configurable setting to create a OTTL condition filter
// sampling policy evaluator.
type OttlConditionConfig struct {
	ErrorMode           ErrorMode `river:"error_mode,attr"`
	SpanConditions      []string  `river:"span,attr,optional"`
	SpanEventConditions []string  `river:"spanevent,attr,optional"`
}

func (ottlConditionConfig OttlConditionConfig) Convert() tsp.OTTLConditionCfg {
	var otelConfig tsp.OTTLConditionCfg

	mustDecodeMapStructure(map[string]interface{}{
		"error_mode": ottlConditionConfig.ErrorMode.Convert(),
		"span":       ottlConditionConfig.SpanConditions,
		"spanevent":  ottlConditionConfig.SpanEventConditions,
	}, &otelConfig)

	return otelConfig
}

type TraceStateConfig struct {
	// Tag that the filter is going to be matching against.
	Key string `river:"key,attr"`
	// Values indicate the set of values to use when matching against trace_state values.
	Values []string `river:"values,attr"`
}

func (traceStateConfig TraceStateConfig) Convert() tsp.TraceStateCfg {
	var otelConfig tsp.TraceStateCfg

	mustDecodeMapStructure(map[string]interface{}{
		"key":    traceStateConfig.Key,
		"values": traceStateConfig.Values,
	}, &otelConfig)

	return otelConfig
}

// CompositeConfig holds the configurable settings to create a composite
// sampling policy evaluator.
type CompositeConfig struct {
	MaxTotalSpansPerSecond int64                      `river:"max_total_spans_per_second,attr"`
	PolicyOrder            []string                   `river:"policy_order,attr"`
	SubPolicyCfg           []CompositeSubPolicyConfig `river:"composite_sub_policy,block,optional"`
	RateAllocation         []RateAllocationConfig     `river:"rate_allocation,block,optional"`
}

func (compositeConfig CompositeConfig) Convert() tsp.CompositeCfg {
	var otelConfig tsp.CompositeCfg

	var otelCompositeSubPolicyCfg []tsp.CompositeSubPolicyCfg
	for _, subPolicyCfg := range compositeConfig.SubPolicyCfg {
		otelCompositeSubPolicyCfg = append(otelCompositeSubPolicyCfg, subPolicyCfg.Convert())
	}

	var otelRateAllocationCfg []tsp.RateAllocationCfg
	for _, rateAllocation := range compositeConfig.RateAllocation {
		otelRateAllocationCfg = append(otelRateAllocationCfg, rateAllocation.Convert())
	}

	mustDecodeMapStructure(map[string]interface{}{
		"max_total_spans_per_second": compositeConfig.MaxTotalSpansPerSecond,
		"policy_order":               compositeConfig.PolicyOrder,
		"composite_sub_policy":       otelCompositeSubPolicyCfg,
		"rate_allocation":            otelRateAllocationCfg,
	}, &otelConfig)

	return otelConfig
}

// CompositeSubPolicyConfig holds the common configuration to all policies under composite policy.
type CompositeSubPolicyConfig struct {
	SharedPolicyConfig SharedPolicyConfig `river:",squash"`

	// Configs for and policy evaluator.
	AndConfig AndConfig `river:"and,block,optional"`
}

func (compositeSubPolicyConfig CompositeSubPolicyConfig) Convert() tsp.CompositeSubPolicyCfg {
	var otelConfig tsp.CompositeSubPolicyCfg

	mustDecodeMapStructure(map[string]interface{}{
		"name":              compositeSubPolicyConfig.SharedPolicyConfig.Name,
		"type":              compositeSubPolicyConfig.SharedPolicyConfig.Type,
		"latency":           compositeSubPolicyConfig.SharedPolicyConfig.LatencyConfig.Convert(),
		"numeric_attribute": compositeSubPolicyConfig.SharedPolicyConfig.NumericAttributeConfig.Convert(),
		"probabilistic":     compositeSubPolicyConfig.SharedPolicyConfig.ProbabilisticConfig.Convert(),
		"status_code":       compositeSubPolicyConfig.SharedPolicyConfig.StatusCodeConfig.Convert(),
		"string_attribute":  compositeSubPolicyConfig.SharedPolicyConfig.StringAttributeConfig.Convert(),
		"rate_limiting":     compositeSubPolicyConfig.SharedPolicyConfig.RateLimitingConfig.Convert(),
		"span_count":        compositeSubPolicyConfig.SharedPolicyConfig.SpanCountConfig.Convert(),
		"boolean_attribute": compositeSubPolicyConfig.SharedPolicyConfig.BooleanAttributeConfig.Convert(),
		"ottl_condition":    compositeSubPolicyConfig.SharedPolicyConfig.OttlConditionConfig.Convert(),
		"trace_state":       compositeSubPolicyConfig.SharedPolicyConfig.TraceStateConfig.Convert(),
		"and":               compositeSubPolicyConfig.AndConfig.Convert(),
	}, &otelConfig)

	return otelConfig
}

// RateAllocationConfig  used within composite policy
type RateAllocationConfig struct {
	Policy  string `river:"policy,attr"`
	Percent int64  `river:"percent,attr"`
}

func (rateAllocationConfig RateAllocationConfig) Convert() tsp.RateAllocationCfg {
	var otelConfig tsp.RateAllocationCfg

	mustDecodeMapStructure(map[string]interface{}{
		"policy":  rateAllocationConfig.Policy,
		"percent": rateAllocationConfig.Percent,
	}, &otelConfig)

	return otelConfig
}

type AndConfig struct {
	SubPolicyConfig []AndSubPolicyConfig `river:"and_sub_policy,block"`
}

func (andConfig AndConfig) Convert() tsp.AndCfg {
	var otelConfig tsp.AndCfg

	var otelPolicyCfgs []tsp.AndSubPolicyCfg
	for _, subPolicyCfg := range andConfig.SubPolicyConfig {
		otelPolicyCfgs = append(otelPolicyCfgs, subPolicyCfg.Convert())
	}

	mustDecodeMapStructure(map[string]interface{}{
		"and_sub_policy": otelPolicyCfgs,
	}, &otelConfig)

	return otelConfig
}

// AndSubPolicyConfig holds the common configuration to all policies under and policy.
type AndSubPolicyConfig struct {
	SharedPolicyConfig SharedPolicyConfig `river:",squash"`
}

func (andSubPolicyConfig AndSubPolicyConfig) Convert() tsp.AndSubPolicyCfg {
	var otelConfig tsp.AndSubPolicyCfg

	mustDecodeMapStructure(map[string]interface{}{
		"name":              andSubPolicyConfig.SharedPolicyConfig.Name,
		"type":              andSubPolicyConfig.SharedPolicyConfig.Type,
		"latency":           andSubPolicyConfig.SharedPolicyConfig.LatencyConfig.Convert(),
		"numeric_attribute": andSubPolicyConfig.SharedPolicyConfig.NumericAttributeConfig.Convert(),
		"probabilistic":     andSubPolicyConfig.SharedPolicyConfig.ProbabilisticConfig.Convert(),
		"status_code":       andSubPolicyConfig.SharedPolicyConfig.StatusCodeConfig.Convert(),
		"string_attribute":  andSubPolicyConfig.SharedPolicyConfig.StringAttributeConfig.Convert(),
		"rate_limiting":     andSubPolicyConfig.SharedPolicyConfig.RateLimitingConfig.Convert(),
		"span_count":        andSubPolicyConfig.SharedPolicyConfig.SpanCountConfig.Convert(),
		"boolean_attribute": andSubPolicyConfig.SharedPolicyConfig.BooleanAttributeConfig.Convert(),
		"ottl_condition":    andSubPolicyConfig.SharedPolicyConfig.OttlConditionConfig.Convert(),
		"trace_state":       andSubPolicyConfig.SharedPolicyConfig.TraceStateConfig.Convert(),
	}, &otelConfig)

	return otelConfig
}

// TODO: Why do we do this? Can we not just create the Otel types directly?
func mustDecodeMapStructure(source map[string]interface{}, otelConfig interface{}) {
	err := mapstructure.Decode(source, otelConfig)

	//TODO: Rework this to return an error instead of panicking
	if err != nil {
		panic(err)
	}
}
