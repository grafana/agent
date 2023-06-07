// Package convert implements conversion utilities to convert between
// OpenTelemetry Collector and Loki data.
//
// It follows the [OpenTelemetry Logs Data Model] and the [loki translator]
// package for implementing the conversion.
//
// [OpenTelemetry Logs Data Model]: https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/logs/data-model.md
// [loki translator]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/pkg/translator/loki
package convert

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component/common/loki"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
)

// Converter implements consumer.Logs and converts received OTel logs into
// Loki-compatible log entries.
type Converter struct {
	log     log.Logger
	metrics *metrics

	mut  sync.RWMutex
	next []loki.LogsReceiver // Location to write converted logs.
}

var _ consumer.Logs = (*Converter)(nil)

// New returns a new Converter. Converted logs are passed to the provided list
// of LogsReceivers.
func New(l log.Logger, r prometheus.Registerer, next []loki.LogsReceiver) *Converter {
	if l == nil {
		l = log.NewNopLogger()
	}
	m := newMetrics(r)
	return &Converter{log: l, metrics: m, next: next}
}

// Capabilities implements consumer.Logs.
func (conv *Converter) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{
		MutatesData: false,
	}
}

// ConsumeLogs converts the provided OpenTelemetry Collector-formatted logs
// into Loki-compatible entries. Each call to ConsumeLogs will forward
// converted entries to the list of channels in the `next` field.
// This is reusing the logic from the OpenTelemetry Collector "contrib"
// distribution and its LogsToLokiRequests function.
func (conv *Converter) ConsumeLogs(ctx context.Context, ld plog.Logs) error {
	var entries []loki.Entry

	rls := ld.ResourceLogs()
	for i := 0; i < rls.Len(); i++ {
		ills := rls.At(i).ScopeLogs()
		for j := 0; j < ills.Len(); j++ {
			logs := ills.At(j).LogRecords()
			scope := ills.At(j).Scope()
			for k := 0; k < logs.Len(); k++ {
				conv.metrics.entriesTotal.Inc()

				// we may remove attributes, so to avoid mutating the original
				// log entry, we make our own copy and change that instead.
				log := plog.NewLogRecord()
				logs.At(k).CopyTo(log)

				// similarly, we may remove resources, so to avoid mutating the
				// original log entry, we make and use our own copy instead.
				resource := pcommon.NewResource()
				rls.At(i).Resource().CopyTo(resource)

				// adds level attribute from log.severityNumber
				addLogLevelAttributeAndHint(log)

				// TODO (@tpaschalis) If we want to pre-populate a tenant
				// label from the OTel hint, it should happen here. with the
				// upstream getTenantFromTenantHint helper.

				format := getFormatFromFormatHint(log.Attributes(), resource.Attributes())

				mergedLabels := convertAttributesAndMerge(log.Attributes(), resource.Attributes())
				// remove the attributes that were promoted to labels
				removeAttributes(log.Attributes(), mergedLabels)
				removeAttributes(resource.Attributes(), mergedLabels)

				entry, err := convertLogToLokiEntry(log, resource, format, scope)
				if err != nil {
					level.Error(conv.log).Log("msg", "failed to convert log to loki entry", "err", err)
					conv.metrics.entriesFailed.Inc()
					continue
				}

				conv.metrics.entriesProcessed.Inc()
				entries = append(entries, loki.Entry{
					Labels: mergedLabels,
					Entry:  *entry,
				})
			}
		}
	}

	for _, entry := range entries {
		conv.mut.RLock()
		for _, receiver := range conv.next {
			select {
			case <-ctx.Done():
				return nil
			case receiver <- entry:
				// no-op, send the entry along
			}
		}
		conv.mut.RUnlock()
	}
	return nil
}

// UpdateFanout sets the locations the converter forwards log entries to.
func (conv *Converter) UpdateFanout(fanout []loki.LogsReceiver) {
	conv.mut.Lock()
	defer conv.mut.Unlock()

	conv.next = fanout
}

func addLogLevelAttributeAndHint(log plog.LogRecord) {
	if log.SeverityNumber() == plog.SeverityNumberUnspecified {
		return
	}
	addHint(log)
	if _, found := log.Attributes().Get(levelAttributeName); !found {
		level := severityNumberToLevel[log.SeverityNumber().String()]
		log.Attributes().PutStr(levelAttributeName, level)
	}
}

func addHint(log plog.LogRecord) {
	if value, found := log.Attributes().Get(hintAttributes); found && !strings.Contains(value.AsString(), levelAttributeName) {
		log.Attributes().PutStr(hintAttributes, fmt.Sprintf("%s,%s", value.AsString(), levelAttributeName))
	} else {
		log.Attributes().PutStr(hintAttributes, levelAttributeName)
	}
}

var severityNumberToLevel = map[string]string{
	plog.SeverityNumberUnspecified.String(): "UNSPECIFIED",
	plog.SeverityNumberTrace.String():       "TRACE",
	plog.SeverityNumberTrace2.String():      "TRACE2",
	plog.SeverityNumberTrace3.String():      "TRACE3",
	plog.SeverityNumberTrace4.String():      "TRACE4",
	plog.SeverityNumberDebug.String():       "DEBUG",
	plog.SeverityNumberDebug2.String():      "DEBUG2",
	plog.SeverityNumberDebug3.String():      "DEBUG3",
	plog.SeverityNumberDebug4.String():      "DEBUG4",
	plog.SeverityNumberInfo.String():        "INFO",
	plog.SeverityNumberInfo2.String():       "INFO2",
	plog.SeverityNumberInfo3.String():       "INFO3",
	plog.SeverityNumberInfo4.String():       "INFO4",
	plog.SeverityNumberWarn.String():        "WARN",
	plog.SeverityNumberWarn2.String():       "WARN2",
	plog.SeverityNumberWarn3.String():       "WARN3",
	plog.SeverityNumberWarn4.String():       "WARN4",
	plog.SeverityNumberError.String():       "ERROR",
	plog.SeverityNumberError2.String():      "ERROR2",
	plog.SeverityNumberError3.String():      "ERROR3",
	plog.SeverityNumberError4.String():      "ERROR4",
	plog.SeverityNumberFatal.String():       "FATAL",
	plog.SeverityNumberFatal2.String():      "FATAL2",
	plog.SeverityNumberFatal3.String():      "FATAL3",
	plog.SeverityNumberFatal4.String():      "FATAL4",
}

func getFormatFromFormatHint(logAttr pcommon.Map, resourceAttr pcommon.Map) string {
	format := formatJSON
	formatVal, found := resourceAttr.Get(hintFormat)
	if !found {
		formatVal, found = logAttr.Get(hintFormat)
	}

	if found {
		format = formatVal.AsString()
	}
	return format
}
