// Package otel holds a collection of OpenTelemetry Collector components.
package otel

import (
	"github.com/grafana/agent/component/otel/internal/errorconsumer"
	"github.com/grafana/agent/component/otel/internal/fanoutconsumer"
	otelconfiggrpc "go.opentelemetry.io/collector/config/configgrpc"
	otelconfighttp "go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/config/confignet"
	otelconsumer "go.opentelemetry.io/collector/consumer"
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
