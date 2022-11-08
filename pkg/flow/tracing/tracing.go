// Package tracing implements the tracing subsystem of Grafana Agent Flow. The
// tracing subsystem exposes a [trace.TraceProvider] which accepts traces and
// forwards them to a running component for further processing.
package tracing

import (
	"context"
	"time"

	"github.com/grafana/agent/component/otelcol"
	"github.com/grafana/agent/pkg/build"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.opentelemetry.io/otel/trace"
)

// Options control the tracing subsystem.
type Options struct {
	// SamplingFraction determines which rate of traces to sample. A value of 1
	// means to keep 100% of traces. A value of 0 means to keep 0% of traces.
	SamplingFraction float64 `river:"sampling_fraction,attr,optional"`

	// WriteTo holds a set of OpenTelemetry Collector consumers where internal
	// traces should be sent.
	WriteTo []otelcol.Consumer `river:"write_to,attr,optional"`
}

// DefaultOptions holds default configuration options.
var DefaultOptions = Options{
	SamplingFraction: 0.1,                  // Keep 10% of spans
	WriteTo:          []otelcol.Consumer{}, // Don't send spans anywhere.
}

// Tracer is the tracing subsystem of Grafana Agent Flow. It implements
// [trace.TracerProvider] and can be used to forward internally generated
// traces to a OpenTelemetry Collector-compatible Flow component.
type Tracer struct {
	sampler *dynamicSampler
	client  *client
	exp     *otlptrace.Exporter
	tp      *tracesdk.TracerProvider
}

var _ trace.TracerProvider = (*Tracer)(nil)

// New creates a new tracing subsystem. Call Run to start the tracing
// subsystem.
func New(cfg Options) (*Tracer, error) {
	res, err := resource.New(
		context.Background(),
		resource.WithSchemaURL(semconv.SchemaURL),
		resource.WithAttributes(
			semconv.ServiceNameKey.String("grafana-agent"),
			semconv.ServiceVersionKey.String(build.Version),
		),
		resource.WithProcessRuntimeDescription(),
		resource.WithTelemetrySDK(),
	)
	if err != nil {
		return nil, err
	}

	sampler := newDynamicSampler(cfg.SamplingFraction)

	shimClient := &client{}
	exp := otlptrace.NewUnstarted(shimClient)

	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp),
		tracesdk.WithSampler(tracesdk.ParentBased(sampler)),
		tracesdk.WithResource(res),
	)

	return &Tracer{
		sampler: sampler,
		client:  shimClient,
		exp:     exp,
		tp:      tp,
	}, nil
}

// Update provides a new config to the tracing subsystem.
func (t *Tracer) Update(opts Options) error {
	t.sampler.UpdateSampleRate(opts.SamplingFraction)
	t.client.UpdateWriteTo(opts.WriteTo)
	return nil
}

// Run starts the tracing subsystem and runs it until the provided context is
// canceled. If the tracing subsystem could not be started, an error is
// returned.
//
// Run returns no error upon normal shutdown.
func (t *Tracer) Run(ctx context.Context) error {
	if err := t.exp.Start(ctx); err != nil {
		return err
	}

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := t.tp.Shutdown(shutdownCtx); err != nil {
		return err
	}
	return nil
}

func (t *Tracer) Tracer(name string, options ...trace.TracerOption) trace.Tracer {
	return t.tp.Tracer(name, options...)
}
