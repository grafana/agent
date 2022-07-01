package jaegerreceiver

import (
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/otel"
	"github.com/hashicorp/hcl/v2"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/jaegerreceiver"
	"github.com/rfratto/gohcl"

	otelcomponent "go.opentelemetry.io/collector/component"
	otelconfig "go.opentelemetry.io/collector/config"
)

func init() {
	component.Register(component.Registration{
		Name: "otel.receiver_jaeger",
		Args: Arguments{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			fact := jaegerreceiver.NewFactory()
			return otel.NewFlowReceiver(opts, fact, args.(Arguments))
		},
	})
}

type (
	// Arguments configures the OTLP receiver.
	Arguments struct {
		GRPC *GRPCServerArguments `hcl:"grpc,block"`
		HTTP *HTTPServerArguments `hcl:"http,block"`

		// Output configures where to send data. Must be provided.
		Output otel.NextReceiverArguments `hcl:"output,block"`
	}

	// GRPCServerArguments includes GRPC arguments with receiver_otlp-specific
	// defaults.
	GRPCServerArguments otel.GRPCServerArguments

	// HTTPServerArguments includes HTTP arguments with receiver_otlp-specific
	// defaults.
	HTTPServerArguments otel.HTTPServerArguments
)

var (
	_ otel.ReceiverArguments = (*Arguments)(nil)
	_ gohcl.Decoder          = (*GRPCServerArguments)(nil)
	_ gohcl.Decoder          = (*HTTPServerArguments)(nil)
)

// Default settings.
var (
	DefaultGRPCServerArguments = GRPCServerArguments{
		Endpoint:  "0.0.0.0:14250",
		Transport: "tcp",
	}

	DefaultHTTPServerArguments = HTTPServerArguments{
		Endpoint: "0.0.0.0:14268",
	}
)

// Convert transforms OTLPReceiverArguments into the upstream otlpreceiver
// Config.
func (args Arguments) Convert() otelconfig.Receiver {
	return &jaegerreceiver.Config{
		ReceiverSettings: otelconfig.NewReceiverSettings(otelconfig.NewComponentID("jaeger")),
		Protocols: jaegerreceiver.Protocols{
			GRPC:       (*otel.GRPCServerArguments)(args.GRPC).Convert(),
			ThriftHTTP: (*otel.HTTPServerArguments)(args.HTTP).Convert(),
			// TODO(rfratto): ProtocolUDP for ThriftBinary/ThiftCompact
		},
	}
}

func (args Arguments) Extensions() map[otelconfig.ComponentID]otelcomponent.Extension {
	return nil
}

func (args Arguments) Exporters() map[otelconfig.DataType]map[otelconfig.ComponentID]otelcomponent.Exporter {
	return nil
}

func (args Arguments) NextArguments() *otel.NextReceiverArguments {
	return &args.Output
}

func (args *GRPCServerArguments) DecodeHCL(body hcl.Body, ctx *hcl.EvalContext) error {
	*args = DefaultGRPCServerArguments
	return gohcl.DecodeBody(body, ctx, (*otel.GRPCServerArguments)(args))
}

func (args *HTTPServerArguments) DecodeHCL(body hcl.Body, ctx *hcl.EvalContext) error {
	*args = DefaultHTTPServerArguments
	return gohcl.DecodeBody(body, ctx, (*otel.HTTPServerArguments)(args))
}
