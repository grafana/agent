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

package attributesprocessor

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenterror"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/consumer/pdata"
	"go.opentelemetry.io/collector/internal/processor/attraction"
	"go.opentelemetry.io/collector/internal/processor/filterspan"
	"go.opentelemetry.io/collector/processor"
)

type attributesProcessor struct {
	nextConsumer consumer.TraceConsumer
	attrProc     *attraction.AttrProc
	include      filterspan.Matcher
	exclude      filterspan.Matcher
}

// newTraceProcessor returns a processor that modifies attributes of a span.
// To construct the attributes processors, the use of the factory methods are required
// in order to validate the inputs.
func newTraceProcessor(nextConsumer consumer.TraceConsumer, attrProc *attraction.AttrProc, include, exclude filterspan.Matcher) (component.TraceProcessor, error) {
	if nextConsumer == nil {
		return nil, componenterror.ErrNilNextConsumer
	}
	ap := &attributesProcessor{
		nextConsumer: nextConsumer,
		attrProc:     attrProc,
		include:      include,
		exclude:      exclude,
	}
	return ap, nil
}

func (a *attributesProcessor) ConsumeTraces(ctx context.Context, td pdata.Traces) error {
	rss := td.ResourceSpans()
	for i := 0; i < rss.Len(); i++ {
		rs := rss.At(i)
		if rs.IsNil() {
			continue
		}
		serviceName := processor.ServiceNameForResource(rs.Resource())
		ilss := rss.At(i).InstrumentationLibrarySpans()
		for j := 0; j < ilss.Len(); j++ {
			ils := ilss.At(j)
			if ils.IsNil() {
				continue
			}
			spans := ils.Spans()
			for k := 0; k < spans.Len(); k++ {
				span := spans.At(k)
				if span.IsNil() {
					// Do not create empty spans just to add attributes
					continue
				}

				if a.skipSpan(span, serviceName) {
					continue
				}

				a.attrProc.Process(span.Attributes())
			}
		}
	}
	return a.nextConsumer.ConsumeTraces(ctx, td)
}

func (a *attributesProcessor) GetCapabilities() component.ProcessorCapabilities {
	return component.ProcessorCapabilities{MutatesConsumedData: true}
}

// Start is invoked during service startup.
func (a *attributesProcessor) Start(_ context.Context, _ component.Host) error {
	return nil
}

// Shutdown is invoked during service shutdown.
func (a *attributesProcessor) Shutdown(context.Context) error {
	return nil
}

// skipSpan determines if a span should be processed.
// True is returned when a span should be skipped.
// False is returned when a span should not be skipped.
// The logic determining if a span should be processed is set
// in the attribute configuration with the include and exclude settings.
// Include properties are checked before exclude settings are checked.
func (a *attributesProcessor) skipSpan(span pdata.Span, serviceName string) bool {
	if a.include != nil {
		// A false returned in this case means the span should not be processed.
		if include := a.include.MatchSpan(span, serviceName); !include {
			return true
		}
	}

	if a.exclude != nil {
		// A true returned in this case means the span should not be processed.
		if exclude := a.exclude.MatchSpan(span, serviceName); exclude {
			return true
		}
	}

	return false
}
