package automaticloggingprocessor

import (
	"context"
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

type automaticLoggingProcessor struct {
	nextConsumer consumer.TracesConsumer
	cfg          *AutomaticLoggingConfig
	lokiChan     chan<- api.Entry
	done         atomic.Bool

	logger log.Logger
}

func newTraceProcessor(nextConsumer consumer.TracesConsumer, cfg *AutomaticLoggingConfig) (component.TracesProcessor, error) {
	logger := log.With(util.Logger, "component", "tempo automatic logging")

	if nextConsumer == nil {
		return nil, componenterror.ErrNilNextConsumer
	}
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

				if p.cfg.EnableSpans {
					p.exportToLoki("span", traceID, p.spanKeyVals(span, svc)...)
				}

				if p.cfg.EnableRoots && span.ParentSpanID().IsEmpty() {
					p.exportToLoki("root", traceID, p.spanKeyVals(span, svc)...)
				}

				if p.cfg.EnableProcesses && lastTraceID != traceID {
					lastTraceID = traceID
					p.exportToLoki("process", traceID, p.processKeyVals(rs.Resource(), svc)...)
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
		return fmt.Errorf("key %s does not contain a Loki instance", contextkeys.Loki)
	}
	lokiInstance := loki.Instance(p.cfg.LokiName)
	if lokiInstance == nil {
		return fmt.Errorf("loki instance %s not found", p.cfg.LokiName)
	}
	p.lokiChan = lokiInstance.Promtail().Client().Chan()
	if p.lokiChan == nil {
		return fmt.Errorf("loki chan is unexpectedly nil")
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
	atts = append(atts, "svc")
	atts = append(atts, svc)

	for _, name := range p.cfg.ProcessAttributes {
		att, ok := rsAtts.Get(name)
		if ok {
			// name/key val pairs
			atts = append(atts, att)
			atts = append(atts, attributeValue(att))
		}
	}

	return atts
}

func (p *automaticLoggingProcessor) spanKeyVals(span pdata.Span, svc string) []interface{} {
	atts := make([]interface{}, 0, 8) // 8 for name, duration and service name

	atts = append(atts, "name")
	atts = append(atts, span.Name())

	atts = append(atts, "dur")
	atts = append(atts, spanDuration(span))

	atts = append(atts, "svc")
	atts = append(atts, svc)

	atts = append(atts, "status")
	atts = append(atts, span.Status().Code())

	span.Status().Code()

	for _, name := range p.cfg.SpanAttributes {
		att, ok := span.Attributes().Get(name)
		if ok {
			atts = append(atts, attributeValue(att))
		}
	}

	return atts
}

func (p *automaticLoggingProcessor) exportToLoki(kind string, traceID string, keyvals ...interface{}) {
	if p.done.Load() {
		return
	}

	keyvals = append(keyvals, []interface{}{"tid", traceID}...)
	line, err := logfmt.MarshalKeyvals(keyvals...)
	if err != nil {
		level.Warn(p.logger).Log("msg", "unable to marshal keyvals", "err", err)
		return
	}

	p.lokiChan <- api.Entry{
		Labels: model.LabelSet{
			"tempolog": model.LabelValue(kind), // jpe - const string kind - tempo-logger? better name?
		},
		Entry: logproto.Entry{
			Timestamp: time.Now(),
			Line:      string(line),
		},
	}
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
	case pdata.AttributeValueMAP: // jpe test this and below?
		return att.MapVal()
	case pdata.AttributeValueARRAY:
		return att.ArrayVal()
	}
	return nil
}
