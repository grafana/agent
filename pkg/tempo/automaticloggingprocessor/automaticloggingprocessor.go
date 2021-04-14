package automaticloggingprocessor

import (
	"context"
	"fmt"
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
)

type automaticLoggingProcessor struct {
	nextConsumer consumer.TracesConsumer
	cfg          *AutomaticLoggingConfig
	lokiChan     chan<- api.Entry

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
	}, nil
}

func (p *automaticLoggingProcessor) ConsumeTraces(ctx context.Context, td pdata.Traces) error {
	rsLen := td.ResourceSpans().Len()
	for i := 0; i < rsLen; i++ {
		var traceID string
		rs := td.ResourceSpans().At(i)
		ilsLen := rs.InstrumentationLibrarySpans().Len()

		for j := 0; j < ilsLen; j++ {
			ils := rs.InstrumentationLibrarySpans().At(j)
			spanLen := ils.Spans().Len()

			for k := 0; k < spanLen; k++ {
				span := ils.Spans().At(k)
				traceID = span.TraceID().HexString()

				if p.cfg.EnableSpans {
					p.exportToLoki("span", traceID, "name", span.Name(), "dur", span.EndTime()-span.StartTime()) // name and duration not working
				}

				if p.cfg.EnableRoots && span.ParentSpanID().IsEmpty() {
					p.exportToLoki("root", traceID, "name", span.Name(), "dur", span.EndTime()-span.StartTime())
				}
			}
		}

		if p.cfg.EnableProcesses { // jpe multiple trace ids in the same batch :(
			atts := rs.Resource().Attributes()
			serviceName, ok := atts.Get(conventions.AttributeServiceName) // jpe include configurable tags
			if ok {
				p.exportToLoki("process", traceID, "name", serviceName.StringVal())
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
	return nil
}

func (p *automaticLoggingProcessor) exportToLoki(kind string, traceID string, keyvals ...interface{}) {
	keyvals = append(keyvals, []interface{}{"tid", traceID}...)
	line, err := logfmt.MarshalKeyvals(keyvals...)
	if err != nil {
		level.Warn(p.logger).Log("msg", "unable to marshal keyvals", "err", err)
		return
	}

	p.lokiChan <- api.Entry{ // jpe do something real
		Labels: model.LabelSet{
			"tempolog": model.LabelValue(kind), // jpe - const string kind - tempo-logger? better name?
		},
		Entry: logproto.Entry{
			Timestamp: time.Now(),
			Line:      string(line),
		},
	}
}
