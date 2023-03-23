package otelcol

import "strings"

// MatchConfig has two optional MatchProperties:
//  1. 'include': to define what is processed by the processor.
//  2. 'exclude': to define what is excluded from the processor.
//
// If both 'include' and 'exclude' are specified, the 'include' properties are checked
// before the 'exclude' properties.
type MatchConfig struct {
	Include *MatchProperties `river:"include,block,optional"`
	Exclude *MatchProperties `river:"exclude,block,optional"`
}

// MatchProperties specifies the set of properties in a spans/log/metric to match
// against and if the input data should be included or excluded from the processor.
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
	Services []string `river:"services,attr,optional"`

	// SpanNames specify the list of items to match span name against.
	// A match occurs if the span name matches at least one item in this list.
	SpanNames []string `river:"span_names,attr,optional"`

	// LogBodies is a list of strings that the LogRecord's body field must match against.
	LogBodies []string `river:"log_bodies,attr,optional"`

	// LogSeverityTexts is a list of strings that the LogRecord's severity text field must match against.
	LogSeverityTexts []string `river:"log_severity_texts,attr,optional"`

	// LogSeverityNumber defines how to match against a log record's SeverityNumber, if defined.
	LogSeverityNumber *LogSeverityNumberMatchProperties `river:"log_severity_number,block,optional"`

	// MetricNames is a list of strings to match metric name against.
	// A match occurs if metric name matches at least one item in the list.
	MetricNames []string `river:"metric_names,attr,optional"`

	// Attributes specifies the list of attributes to match against.
	// All of these attributes must match exactly for a match to occur.
	// Only match_type=strict is allowed if "attributes" are specified.
	Attributes []Attribute `river:"attribute,block,optional"`

	// Resources specify the list of items to match the resources against.
	// A match occurs if the data's resources match at least one item in this list.
	Resources []Attribute `river:"resource,block,optional"`

	// Libraries specify the list of items to match the implementation library against.
	// A match occurs if the span's implementation library matches at least one item in this list.
	Libraries []InstrumentationLibrary `river:"library,block,optional"`

	// SpanKinds specify the list of items to match the span kind against.
	// A match occurs if the span's span kind matches at least one item in this list.
	SpanKinds []string `river:"span_kinds,attr,optional"`
}

// Convert converts args into the upstream type.
func (args *MatchProperties) Convert() map[string]interface{} {
	if args == nil {
		return nil
	}

	res := make(map[string]interface{})

	res["match_type"] = args.MatchType

	if args.RegexpConfig != nil {
		res["regexp"] = args.RegexpConfig.Convert()
	}

	if len(args.Services) > 0 {
		res["services"] = args.Services
	}

	if len(args.SpanNames) > 0 {
		res["span_names"] = args.SpanNames
	}

	if len(args.LogBodies) > 0 {
		res["log_bodies"] = args.LogBodies
	}

	if len(args.LogSeverityTexts) > 0 {
		res["log_severity_texts"] = args.LogSeverityTexts
	}

	if args.LogSeverityNumber != nil {
		res["log_severity_number"] = args.LogSeverityNumber.Convert()
	}

	if len(args.MetricNames) > 0 {
		res["metric_names"] = args.MetricNames
	}

	if subRes := convertAttributeSlice(args.Attributes); len(subRes) > 0 {
		res["attributes"] = subRes
	}

	if subRes := convertAttributeSlice(args.Resources); len(subRes) > 0 {
		res["resources"] = subRes
	}

	if subRes := convertInstrumentationLibrariesSlice(args.Libraries); len(subRes) > 0 {
		res["libraries"] = subRes
	}

	if len(args.SpanKinds) > 0 {
		res["span_kinds"] = args.SpanKinds
	}

	return res
}

// Return an empty slice if the input slice is empty.
func convertAttributeSlice(attrs []Attribute) []interface{} {
	attrArr := make([]interface{}, 0, len(attrs))
	for _, attr := range attrs {
		attrArr = append(attrArr, attr.Convert())
	}
	return attrArr
}

// Return an empty slice if the input slice is empty.
func convertInstrumentationLibrariesSlice(libs []InstrumentationLibrary) []interface{} {
	libsArr := make([]interface{}, 0, len(libs))
	for _, lib := range libs {
		libsArr = append(libsArr, lib.Convert())
	}
	return libsArr
}

type RegexpConfig struct {
	// CacheEnabled determines whether match results are LRU cached to make subsequent matches faster.
	// Cache size is unlimited unless CacheMaxNumEntries is also specified.
	CacheEnabled bool `river:"cacheenabled,attr,optional"`
	// CacheMaxNumEntries is the max number of entries of the LRU cache that stores match results.
	// CacheMaxNumEntries is ignored if CacheEnabled is false.
	CacheMaxNumEntries int `river:"cachemaxnumentries,attr,optional"`
}

func (args RegexpConfig) Convert() map[string]interface{} {
	return map[string]interface{}{
		"cacheenabled":       args.CacheEnabled,
		"cachemaxnumentries": args.CacheMaxNumEntries,
	}
}

// Attribute specifies the attribute key and optional value to match against.
type Attribute struct {
	// Key specifies the attribute key.
	Key string `river:"key,attr"`

	// Values specifies the value to match against.
	// If it is not set, any value will match.
	Value interface{} `river:"value,attr,optional"`
}

func (args Attribute) Convert() map[string]interface{} {
	return map[string]interface{}{
		"key":   args.Key,
		"value": args.Value,
	}
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

func (args InstrumentationLibrary) Convert() map[string]interface{} {
	return map[string]interface{}{
		"name":    args.Name,
		"version": strings.Clone(*args.Version),
	}
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

func (args LogSeverityNumberMatchProperties) Convert() map[string]interface{} {
	return map[string]interface{}{
		"min":             args.Min,
		"match_undefined": args.MatchUndefined,
	}
}
