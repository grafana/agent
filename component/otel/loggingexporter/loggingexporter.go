package loggingexporter

import (
	"context"
	"fmt"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/otel"
	"github.com/grafana/agent/pkg/flow/logging"
	otelcomponent "go.opentelemetry.io/collector/component"
	otelconsumer "go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/model/pdata"
)

func init() {
	component.Register(component.Registration{
		Name:    "otel.exporter_logging",
		Args:    Arguments{},
		Exports: otel.ConsumerExports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

// Arguments configures the logging exporter.
type Arguments struct {
	// TODO(rfratto): don't ignore this field
	Level logging.Level `hcl:"level,optional"`
}

// Component implements the otlp.exporter_logging component.
type Component struct{}

func New(o component.Options, _ Arguments) (*Component, error) {
	impl := &exporterImpl{
		log: o.Logger,
	}

	// The only thing our component needs to do is initially set exports for the
	// implementation. This lasts through the lifetime of our component.
	o.OnStateChange(otel.ConsumerExports{
		Input: &otel.Consumer{CombinedConsumer: impl},
	})

	return &Component{}, nil
}

// Run implements component.Component.
func (c *Component) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

// Update implements component.Component.
func (c *Component) Update(args component.Arguments) error {
	// no-op: nothing to update (Arguments is empty struct).
	return nil
}

type exporterImpl struct {
	log log.Logger
}

func (e *exporterImpl) Capabilities() otelconsumer.Capabilities {
	return otelconsumer.Capabilities{MutatesData: false}
}

func (e *exporterImpl) Start(ctx context.Context, host otelcomponent.Host) error {
	// no-op
	return nil
}

func (e *exporterImpl) Shutdown(ctx context.Context) error {
	// no-op
	return nil
}

func (e *exporterImpl) ConsumeMetrics(ctx context.Context, td pdata.Metrics) error {
	return fmt.Errorf("cannot yet log metrics")
}

func (e *exporterImpl) ConsumeLogs(ctx context.Context, td pdata.Logs) error {
	return fmt.Errorf("cannot yet log logs")
}

func (e *exporterImpl) ConsumeTraces(ctx context.Context, td pdata.Traces) error {
	level.Debug(e.log).Log("msg", "got batch of traces")

	resSpans := td.ResourceSpans()
	for i := 0; i < resSpans.Len(); i++ {
		resSpan := resSpans.At(i)

		libSpans := resSpan.InstrumentationLibrarySpans()
		for j := 0; j < libSpans.Len(); j++ {
			libSpan := libSpans.At(j)

			spans := libSpan.Spans()
			for k := 0; k < spans.Len(); k++ {
				span := spans.At(k)

				level.Info(e.log).Log(
					"msg", "received span",
					"trace_id", span.TraceID().HexString(),
					"parent_id", span.ParentSpanID().HexString(),
					"id", span.SpanID().HexString(),
					"name", span.Name(),
					"kind", span.Kind(),
					"start_time", span.StartTimestamp().String(),
					"end_time", span.EndTimestamp().String(),
					"status_code", span.Status().Code().String(),
					"status_message", span.Status().Message(),
				)
			}
		}
	}

	return nil
}
