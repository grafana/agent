package automaticloggingprocessor

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	util "github.com/cortexproject/cortex/pkg/util/log"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/go-logfmt/logfmt"
	"github.com/grafana/agent/pkg/logs"
	"github.com/grafana/agent/pkg/tempo/contextkeys"
	"github.com/grafana/loki/clients/pkg/promtail/api"
	"github.com/grafana/loki/pkg/logproto"
	"github.com/prometheus/common/model"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenterror"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/consumer/pdata"
	"go.opentelemetry.io/collector/translator/conventions"
	"go.uber.org/atomic"
)

const (
	defaultLogsTag     = "tempo"
	defaultServiceKey  = "svc"
	defaultSpanNameKey = "span"
	defaultStatusKey   = "status"
	defaultDurationKey = "dur"
	defaultTraceIDKey  = "tid"

	defaultTimeout = time.Millisecond

	typeSpan    = "span"
	typeRoot    = "root"
	typeProcess = "process"
)

type automaticLoggingProcessor struct {
	nextConsumer consumer.Traces

	cfg          *AutomaticLoggingConfig
	logToStdout  bool
	logsInstance *logs.Instance
	done         atomic.Bool

	labels map[string]bool

	logger log.Logger
}

func newTraceProcessor(nextConsumer consumer.Traces, cfg *AutomaticLoggingConfig) (component.TracesProcessor, error) {
	logger := log.With(util.Logger, "component", "tempo automatic logging")

	if nextConsumer == nil {
		return nil, componenterror.ErrNilNextConsumer
	}

	if !cfg.Roots && !cfg.Processes && !cfg.Spans {
		return nil, errors.New("automaticLoggingProcessor requires one of roots, processes, or spans to be enabled")
	}

	if cfg.Timeout == 0 {
		cfg.Timeout = defaultTimeout
	}

	if cfg.Backend == "" {
		cfg.Backend = BackendStdout
	}

	if cfg.Backend != BackendLogs && cfg.Backend != BackendStdout {
		return nil, fmt.Errorf("automaticLoggingProcessor requires a backend of type '%s' or '%s'", BackendLogs, BackendStdout)
	}

	logToStdout := false
	if cfg.Backend == BackendStdout {
		logToStdout = true
	}

	cfg.Overrides.LogsTag = override(cfg.Overrides.LogsTag, defaultLogsTag)
	cfg.Overrides.ServiceKey = override(cfg.Overrides.ServiceKey, defaultServiceKey)
	cfg.Overrides.SpanNameKey = override(cfg.Overrides.SpanNameKey, defaultSpanNameKey)
	cfg.Overrides.StatusKey = override(cfg.Overrides.StatusKey, defaultStatusKey)
	cfg.Overrides.DurationKey = override(cfg.Overrides.DurationKey, defaultDurationKey)
	cfg.Overrides.TraceIDKey = override(cfg.Overrides.TraceIDKey, defaultTraceIDKey)

	labels := make(map[string]bool, len(cfg.Labels))
	for _, l := range cfg.Labels {
		labels[l] = true
	}

	return &automaticLoggingProcessor{
		nextConsumer: nextConsumer,
		cfg:          cfg,
		logToStdout:  logToStdout,
		logger:       logger,
		done:         atomic.Bool{},
		labels:       labels,
	}, nil
}

func (p *automaticLoggingProcessor) ConsumeTraces(ctx context.Context, td pdata.Traces) error {
	rsLen := td.ResourceSpans().Len()
	for i := 0; i < rsLen; i++ {
		rs := td.ResourceSpans().At(i)
		ilsLen := rs.InstrumentationLibrarySpans().Len()

		var svc string
		svcAtt, ok := rs.Resource().Attributes().Get(conventions.AttributeServiceName)
		if ok {
			svc = svcAtt.StringVal()
		}

		for j := 0; j < ilsLen; j++ {
			ils := rs.InstrumentationLibrarySpans().At(j)
			spanLen := ils.Spans().Len()

			lastTraceID := ""
			for k := 0; k < spanLen; k++ {
				span := ils.Spans().At(k)
				traceID := span.TraceID().HexString()

				if p.cfg.Spans {
					keyValues := append(p.spanKeyVals(span), p.processKeyVals(rs.Resource(), svc)...)
					p.exportToLogsInstance(typeSpan, traceID, p.spanLabels(keyValues), keyValues...)
				}

				if p.cfg.Roots && span.ParentSpanID().IsEmpty() {
					keyValues := append(p.spanKeyVals(span), p.processKeyVals(rs.Resource(), svc)...)
					p.exportToLogsInstance(typeSpan, traceID, p.spanLabels(keyValues), keyValues...)
				}

				if p.cfg.Processes && lastTraceID != traceID {
					lastTraceID = traceID
					keyValues := p.processKeyVals(rs.Resource(), svc)
					p.exportToLogsInstance(typeProcess, traceID, p.spanLabels(keyValues), keyValues...)
				}
			}
		}
	}

	return p.nextConsumer.ConsumeTraces(ctx, td)
}

