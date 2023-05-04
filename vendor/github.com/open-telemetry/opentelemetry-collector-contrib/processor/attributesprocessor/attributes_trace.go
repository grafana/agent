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

package attributesprocessor // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/attributesprocessor"

import (
	"context"

	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal/attraction"
	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal/processor/filterspan"
)

type spanAttributesProcessor struct {
	logger   *zap.Logger
	attrProc *attraction.AttrProc
	include  filterspan.Matcher
	exclude  filterspan.Matcher
}

// newTracesProcessor returns a processor that modifies attributes of a span.
// To construct the attributes processors, the use of the factory methods are required
// in order to validate the inputs.
func newSpanAttributesProcessor(logger *zap.Logger, attrProc *attraction.AttrProc, include, exclude filterspan.Matcher) *spanAttributesProcessor {
	return &spanAttributesProcessor{
		logger:   logger,
		attrProc: attrProc,
		include:  include,
		exclude:  exclude,
	}
}

func (a *spanAttributesProcessor) processTraces(ctx context.Context, td ptrace.Traces) (ptrace.Traces, error) {
	rss := td.ResourceSpans()
	for i := 0; i < rss.Len(); i++ {
		rs := rss.At(i)
		resource := rs.Resource()
		ilss := rs.ScopeSpans()
		for j := 0; j < ilss.Len(); j++ {
			ils := ilss.At(j)
			spans := ils.Spans()
			library := ils.Scope()
			for k := 0; k < spans.Len(); k++ {
				span := spans.At(k)
				if filterspan.SkipSpan(a.include, a.exclude, span, resource, library) {
					continue
				}

				a.attrProc.Process(ctx, a.logger, span.Attributes())
			}
		}
	}
	return td, nil
}
