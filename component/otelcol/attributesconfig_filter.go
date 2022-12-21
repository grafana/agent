package otelcol

import (
	"github.com/open-telemetry/opentelemetry-collector-contrib/external/coreinternal/processor/filterconfig"
	"github.com/open-telemetry/opentelemetry-collector-contrib/external/coreinternal/processor/filterset"
	"github.com/open-telemetry/opentelemetry-collector-contrib/external/coreinternal/processor/filterset/regexp"
	"go.opentelemetry.io/collector/pdata/plog"
)

// MatchConfig has two optional MatchProperties one to define what is processed
// by the processor, captured under the 'include' and the second, exclude, to
// define what is excluded from the processor.
type MatchConfig struct {
	Include *MatchProperties `river:"include,block,optional"`
	Exclude *MatchProperties `river:"exclude,block,optional"`
}

// Convert converts args into the upstream type.
func (args *MatchConfig) Convert() *filterconfig.MatchConfig {
	if args == nil {
		return nil
	}

	return &filterconfig.MatchConfig{
		Include: args.Include.Convert(),
		Exclude: args.Exclude.Convert(),
	}
}

type RegexpConfig struct {
	// CacheEnabled determines whether match results are LRU cached to make subsequent matches faster.
	// Cache size is unlimited unless CacheMaxNumEntries is also specified.
	CacheEnabled bool `river:"cacheenabled,attr,optional"`
	// CacheMaxNumEntries is the max number of entries of the LRU cache that stores match results.
	// CacheMaxNumEntries is ignored if CacheEnabled is false.
	CacheMaxNumEntries int `river:"cachemaxnumentries,attr,optional"`
}

// Convert converts args into the upstream type.
func (args *RegexpConfig) Convert() *regexp.Config {
	if args == nil {
		return nil
	}

	return &regexp.Config{
		CacheEnabled:       args.CacheEnabled,
		CacheMaxNumEntries: args.CacheMaxNumEntries,
	}
}

//TODO: Does everything work correctly if neither exlcude nor include is specified?

// MatchProperties specifies the set of properties in a spans/log/metric to match
// against and if the input data should be included or excluded from the
// processor. At least one of services (spans only), names or
// attributes must be specified. It is supported to have all specified, but
// this requires all the properties to match for the inclusion/exclusion to
// occur.
// The following are examples of invalid configurations:
//
//	attributes/bad1:
//	  # This is invalid because include is specified with neither services or
//	  # attributes.
//	  include:
//	  actions: ...
//
//	span/bad2:
//	  exclude:
//	  	# This is invalid because services, span_names and attributes have empty values.
//	    services:
//	    span_names:
//	    attributes:
//	  actions: ...
//
// Please refer to processor/attributesprocessor/testdata/config.yaml and
// processor/spanprocessor/testdata/config.yaml for valid configurations.
type MatchProperties struct {
	MatchType    string        `river:"match_type,attr"`
	RegexpConfig *RegexpConfig `river:"regexp,block,optional"`

	// Note: For spans, one of Services, SpanNames, Attributes, Resources or Libraries must be specified with a
	// non-empty value for a valid configuration.

	// For logs, one of LogNames, Attributes, Resources or Libraries must be specified with a
	// non-empty value for a valid configuration.

	// For metrics, one of MetricNames, Expressions, or ResourceAttributes must be specified with a
	// non-empty value for a valid configuration.

	// Services specify the list of items to match service name against.
	// A match occurs if the span's service name matches at least one item in this list.
	// This is an optional field.
	Services []string `river:"services,attr,optional"`

	// SpanNames specify the list of items to match span name against.
	// A match occurs if the span name matches at least one item in this list.
	// This is an optional field.
	SpanNames []string `river:"span_names,attr,optional"`

	// LogBodies is a list of strings that the LogRecord's body field must match
	// against.
	LogBodies []string `river:"log_bodies,attr,optional"`

	// LogSeverityTexts is a list of strings that the LogRecord's severity text field must match
	// against.
	LogSeverityTexts []string `river:"log_severity_texts,attr,optional"`

	// LogSeverityNumber defines how to match against a log record's SeverityNumber, if defined.
	LogSeverityNumber *LogSeverityNumberMatchProperties `river:"log_severity_number,block,optional"`

	// MetricNames is a list of strings to match metric name against.
	// A match occurs if metric name matches at least one item in the list.
	// This field is optional.
	MetricNames []string `river:"metric_names,attr,optional"`

	// Attributes specifies the list of attributes to match against.
	// All of these attributes must match exactly for a match to occur.
	// Only match_type=strict is allowed if "attributes" are specified.
	// This is an optional field.
	Attributes AttributesCollection `river:"attributes,block,optional"`

	// Resources specify the list of items to match the resources against.
	// A match occurs if the data's resources match at least one item in this list.
	// This is an optional field.
	Resources AttributesCollection `river:"resources,block,optional"`

	// Libraries specify the list of items to match the implementation library against.
	// A match occurs if the span's implementation library matches at least one item in this list.
	// This is an optional field.
	Libraries InstrumentationLibraries `river:"libraries,block,optional"`

	// SpanKinds specify the list of items to match the span kind against.
	// A match occurs if the span's span kind matches at least one item in this list.
	// This is an optional field
	SpanKinds []string `river:"span_kinds,attr,optional"`
}

