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

package pdata

import (
	"github.com/gogo/protobuf/proto"

	otlplogs "go.opentelemetry.io/collector/internal/data/opentelemetry-proto-gen/logs/v1"
)

// This file defines in-memory data structures to represent logs.

// Logs is the top-level struct that is propagated through the logs pipeline.
//
// This is a reference type (like builtin map).
//
// Must use NewLogs functions to create new instances.
// Important: zero-initialized instance is not valid for use.
type Logs struct {
	orig *[]*otlplogs.ResourceLogs
}

// LogsFromOtlp creates the internal Logs representation from the ProtoBuf.
func LogsFromOtlp(orig []*otlplogs.ResourceLogs) Logs {
	return Logs{&orig}
}

// LogsToOtlp converts the internal Logs to the ProtoBuf.
func LogsToOtlp(ld Logs) []*otlplogs.ResourceLogs {
	return *ld.orig
}

// NewLogs creates a new Logs.
func NewLogs() Logs {
	orig := []*otlplogs.ResourceLogs(nil)
	return Logs{&orig}
}

// Clone returns a copy of Logs.
func (ld Logs) Clone() Logs {
	otlp := LogsToOtlp(ld)
	resourceSpansClones := make([]*otlplogs.ResourceLogs, 0, len(otlp))
	for _, resourceSpans := range otlp {
		resourceSpansClones = append(resourceSpansClones,
			proto.Clone(resourceSpans).(*otlplogs.ResourceLogs))
	}
	return LogsFromOtlp(resourceSpansClones)
}

// LogRecordCount calculates the total number of log records.
func (ld Logs) LogRecordCount() int {
	logCount := 0
	rss := ld.ResourceLogs()
	for i := 0; i < rss.Len(); i++ {
		rs := rss.At(i)
		if rs.IsNil() {
			continue
		}

		ill := rs.InstrumentationLibraryLogs()
		for i := 0; i < ill.Len(); i++ {
			logs := ill.At(i)
			if logs.IsNil() {
				continue
			}
			logCount += logs.Logs().Len()
		}
	}
	return logCount
}

func (ld Logs) ResourceLogs() ResourceLogsSlice {
	return ResourceLogsSlice(ld)
}

// SeverityNumber is the public alias of otlplogs.SeverityNumber from internal package.
type SeverityNumber otlplogs.SeverityNumber

const (
	SeverityNumberUNDEFINED = SeverityNumber(otlplogs.SeverityNumber_UNDEFINED_SEVERITY_NUMBER)
	SeverityNumberTRACE     = SeverityNumber(otlplogs.SeverityNumber_TRACE)
	SeverityNumberTRACE2    = SeverityNumber(otlplogs.SeverityNumber_TRACE2)
	SeverityNumberTRACE3    = SeverityNumber(otlplogs.SeverityNumber_TRACE3)
	SeverityNumberTRACE4    = SeverityNumber(otlplogs.SeverityNumber_TRACE4)
	SeverityNumberDEBUG     = SeverityNumber(otlplogs.SeverityNumber_DEBUG)
	SeverityNumberDEBUG2    = SeverityNumber(otlplogs.SeverityNumber_DEBUG2)
	SeverityNumberDEBUG3    = SeverityNumber(otlplogs.SeverityNumber_DEBUG3)
	SeverityNumberDEBUG4    = SeverityNumber(otlplogs.SeverityNumber_DEBUG4)
	SeverityNumberINFO      = SeverityNumber(otlplogs.SeverityNumber_INFO)
	SeverityNumberINFO2     = SeverityNumber(otlplogs.SeverityNumber_INFO2)
	SeverityNumberINFO3     = SeverityNumber(otlplogs.SeverityNumber_INFO3)
	SeverityNumberINFO4     = SeverityNumber(otlplogs.SeverityNumber_INFO4)
	SeverityNumberWARN      = SeverityNumber(otlplogs.SeverityNumber_WARN)
	SeverityNumberWARN2     = SeverityNumber(otlplogs.SeverityNumber_WARN2)
	SeverityNumberWARN3     = SeverityNumber(otlplogs.SeverityNumber_WARN3)
	SeverityNumberWARN4     = SeverityNumber(otlplogs.SeverityNumber_WARN4)
	SeverityNumberERROR     = SeverityNumber(otlplogs.SeverityNumber_ERROR)
	SeverityNumberERROR2    = SeverityNumber(otlplogs.SeverityNumber_ERROR2)
	SeverityNumberERROR3    = SeverityNumber(otlplogs.SeverityNumber_ERROR3)
	SeverityNumberERROR4    = SeverityNumber(otlplogs.SeverityNumber_ERROR4)
	SeverityNumberFATAL     = SeverityNumber(otlplogs.SeverityNumber_FATAL)
	SeverityNumberFATAL2    = SeverityNumber(otlplogs.SeverityNumber_FATAL2)
	SeverityNumberFATAL3    = SeverityNumber(otlplogs.SeverityNumber_FATAL3)
	SeverityNumberFATAL4    = SeverityNumber(otlplogs.SeverityNumber_FATAL4)
)
