package traces

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var (
	componentIDAttributeKey = "grafana_agent.component_id"
)

// WrapTracer returns a new trace.TracerProvider which will inject the provided
// componentID as an attribute to each span.
func WrapTracer(inner trace.TracerProvider, componentID string) trace.TracerProvider {
	return &wrappedProvider{
		inner: inner,
		id:    componentID,
	}
}

type wrappedProvider struct {
	inner trace.TracerProvider
	id    string
}

var _ trace.TracerProvider = (*wrappedProvider)(nil)

func (wp *wrappedProvider) Tracer(name string, options ...trace.TracerOption) trace.Tracer {
	innerTracer := wp.inner.Tracer(name, options...)
	return &wrappedTracer{
		inner: innerTracer,
		id:    wp.id,
	}
}

type wrappedTracer struct {
	inner trace.Tracer
	id    string
}

var _ trace.Tracer = (*wrappedTracer)(nil)

func (tp *wrappedTracer) Start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	ctx, span := tp.inner.Start(ctx, spanName, opts...)
	span.SetAttributes(
		attribute.String(componentIDAttributeKey, tp.id),
	)

	return ctx, span
}