// Convert converts args into the upstream type.
func (args *MatchProperties) Convert() *filterconfig.MatchProperties {
	if args == nil {
		return nil
	}

	//TODO: Deep copy the RegexpConfig pointer
	return &filterconfig.MatchProperties{
		Config: filterset.Config{
			MatchType:    filterset.MatchType(args.MatchType),
			RegexpConfig: args.RegexpConfig.Convert(),
		},
		Services:          args.Services,
		SpanNames:         args.SpanNames,
		LogBodies:         args.LogBodies,
		LogSeverityTexts:  args.LogSeverityTexts,
		LogSeverityNumber: args.LogSeverityNumber.Convert(),
		MetricNames:       args.MetricNames,
		Attributes:        convertAttributesCollection(args.Attributes),
		Resources:         convertAttributesCollection(args.Resources),
		Libraries:         convertInstrumentationLibraries(args.Libraries),
		SpanKinds:         args.SpanKinds,
	}
}

// TODO: Use generics for these?
func convertAttributesCollection(v AttributesCollection) []filterconfig.Attribute {
	res := make([]filterconfig.Attribute, 0)

	for _, elem := range v.Attributes {
		res = append(res, *elem.Convert())
	}

	return res
}

func convertInstrumentationLibraries(v InstrumentationLibraries) []filterconfig.InstrumentationLibrary {
	res := make([]filterconfig.InstrumentationLibrary, 0)

	for _, elem := range v.Libraries {
		res = append(res, *elem.Convert())
	}

	return res
}

// TODO: This is used by both "resources" and "attributes", so maybe it should not be called "attribute"
type AttributesCollection struct {
	Attributes []*Attribute `river:"attribute,block,optional"`
}

// Attribute specifies the attribute key and optional value to match against.
type Attribute struct {
	// Key specifies the attribute key.
	Key string `river:"key,attr"`

	// Values specifies the value to match against.
	// If it is not set, any value will match.
	Value interface{} `river:"value,attr,optional"`
}

// Convert converts args into the upstream type.
func (args *Attribute) Convert() *filterconfig.Attribute {
	if args == nil {
		return nil
	}

	return &filterconfig.Attribute{
		Key:   args.Key,
		Value: args.Value,
	}
}

type InstrumentationLibraries struct {
	Libraries []*InstrumentationLibrary `river:"library,block,optional"`
}

// InstrumentationLibrary specifies the instrumentation library and optional version to match against.
type InstrumentationLibrary struct {
	Name string `river:"name,attr"`
	// version match
	//  expected actual  match
	//  nil      <blank> yes
	//  nil      1       yes
	//  <blank>  <blank> yes
	//  <blank>  1       no
	//  1        <blank> no
	//  1        1       yes
	Version *string `river:"version,attr"`
}

// Convert converts args into the upstream type.
func (args *InstrumentationLibrary) Convert() *filterconfig.InstrumentationLibrary {
	if args == nil {
		return nil
	}

	return &filterconfig.InstrumentationLibrary{
		Name:    args.Name,
		Version: args.Version,
	}
	//TODO: The Version should copy the string, not point to the same pointer
}

// LogSeverityNumberMatchProperties defines how to match based on a log record's SeverityNumber field.
type LogSeverityNumberMatchProperties struct {
	// Min is the lowest severity that may be matched.
	// e.g. if this is plog.SeverityNumberInfo, INFO, WARN, ERROR, and FATAL logs will match.
	Min int32 `river:"min,attr"`

	// MatchUndefined controls whether logs with "undefined" severity matches.
	// If this is true, entries with undefined severity will match.
	MatchUndefined bool `river:"match_undefined,attr"`
}

// Convert converts args into the upstream type.
func (args *LogSeverityNumberMatchProperties) Convert() *filterconfig.LogSeverityNumberMatchProperties {
	if args == nil {
		return nil
	}

	return &filterconfig.LogSeverityNumberMatchProperties{
		Min:            plog.SeverityNumber(args.Min),
		MatchUndefined: args.MatchUndefined,
	}
}
