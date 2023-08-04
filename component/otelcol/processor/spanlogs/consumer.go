package spanlogs

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/go-logfmt/logfmt"
	otelcomponent "go.opentelemetry.io/collector/component"
	otelconsumer "go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/ptrace"
	semconv "go.opentelemetry.io/collector/semconv/v1.5.0"
)

const (
	typeSpan    = "span"
	typeRoot    = "root"
	typeProcess = "process"
)

type consumer struct {
	optsMut sync.RWMutex
	opts    options
}

type options struct {
	spans             bool
	roots             bool
	processes         bool
	spanAttributes    []string
	processAttributes []string
	overrides         OverrideConfig
	labels            map[string]struct{}
	nextConsumer      otelconsumer.Logs
}

var _ otelconsumer.Traces = (*consumer)(nil)

func NewConsumer(args Arguments, nextConsumer otelconsumer.Logs) (*consumer, error) {
	c := &consumer{}

	c.UpdateOptions(args, nextConsumer)

	return c, nil
}

func (c *consumer) UpdateOptions(args Arguments, nextConsumer otelconsumer.Logs) error {
	c.optsMut.Lock()
	defer c.optsMut.Unlock()

	if nextConsumer == nil {
		return otelcomponent.ErrNilNextConsumer
	}

	labels := make(map[string]struct{}, len(args.Labels))
	for _, l := range args.Labels {
		labels[l] = struct{}{}
	}

	c.opts = options{
		spans:             args.Spans,
		roots:             args.Roots,
		processes:         args.Processes,
		spanAttributes:    args.SpanAttributes,
		processAttributes: args.ProcessAttributes,
		overrides:         args.Overrides,
		labels:            labels,
		nextConsumer:      nextConsumer,
	}

	return nil
}

func (c *consumer) ConsumeTraces(ctx context.Context, td ptrace.Traces) error {
	c.optsMut.RLock()
	defer c.optsMut.RUnlock()

	logs := plog.NewLogs()
	resourceLogs := logs.ResourceLogs()

	rsLen := td.ResourceSpans().Len()
	for i := 0; i < rsLen; i++ {
		resLog := resourceLogs.AppendEmpty()
		scopeLogs := resLog.ScopeLogs()

		rs := td.ResourceSpans().At(i)
		ssLen := rs.ScopeSpans().Len()

		var svc string
		svcAtt, ok := rs.Resource().Attributes().Get(semconv.AttributeServiceName)
		if ok {
			svc = svcAtt.Str()
		}

		for j := 0; j < ssLen; j++ {
			scopeLog := scopeLogs.AppendEmpty()
			logRecords := scopeLog.LogRecords()

			ss := rs.ScopeSpans().At(j)
			spanLen := ss.Spans().Len()

			lastTraceID := ""
			for k := 0; k < spanLen; k++ {
				span := ss.Spans().At(k)
				traceID := span.TraceID().String()

				//TODO: We could reuse the same maps in a lot of this processing. It's very repetitive.
				//TODO: Use a Go map instead of pcommon.NewMap, and convert to pcommon.NewMap only at the very end?
				//      Unfortunately pcommon.NewMap can be inefficient during insertions,
				//      because it traverses the whole map on every insertion.
				if c.opts.spans {
					keyValues := pcommon.NewMap()

					c.spanKeyVals(keyValues, span)
					c.processKeyVals(keyValues, rs.Resource(), svc)

					newLogRecord, err := c.createLogRecord(typeSpan, traceID, keyValues)
					if err != nil {
						return err
					}
					logRecord := logRecords.AppendEmpty()
					newLogRecord.MoveTo(logRecord)
				}

				if c.opts.roots && span.ParentSpanID().IsEmpty() {
					keyValues := pcommon.NewMap()

					c.spanKeyVals(keyValues, span)
					c.processKeyVals(keyValues, rs.Resource(), svc)

					newLogRecord, err := c.createLogRecord(typeRoot, traceID, keyValues)
					if err != nil {
						return err
					}
					logRecord := logRecords.AppendEmpty()
					newLogRecord.MoveTo(logRecord)
				}

				if c.opts.processes && lastTraceID != traceID {
					keyValues := pcommon.NewMap()

					c.processKeyVals(keyValues, rs.Resource(), svc)

					newLogRecord, err := c.createLogRecord(typeProcess, traceID, keyValues)
					if err != nil {
						return err
					}
					logRecord := logRecords.AppendEmpty()
					newLogRecord.MoveTo(logRecord)
				}
			}
		}
	}

	return c.opts.nextConsumer.ConsumeLogs(ctx, logs)
}

func copyOtelMap(out pcommon.Map, in pcommon.Map, pred func(key string, val pcommon.Value) bool) {
	in.Range(func(k string, v pcommon.Value) bool {
		if pred(k, v) {
			newVal := out.PutEmpty(k)
			v.CopyTo(newVal)
		}
		return true
	})
}

func convertOtelMapToSlice(m pcommon.Map) []any {
	res := make([]any, 0, m.Len())

	m.Range(func(k string, v pcommon.Value) bool {
		res = append(res, k)
		res = append(res, v.AsRaw())
		return true
	})

	return res
}

func (c *consumer) createLogRecord(kind string, traceID string, keyValues pcommon.Map) (*plog.LogRecord, error) {
	// Create an empty log record
	res := plog.NewLogRecord()

	// Add a trace ID to the key values
	keyValues.PutStr(c.opts.overrides.TraceIDKey, traceID)

	// Add the log line
	keyValuesSlice := convertOtelMapToSlice(keyValues)
	logLine, err := logfmt.MarshalKeyvals(keyValuesSlice...)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal keyvals due to error: %w", err)
	}

	if logLine != nil {
		res.Body().SetStr(string(logLine))
	}

	// Add the attributes
	logAttributes := res.Attributes()

	// Add logs instance label
	logAttributes.PutStr(c.opts.overrides.LogsTag, kind)

	copyOtelMap(logAttributes, keyValues, func(key string, val pcommon.Value) bool {
		// Check if we have to include this label
		if _, ok := c.opts.labels[key]; ok {
			return true
		}
		return false
	})

	return &res, nil
}

func (c *consumer) processKeyVals(output pcommon.Map, resource pcommon.Resource, svc string) {
	rsAtts := resource.Attributes()

	// name
	output.PutStr(c.opts.overrides.ServiceKey, svc)

	for _, name := range c.opts.processAttributes {
		att, ok := rsAtts.Get(name)
		if ok {
			// name/key val pairs
			val := output.PutEmpty(name)
			att.CopyTo(val)
		}
	}
}

func (c *consumer) spanKeyVals(output pcommon.Map, span ptrace.Span) {
	output.PutStr(c.opts.overrides.SpanNameKey, span.Name())
	output.PutStr(c.opts.overrides.DurationKey, spanDuration(span))

	// Skip STATUS_CODE_UNSET to be less spammy
	if span.Status().Code() != ptrace.StatusCodeUnset {
		output.PutStr(c.opts.overrides.StatusKey, span.Status().Code().String())
	}

	for _, name := range c.opts.spanAttributes {
		att, ok := span.Attributes().Get(name)
		if ok {
			val := output.PutEmpty(name)
			att.CopyTo(val)
		}
	}
}

func spanDuration(span ptrace.Span) string {
	dur := int64(span.EndTimestamp() - span.StartTimestamp())
	return strconv.FormatInt(dur, 10) + "ns"
}

func (c *consumer) Capabilities() otelconsumer.Capabilities {
	return otelconsumer.Capabilities{}
}
