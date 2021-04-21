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
	"github.com/grafana/agent/pkg/loki"
	"github.com/grafana/agent/pkg/tempo/contextkeys"
	"github.com/grafana/loki/pkg/logproto"
	"github.com/grafana/loki/pkg/promtail/api"
	"github.com/prometheus/common/model"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenterror"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/consumer/pdata"
	"go.opentelemetry.io/collector/translator/conventions"
	"go.uber.org/atomic"
)

const (
	defaultLokiTag     = "tempologging"
	defaultServiceKey  = "svc"
	defaultSpanNameKey = "span"
	defaultStatusKey   = "status"
	defaultDurationKey = "dur"
	defaultTraceIDKey  = "tid"

	defaultTimeout = 100 * time.Millisecond

	typeSpan    = "span"
	typeRoot    = "root"
	typeProcess = "process"
)

type automaticLoggingProcessor struct {
	nextConsumer consumer.TracesConsumer
	cfg          *AutomaticLoggingConfig
	lokiInstance *loki.Instance
	done         atomic.Bool

	logger log.Logger
}

func newTraceProcessor(nextConsumer consumer.TracesConsumer, cfg *AutomaticLoggingConfig) (component.TracesProcessor, error) {
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

	cfg.Overrides.LokiTag = override(cfg.Overrides.LokiTag, defaultLokiTag)
	cfg.Overrides.ServiceKey = override(cfg.Overrides.ServiceKey, defaultServiceKey)
	cfg.Overrides.SpanNameKey = override(cfg.Overrides.SpanNameKey, defaultSpanNameKey)
	cfg.Overrides.StatusKey = override(cfg.Overrides.StatusKey, defaultStatusKey)
	cfg.Overrides.DurationKey = override(cfg.Overrides.DurationKey, defaultDurationKey)
	cfg.Overrides.TraceIDKey = override(cfg.Overrides.TraceIDKey, defaultTraceIDKey)

	return &automaticLoggingProcessor{
		nextConsumer: nextConsumer,
		cfg:          cfg,
		logger:       logger,
		done:         atomic.Bool{},
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
					p.exportToLoki(typeSpan, traceID, p.spanKeyVals(span, svc)...)
				}

				if p.cfg.Roots && span.ParentSpanID().IsEmpty() {
					p.exportToLoki(typeRoot, traceID, p.spanKeyVals(span, svc)...)
				}

				if p.cfg.Processes && lastTraceID != traceID {
					lastTraceID = traceID
					p.exportToLoki(typeProcess, traceID, p.processKeyVals(rs.Resource(), svc)...)
				}
			}
		}
	}

	return p.nextConsumer.ConsumeTraces(ctx, td)
}

func (p *automaticLoggingProcessor) GetCapabilities() component.ProcessorCapabilities {
	return component.ProcessorCapabilities{}
}

// Start is invoked during service startup.
func (p *automaticLoggingProcessor) Start(ctx context.Context, _ component.Host) error {
	loki := ctx.Value(contextkeys.Loki).(*loki.Loki)
	if loki == nil {
		return fmt.Errorf("key does not contain a Loki instance")
	}
	p.lokiInstance = loki.Instance(p.cfg.LokiName)
	if p.lokiInstance == nil {
		return fmt.Errorf("loki instance %s not found", p.cfg.LokiName)
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

func (p *automaticLoggingProcessor) spanKeyVals(span pdata.Span, svc string) []interface{} {
	atts := make([]interface{}, 0, 8) // 8 for name, duration and service name

	atts = append(atts, p.cfg.Overrides.SpanNameKey)
	atts = append(atts, span.Name())

	atts = append(atts, p.cfg.Overrides.DurationKey)
	atts = append(atts, spanDuration(span))

	atts = append(atts, p.cfg.Overrides.ServiceKey)
	atts = append(atts, svc)

	atts = append(atts, p.cfg.Overrides.StatusKey)
	atts = append(atts, span.Status().Code())

	span.Status().Code()

	for _, name := range p.cfg.SpanAttributes {
		att, ok := span.Attributes().Get(name)
		if ok {
			atts = append(atts, name)
			atts = append(atts, attributeValue(att))
		}
	}

	return atts
}

func (p *automaticLoggingProcessor) exportToLoki(kind string, traceID string, keyvals ...interface{}) {
	if p.done.Load() {
		return
	}

	keyvals = append(keyvals, []interface{}{p.cfg.Overrides.TraceIDKey, traceID}...)
	line, err := logfmt.MarshalKeyvals(keyvals...)
	if err != nil {
		level.Warn(p.logger).Log("msg", "unable to marshal keyvals", "err", err)
		return
	}

	sent := p.lokiInstance.SendEntry(api.Entry{
		Labels: model.LabelSet{
			model.LabelName(p.cfg.Overrides.LokiTag): model.LabelValue(kind),
		},
		Entry: logproto.Entry{
			Timestamp: time.Now(),
			Line:      string(line),
		},
	}, p.cfg.Timeout)

	level.Warn(p.logger).Log("msg", "failed to autolog to loki", "kind", kind, "traceid", traceID)
}

func spanDuration(span pdata.Span) string {
	dur := int64(span.EndTime() - span.StartTime())
	return strconv.FormatInt(dur, 10) + "ns"
}

func attributeValue(att pdata.AttributeValue) interface{} {
	switch att.Type() {
	case pdata.AttributeValueSTRING:
		return att.StringVal()
	case pdata.AttributeValueINT:
		return att.IntVal()
	case pdata.AttributeValueDOUBLE:
		return att.DoubleVal()
	case pdata.AttributeValueBOOL:
		return att.BoolVal()
	case pdata.AttributeValueMAP:
		return att.MapVal()
	case pdata.AttributeValueARRAY:
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
