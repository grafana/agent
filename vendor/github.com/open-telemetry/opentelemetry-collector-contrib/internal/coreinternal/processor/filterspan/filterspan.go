// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package filterspan // import "github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal/processor/filterspan"

import (
	"fmt"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
	conventions "go.opentelemetry.io/collector/semconv/v1.6.1"

	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal/processor/filterconfig"
	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal/processor/filtermatcher"
	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal/processor/filterset"
)

// Matcher is an interface that allows matching a span against a configuration
// of a match.
// TODO: Modify Matcher to invoke both the include and exclude properties so
//
//	calling processors will always have the same logic.
type Matcher interface {
	MatchSpan(span ptrace.Span, resource pcommon.Resource, library pcommon.InstrumentationScope) bool
}

// propertiesMatcher allows matching a span against various span properties.
type propertiesMatcher struct {
	filtermatcher.PropertiesMatcher

	// Service names to compare to.
	serviceFilters filterset.FilterSet

	// Span names to compare to.
	nameFilters filterset.FilterSet

	// Span kinds to compare to
	kindFilters filterset.FilterSet
}

// NewMatcher creates a span Matcher that matches based on the given MatchProperties.
func NewMatcher(mp *filterconfig.MatchProperties) (Matcher, error) {
	if mp == nil {
		return nil, nil
	}

	if err := mp.ValidateForSpans(); err != nil {
		return nil, err
	}

	rm, err := filtermatcher.NewMatcher(mp)
	if err != nil {
		return nil, err
	}

	var serviceFS filterset.FilterSet
	if len(mp.Services) > 0 {
		serviceFS, err = filterset.CreateFilterSet(mp.Services, &mp.Config)
		if err != nil {
			return nil, fmt.Errorf("error creating service name filters: %w", err)
		}
	}

	var nameFS filterset.FilterSet
	if len(mp.SpanNames) > 0 {
		nameFS, err = filterset.CreateFilterSet(mp.SpanNames, &mp.Config)
		if err != nil {
			return nil, fmt.Errorf("error creating span name filters: %w", err)
		}
	}

	var kindFS filterset.FilterSet
	if len(mp.SpanKinds) > 0 {
		kindFS, err = filterset.CreateFilterSet(mp.SpanKinds, &mp.Config)
		if err != nil {
			return nil, fmt.Errorf("error creating span kind filters: %w", err)
		}
	}

	return &propertiesMatcher{
		PropertiesMatcher: rm,
		serviceFilters:    serviceFS,
		nameFilters:       nameFS,
		kindFilters:       kindFS,
	}, nil
}

// SkipSpan determines if a span should be processed.
// True is returned when a span should be skipped.
// False is returned when a span should not be skipped.
// The logic determining if a span should be processed is set
// in the attribute configuration with the include and exclude settings.
// Include properties are checked before exclude settings are checked.
func SkipSpan(include Matcher, exclude Matcher, span ptrace.Span, resource pcommon.Resource, library pcommon.InstrumentationScope) bool {
	if include != nil {
		// A false returned in this case means the span should not be processed.
		if i := include.MatchSpan(span, resource, library); !i {
			return true
		}
	}

	if exclude != nil {
		// A true returned in this case means the span should not be processed.
		if e := exclude.MatchSpan(span, resource, library); e {
			return true
		}
	}

	return false
}

// MatchSpan matches a span and service to a set of properties.
// see filterconfig.MatchProperties for more details
func (mp *propertiesMatcher) MatchSpan(span ptrace.Span, resource pcommon.Resource, library pcommon.InstrumentationScope) bool {
	// If a set of properties was not in the mp, all spans are considered to match on that property
	if mp.serviceFilters != nil {
		// Check resource and spans for service.name
		serviceName := serviceNameForResource(resource)

		if !mp.serviceFilters.Matches(serviceName) {
			return false
		}
	}

	if mp.nameFilters != nil && !mp.nameFilters.Matches(span.Name()) {
		return false
	}

	if mp.kindFilters != nil && !mp.kindFilters.Matches(span.Kind().String()) {
		return false
	}

	return mp.PropertiesMatcher.Match(span.Attributes(), resource, library)
}

// serviceNameForResource gets the service name for a specified Resource.
func serviceNameForResource(resource pcommon.Resource) string {
	service, found := resource.Attributes().Get(conventions.AttributeServiceName)
	if !found {
		return "<nil-service-name>"
	}
	return service.AsString()
}
