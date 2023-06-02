package tracing

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var (
	componentIDAttributeKey  = "grafana_agent.component_id"
	controllerIDAttributeKey = "grafana_agent.controller_id"
)

// WrapTracer returns a new trace.TracerProvider which will inject the provided
// componentID as an attribute to each span.
func WrapTracer(inner trace.TracerProvider, componentID string) trace.TracerProvider {
	return &wrappedProvider{
		inner:    inner,
		id:       componentID,
		spanName: componentIDAttributeKey,
	}
}

// WrapTracerForLoader returns a new trace.TracerProvider which will inject the provided
// controller id as an attribute to each span.
func WrapTracerForLoader(inner trace.TracerProvider, componentID string) trace.TracerProvider {
	return &wrappedProvider{
		inner:    inner,
		id:       componentID,
		spanName: controllerIDAttributeKey,
	}
}

type wrappedProvider struct {
	inner    trace.TracerProvider
	id       string
	spanName string
}

var _ trace.TracerProvider = (*wrappedProvider)(nil)

func (wp *wrappedProvider) Tracer(name string, options ...trace.TracerOption) trace.Tracer {
	innerTracer := wp.inner.Tracer(name, options...)
	return &wrappedTracer{
		inner:    innerTracer,
		id:       wp.id,
		spanName: wp.spanName,
	}
}

type wrappedTracer struct {
	inner    trace.Tracer
	id       string
	spanName string
}

var _ trace.Tracer = (*wrappedTracer)(nil)

func (tp *wrappedTracer) Start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	ctx, span := tp.inner.Start(ctx, spanName, opts...)
	if tp.id != "" {
		span.SetAttributes(
			attribute.String(tp.spanName, tp.id),
		)
	}

	return ctx, span
}
