package spanlogs

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
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

	logger log.Logger
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

func NewConsumer(args Arguments, nextConsumer otelconsumer.Logs, logger log.Logger) (*consumer, error) {
	c := &consumer{
		logger: logger,
	}

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

				if c.opts.spans {
					keyValues := append(c.spanKeyVals(span), c.processKeyVals(rs.Resource(), svc)...)

					newLogRecord := c.createLogRecord(typeSpan, traceID, keyValues)
					if newLogRecord == nil {
						//TODO: Make a meaningful error
						return fmt.Errorf("")
					}
					logRecord := logRecords.AppendEmpty()
					newLogRecord.MoveTo(logRecord)
				}

				if c.opts.roots && span.ParentSpanID().IsEmpty() {
					keyValues := append(c.spanKeyVals(span), c.processKeyVals(rs.Resource(), svc)...)

					newLogRecord := c.createLogRecord(typeRoot, traceID, keyValues)
					if newLogRecord == nil {
						//TODO: Make a meaningful error
						return fmt.Errorf("")
					}
					logRecord := logRecords.AppendEmpty()
					newLogRecord.MoveTo(logRecord)
				}

				if c.opts.processes && lastTraceID != traceID {
					lastTraceID = traceID
					keyValues := c.processKeyVals(rs.Resource(), svc)

					newLogRecord := c.createLogRecord(typeProcess, traceID, keyValues)
					if newLogRecord == nil {
						//TODO: Make a meaningful error
						return fmt.Errorf("")
					}
					logRecord := logRecords.AppendEmpty()
					newLogRecord.MoveTo(logRecord)
				}
			}
		}
	}

	//TODO: If the log records are empty, should we send anything downstream?
	return c.opts.nextConsumer.ConsumeLogs(ctx, logs)
}

func (c *consumer) createLogRecord(kind string, traceID string, keyValues []interface{}) *plog.LogRecord {
	// Create an empty log record
	res := plog.NewLogRecord()

	// Add the log line
	keyValues = append(keyValues, []interface{}{c.opts.overrides.TraceIDKey, traceID}...)
	logLine, err := logfmt.MarshalKeyvals(keyValues...)
	if err != nil {
		level.Warn(c.logger).Log("msg", "unable to marshal keyvals", "err", err)
		return nil
	}
	if logLine != nil {
		res.Body().SetStr(string(logLine))
	}

	// Add the attributes
	logAttributes := res.Attributes()

	// Add logs instance label
	logAttributes.PutStr(c.opts.overrides.LogsTag, kind)

	var (
		k  string
		ok bool
	)
	for i := 0; i < len(keyValues); i += 2 {
		if k, ok = keyValues[i].(string); !ok {
			// Should never happen, all keys are strings
			level.Error(c.logger).Log("msg", "error casting label key to string", "key", keyValues[i])
			continue
		}

		// Check if we have to include this label
		if _, ok := c.opts.labels[k]; !ok {
			continue
		}

		//TODO: Can we make this more accurate?
		switch v := keyValues[i+1].(type) {
		case int64:
			logAttributes.PutInt(k, v)
		case bool:
			logAttributes.PutBool(k, v)
		case float64:
			logAttributes.PutDouble(k, v)
		default:
			var val pcommon.Value
			val.FromRaw(v)
			logAttributes.PutStr(k, val.AsString())
		}
	}

	return &res
}

func (c *consumer) processKeyVals(resource pcommon.Resource, svc string) []interface{} {
	atts := make([]interface{}, 0, 2) // 2 for service name
	rsAtts := resource.Attributes()

	// name
	atts = append(atts, c.opts.overrides.ServiceKey)
	atts = append(atts, svc)

	for _, name := range c.opts.processAttributes {
		att, ok := rsAtts.Get(name)
		if ok {
			// name/key val pairs
			atts = append(atts, name)
			atts = append(atts, att.AsRaw())
		}
	}

	return atts
}

func (c *consumer) spanKeyVals(span ptrace.Span) []interface{} {
	atts := make([]interface{}, 0, 8) // 8 for name, duration, service name and status

	atts = append(atts, c.opts.overrides.SpanNameKey)
	atts = append(atts, span.Name())

	atts = append(atts, c.opts.overrides.DurationKey)
	atts = append(atts, spanDuration(span))

	// Skip STATUS_CODE_UNSET to be less spammy
	if span.Status().Code() != ptrace.StatusCodeUnset {
		atts = append(atts, c.opts.overrides.StatusKey)
		atts = append(atts, span.Status().Code())
	}

	for _, name := range c.opts.spanAttributes {
		att, ok := span.Attributes().Get(name)
		if ok {
			atts = append(atts, name)
			atts = append(atts, att.AsRaw())
		}
	}

	return atts
}

func spanDuration(span ptrace.Span) string {
	dur := int64(span.EndTimestamp() - span.StartTimestamp())
	return strconv.FormatInt(dur, 10) + "ns"
}

func (c *consumer) Capabilities() otelconsumer.Capabilities {
	return otelconsumer.Capabilities{}
}
