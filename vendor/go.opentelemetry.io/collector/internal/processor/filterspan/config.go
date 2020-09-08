// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package filterspan

import (
	"go.opentelemetry.io/collector/internal/processor/filterset"
)

// MatchConfig has two optional MatchProperties one to define what is processed
// by the processor, captured under the 'include' and the second, exclude, to
// define what is excluded from the processor.
type MatchConfig struct {
	// Include specifies the set of span properties that must be present in order
	// for this processor to apply to it.
	// Note: If `exclude` is specified, the span is compared against those
	// properties after the `include` properties.
	// This is an optional field. If neither `include` and `exclude` are set, all spans
	// are processed. If `include` is set and `exclude` isn't set, then all
	// spans matching the properties in this structure are processed.
	Include *MatchProperties `mapstructure:"include"`

	// Exclude specifies when this processor will not be applied to the Spans
	// which match the specified properties.
	// Note: The `exclude` properties are checked after the `include` properties,
	// if they exist, are checked.
	// If `include` isn't specified, the `exclude` properties are checked against
	// all spans.
	// This is an optional field. If neither `include` and `exclude` are set, all spans
	// are processed. If `exclude` is set and `include` isn't set, then all
	// spans  that do no match the properties in this structure are processed.
	Exclude *MatchProperties `mapstructure:"exclude"`
}

// MatchProperties specifies the set of properties in a span to match against
// and if the span should be included or excluded from the processor.
// At least one of services, span names or attributes must be specified. It is
// supported to have all specified, but this requires all of the properties to
// match for the inclusion/exclusion to occur.
// The following are examples of invalid configurations:
//  attributes/bad1:
//    # This is invalid because include is specified with neither services or
//    # attributes.
//    include:
//    actions: ...
//
//  span/bad2:
//    exclude:
//    	# This is invalid because services, span_names and attributes have empty values.
//      services:
//      span_names:
//      attributes:
//    actions: ...
// Please refer to processor/attributesprocessor/testdata/config.yaml and
// processor/spanprocessor/testdata/config.yaml for valid configurations.
type MatchProperties struct {
	// Config configures the matching patterns used when matching span properties.
	filterset.Config `mapstructure:",squash"`

	// Note: one of Services, SpanNames or Attributes must be specified with a
	// non-empty value for a valid configuration.

	// Services specify the list of of items to match service name against.
	// A match occurs if the span's service name matches at least one item in this list.
	// This is an optional field.
	Services []string `mapstructure:"services"`

	// SpanNames specify the list of items to match span name against.
	// A match occurs if the span name matches at least one item in this list.
	// This is an optional field.
	SpanNames []string `mapstructure:"span_names"`

	// Attributes specifies the list of attributes to match against.
	// All of these attributes must match exactly for a match to occur.
	// Only match_type=strict is allowed if "attributes" are specified.
	// This is an optional field.
	Attributes []Attribute `mapstructure:"attributes"`
}

// MatchTypeFieldName is the mapstructure field name for MatchProperties.Attributes field.
const AttributesFieldName = "attributes"

// Attribute specifies the attribute key and optional value to match against.
type Attribute struct {
	// Key specifies the attribute key.
	Key string `mapstructure:"key"`

	// Values specifies the value to match against.
	// If it is not set, any value will match.
	Value interface{} `mapstructure:"value"`
}
