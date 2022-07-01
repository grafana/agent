// Package otel holds a collection of OpenTelemetry Collector components.
package otel

import (
	"context"
	"fmt"
	"sync"

	"github.com/grafana/agent/component/otel/internal/errorconsumer"
	"github.com/grafana/agent/component/otel/internal/fanoutconsumer"
	otelconfiggrpc "go.opentelemetry.io/collector/config/configgrpc"
	otelconfighttp "go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/config/confignet"
	otelconsumer "go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/model/pdata"
)

// GRPCServerArguments holds shared gRPC settings for components which launch gRPC
// servers.
type GRPCServerArguments struct {
	Endpoint  string `hcl:"endpoint,optional"`
	Transport string `hcl:"transport,optional"`

	// TODO(rfratto): TLS

	MaxRecvMsgSizeMiB    uint64 `hcl:"max_recv_msg_size_mib,optional"`
	MaxConcurrentStreams uint32 `hcl:"max_concurrent_streams,optional"`
	ReadBufferSize       int    `hcl:"read_buffer_size,optional"`
	WriteBufferSize      int    `hcl:"write_buffer_size,optional"`

	// TODO(rfratto): keepalive
	// TODO(rfratto): auth

	IncludeMetadata bool `hcl:"include_metadata,optional"`
}

func (args *GRPCServerArguments) Convert() *otelconfiggrpc.GRPCServerSettings {
	if args == nil {
		return nil
	}

	return &otelconfiggrpc.GRPCServerSettings{
		NetAddr: confignet.NetAddr{
			Endpoint:  args.Endpoint,
			Transport: args.Transport,
		},
		// TODO(rfratto): TLS
		MaxRecvMsgSizeMiB:    args.MaxRecvMsgSizeMiB,
		MaxConcurrentStreams: args.MaxConcurrentStreams,
		ReadBufferSize:       args.ReadBufferSize,
		WriteBufferSize:      args.WriteBufferSize,
		// TODO(rfratto): keepalive
		// TODO(rfratto): auth
		IncludeMetadata: args.IncludeMetadata,
	}
}

func NewGRPCServerSettings(args *GRPCServerArguments) *otelconfiggrpc.GRPCServerSettings {
	if args == nil {
		return nil
	}

	return &otelconfiggrpc.GRPCServerSettings{}
}

// HTTPServerArguments holds shared gRPC settings for components which launch HTTP
// servers.
type HTTPServerArguments struct {
	Endpoint string `hcl:"endpoint,optional"`

	// TODO(rfratto): TLS
	// TODO(rfratto): CORS
	// TODO(rfratto): Auth

	MaxRequestBodySize int64 `hcl:"max_request_body_size,optional"`

	IncludeMetadata bool `hcl:"include_metadata,optional"`
}

func (args *HTTPServerArguments) Convert() *otelconfighttp.HTTPServerSettings {
	if args == nil {
		return nil
	}

	return &otelconfighttp.HTTPServerSettings{
		Endpoint: args.Endpoint,
		// TODO(rfratto): TLS
		// TODO(rfratto): CORS
		// TODO(rfratto): Auth
		MaxRequestBodySize: args.MaxRequestBodySize,
		IncludeMetadata:    args.IncludeMetadata,
	}
}

type NextReceiverArguments struct {
	Metrics []*Consumer `hcl:"metrics,optional"`
	Logs    []*Consumer `hcl:"logs,optional"`
	Traces  []*Consumer `hcl:"traces,optional"`
}

type (
	MetricsConsumer struct{ otelconsumer.Metrics }
	LogsConsumer    struct{ otelconsumer.Logs }
	TracesConsumer  struct{ otelconsumer.Traces }
)

func (args *NextReceiverArguments) MetricsConsumer() otelconsumer.Metrics {
	if args == nil || len(args.Metrics) == 0 {
		return errorconsumer.Metrics
	}

	conv := make([]otelconsumer.Metrics, len(args.Metrics))
	for i := range args.Metrics {
		conv[i] = args.Metrics[i]
	}
	return fanoutconsumer.Metrics(conv)
}

func (args *NextReceiverArguments) LogsConsumer() otelconsumer.Logs {
	if args == nil || len(args.Logs) == 0 {
		return errorconsumer.Logs
	}

	conv := make([]otelconsumer.Logs, len(args.Logs))
	for i := range args.Logs {
		conv[i] = args.Logs[i]
	}
	return fanoutconsumer.Logs(conv)
}

func (args *NextReceiverArguments) TracesConsumer() otelconsumer.Traces {
	if args == nil || len(args.Traces) == 0 {
		return errorconsumer.Traces
	}

	conv := make([]otelconsumer.Traces, len(args.Traces))
	for i := range args.Traces {
		conv[i] = args.Traces[i]
	}
	return fanoutconsumer.Traces(conv)
}

type lazyConsumer struct {
	mut     sync.RWMutex
	metrics otelconsumer.Metrics
	logs    otelconsumer.Logs
	traces  otelconsumer.Traces
}

var (
	_ otelconsumer.Metrics = (*lazyConsumer)(nil)
	_ otelconsumer.Logs    = (*lazyConsumer)(nil)
	_ otelconsumer.Traces  = (*lazyConsumer)(nil)
)

func (lazy *lazyConsumer) Capabilities() otelconsumer.Capabilities {
	// TODO(rfratto): this implementation is inefficient since it always requires
	// data to be copied in the pipeline.
	//
	// To make things more efficient, we would have to:
	//
	// - Block on calls to Capabilities until our inner consumers are set
	// - Split up lazyConsumer into a lazyConsumer for Metrics/Logs/Traces
	return otelconsumer.Capabilities{MutatesData: true}
}

func (lazy *lazyConsumer) ConsumeMetrics(ctx context.Context, md pdata.Metrics) error {
	lazy.mut.RLock()
	defer lazy.mut.RUnlock()

	if lazy.metrics == nil {
		return fmt.Errorf("metrics consumer doesn't exist")
	}
	return lazy.metrics.ConsumeMetrics(ctx, md)
}

func (lazy *lazyConsumer) ConsumeLogs(ctx context.Context, md pdata.Logs) error {
	lazy.mut.RLock()
	defer lazy.mut.RUnlock()

	if lazy.logs == nil {
		return fmt.Errorf("logs consumer doesn't exist")
	}
	return lazy.logs.ConsumeLogs(ctx, md)
}

func (lazy *lazyConsumer) ConsumeTraces(ctx context.Context, md pdata.Traces) error {
	lazy.mut.RLock()
	defer lazy.mut.RUnlock()

	if lazy.traces == nil {
		return fmt.Errorf("traces consumer doesn't exist")
	}
	return lazy.traces.ConsumeTraces(ctx, md)
}

type updateConsumerFunc func() (otelconsumer.Metrics, otelconsumer.Logs, otelconsumer.Traces)

// Update updates the lazy consumer with the result of calling f. Data sent to
// the consumer is blocked while f is being invoked.
func (lazy *lazyConsumer) Update(f updateConsumerFunc) {
	lazy.mut.Lock()
	defer lazy.mut.Unlock()

	// TODO(rfratto): we're assuming that metrics/logs/traces will be nil when
	// something failed, but it'd be nice to have some kind of unhealthy consumer
	// which reports specific errors back to the caller when a component goes
	// unhealthy.
	lazy.metrics, lazy.logs, lazy.traces = f()
}
