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

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal/attraction"
	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal/processor/filterlog"
)

type logAttributesProcessor struct {
	logger   *zap.Logger
	attrProc *attraction.AttrProc
	include  filterlog.Matcher
	exclude  filterlog.Matcher
}

// newLogAttributesProcessor returns a processor that modifies attributes of a
// log record. To construct the attributes processors, the use of the factory
// methods are required in order to validate the inputs.
func newLogAttributesProcessor(logger *zap.Logger, attrProc *attraction.AttrProc, include, exclude filterlog.Matcher) *logAttributesProcessor {
	return &logAttributesProcessor{
		logger:   logger,
		attrProc: attrProc,
		include:  include,
		exclude:  exclude,
	}
}

func (a *logAttributesProcessor) processLogs(ctx context.Context, ld plog.Logs) (plog.Logs, error) {
	rls := ld.ResourceLogs()
	for i := 0; i < rls.Len(); i++ {
		rs := rls.At(i)
		ilss := rs.ScopeLogs()
		resource := rs.Resource()
		for j := 0; j < ilss.Len(); j++ {
			ils := ilss.At(j)
			logs := ils.LogRecords()
			library := ils.Scope()
			for k := 0; k < logs.Len(); k++ {
				lr := logs.At(k)
				if a.skipLog(lr, resource, library) {
					continue
				}

				a.attrProc.Process(ctx, a.logger, lr.Attributes())
			}
		}
	}
	return ld, nil
}

// skipLog determines if a log should be processed.
// True is returned when a log should be skipped.
// False is returned when a log should not be skipped.
// The logic determining if a log should be processed is set
// in the attribute configuration with the include and exclude settings.
// Include properties are checked before exclude settings are checked.
func (a *logAttributesProcessor) skipLog(lr plog.LogRecord, resource pcommon.Resource, library pcommon.InstrumentationScope) bool {
	if a.include != nil {
		// A false returned in this case means the log should not be processed.
		if include := a.include.MatchLogRecord(lr, resource, library); !include {
			return true
		}
	}

	if a.exclude != nil {
		// A true returned in this case means the log should not be processed.
		if exclude := a.exclude.MatchLogRecord(lr, resource, library); exclude {
			return true
		}
	}

	return false
}
