package views

import (
	semconv "go.opentelemetry.io/collector/semconv/v1.13.0"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"go.opentelemetry.io/otel/sdk/metric"
)

var (
	grpcScope = "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	// grpcUnacceptableKeyValues is a list of high cardinality grpc attributes that should be filtered out.
	grpcUnacceptableKeyValues = []attribute.KeyValue{
		attribute.String(semconv.AttributeNetSockPeerAddr, ""),
		attribute.String(semconv.AttributeNetSockPeerPort, ""),
		attribute.String(semconv.AttributeNetSockPeerName, ""),
	}

	httpScope = "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	// httpUnacceptableKeyValues is a list of high cardinality http attributes that should be filtered out.
	httpUnacceptableKeyValues = []attribute.KeyValue{
		attribute.String(semconv.AttributeNetHostName, ""),
		attribute.String(semconv.AttributeNetHostPort, ""),
		attribute.String(semconv.AttributeNetSockPeerPort, ""),
		attribute.String(semconv.AttributeNetSockPeerAddr, ""),
		attribute.String(semconv.AttributeHTTPClientIP, ""),
	}
)

func cardinalityFilter(kvs ...attribute.KeyValue) attribute.Filter {
	filter := attribute.NewSet(kvs...)
	return func(kv attribute.KeyValue) bool {
		return !filter.HasValue(kv.Key)
	}
}

// DropHighCardinalityServerAttributes drops certain high cardinality attributes from grpc/http server metrics
//
// This is a hopefully temporary fix to this upstream issue https://github.com/open-telemetry/opentelemetry-go-contrib/issues/3765
func DropHighCardinalityServerAttributes() []metric.View {
	var views []metric.View

	views = append(views, metric.NewView(
		metric.Instrument{Scope: instrumentation.Scope{Name: grpcScope}},
		metric.Stream{AttributeFilter: cardinalityFilter(grpcUnacceptableKeyValues...)}))

	views = append(views, metric.NewView(
		metric.Instrument{Scope: instrumentation.Scope{Name: httpScope}},
		metric.Stream{AttributeFilter: cardinalityFilter(httpUnacceptableKeyValues...)},
	))

	return views
}