func (p *automaticLoggingProcessor) spanLabels(keyValues []interface{}) model.LabelSet {
	if len(keyValues) == 0 {
		return nil
	}
	ls := make(map[model.LabelName]model.LabelValue, len(keyValues))
	var (
		k, v string
		ok   bool
	)
	for i := 0; i < len(keyValues); i += 2 {
		if k, ok = keyValues[i].(string); !ok {
			// Should never happen, all keys are strings
			level.Error(p.logger).Log("msg", "error casting label key to string", "key", keyValues[i])
			continue
		}
		// Try to cast value to string
		if v, ok = keyValues[i+1].(string); !ok {
			// If it's not a string, format it to its string representation
			v = fmt.Sprintf("%v", keyValues[i+1])
		}
		if p.labels[k] {
			ls[model.LabelName(k)] = model.LabelValue(v)
		}
	}
	return ls
}

func (p *automaticLoggingProcessor) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{}
}

// Start is invoked during service startup.
func (p *automaticLoggingProcessor) Start(ctx context.Context, _ component.Host) error {
	logs := ctx.Value(contextkeys.Logs).(*logs.Logs)
	if logs == nil {
		return fmt.Errorf("key does not contain a logs instance")
	}

	if !p.logToStdout {
		p.logsInstance = logs.Instance(p.cfg.LogsName)
		if p.logsInstance == nil {
			return fmt.Errorf("logs instance %s not found", p.cfg.LogsName)
		}
	}
	return nil
}

// Shutdown is invoked during service shutdown.
func (p *automaticLoggingProcessor) Shutdown(context.Context) error {
	p.done.Store(true)

	return nil
}

func (p *automaticLoggingProcessor) processKeyVals(resource pdata.Resource, svc string) []interface{} {
	atts := make([]interface{}, 0, 2) // 2 for service name
	rsAtts := resource.Attributes()

	// name
	atts = append(atts, p.cfg.Overrides.ServiceKey)
	atts = append(atts, svc)

	for _, name := range p.cfg.ProcessAttributes {
		att, ok := rsAtts.Get(name)
		if ok {
			// name/key val pairs
			atts = append(atts, name)
			atts = append(atts, attributeValue(att))
		}
	}

	return atts
}

func (p *automaticLoggingProcessor) spanKeyVals(span pdata.Span) []interface{} {
	atts := make([]interface{}, 0, 8) // 8 for name, duration, service name and status

	atts = append(atts, p.cfg.Overrides.SpanNameKey)
	atts = append(atts, span.Name())

	atts = append(atts, p.cfg.Overrides.DurationKey)
	atts = append(atts, spanDuration(span))

	atts = append(atts, p.cfg.Overrides.StatusKey)
	atts = append(atts, span.Status().Code())

	for _, name := range p.cfg.SpanAttributes {
		att, ok := span.Attributes().Get(name)
		if ok {
			atts = append(atts, name)
			atts = append(atts, attributeValue(att))
		}
	}

	return atts
}

func (p *automaticLoggingProcessor) exportToLogsInstance(kind string, traceID string, labels model.LabelSet, keyvals ...interface{}) {
	if p.done.Load() {
		return
	}

	keyvals = append(keyvals, []interface{}{p.cfg.Overrides.TraceIDKey, traceID}...)
	line, err := logfmt.MarshalKeyvals(keyvals...)
	if err != nil {
		level.Warn(p.logger).Log("msg", "unable to marshal keyvals", "err", err)
		return
	}

	// if we're logging to stdout, log and bail
	if p.logToStdout {
		level.Info(p.logger).Log(keyvals...)
		return
	}


	// Add loki label
	labels[model.LabelName(p.cfg.Overrides.LokiTag)] = model.LabelValue(kind)

	sent := p.logsInstance.SendEntry(api.Entry{
		Labels: labels,
		Entry: logproto.Entry{
			Timestamp: time.Now(),
			Line:      string(line),
		},
	}, p.cfg.Timeout)

	if !sent {
		level.Warn(p.logger).Log("msg", "failed to autolog to logs pipeline", "kind", kind, "traceid", traceID)
	}
}

func spanDuration(span pdata.Span) string {
	dur := int64(span.EndTimestamp() - span.StartTimestamp())
	return strconv.FormatInt(dur, 10) + "ns"
}

func attributeValue(att pdata.AttributeValue) interface{} {
	switch att.Type() {
	case pdata.AttributeValueTypeString:
		return att.StringVal()
	case pdata.AttributeValueTypeInt:
		return att.IntVal()
	case pdata.AttributeValueTypeDouble:
		return att.DoubleVal()
	case pdata.AttributeValueTypeBool:
		return att.BoolVal()
	case pdata.AttributeValueTypeMap:
		return att.MapVal()
	case pdata.AttributeValueTypeArray:
		return att.ArrayVal()
	}
	return nil
}

func override(cfgValue string, defaultValue string) string {
	if cfgValue == "" {
		return defaultValue
	}
	return cfgValue
}
