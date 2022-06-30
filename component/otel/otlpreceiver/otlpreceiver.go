package otlpreceiver

import (
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/otel"
	"github.com/hashicorp/hcl/v2"
	"github.com/rfratto/gohcl"

	otelcomponent "go.opentelemetry.io/collector/component"
	otelconfig "go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/receiver/otlpreceiver"
)

func init() {
	component.Register(component.Registration{
		Name: "otel.receiver_otlp",
		Args: Arguments{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			fact := otlpreceiver.NewFactory()
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
		Endpoint:       "0.0.0.0:4317",
		Transport:      "tcp",
		ReadBufferSize: 512 * 1024,
		// We almost write 0 bytes, so no need to tune WriteBufferSize.
	}

	DefaultHTTPServerArguments = HTTPServerArguments{
		Endpoint: "0.0.0.0:4318",
	}
)

// Convert transforms OTLPReceiverArguments into the upstream otlpreceiver
// Config.
func (args Arguments) Convert() otelconfig.Receiver {
	return &otlpreceiver.Config{
		ReceiverSettings: otelconfig.NewReceiverSettings(otelconfig.NewComponentID("otlp")),
		Protocols: otlpreceiver.Protocols{
			GRPC: (*otel.GRPCServerArguments)(args.GRPC).Convert(),
			HTTP: (*otel.HTTPServerArguments)(args.HTTP).Convert(),
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
