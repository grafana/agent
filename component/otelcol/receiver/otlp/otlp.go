// Package otlp provides an otelcol.receiver.otlp component.
package otlp

import (
	"github.com/alecthomas/units"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/otelcol"
	"github.com/grafana/agent/component/otelcol/receiver"
	"github.com/grafana/agent/pkg/river"
	otelcomponent "go.opentelemetry.io/collector/component"
	otelconfig "go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/receiver/otlpreceiver"
)

func init() {
	component.Register(component.Registration{
		Name: "otelcol.receiver.otlp",
		Args: Arguments{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			fact := otlpreceiver.NewFactory()
			return receiver.New(opts, fact, args.(Arguments))
		},
	})
}

// Arguments configures the otelcol.receiver.otlp component.
type Arguments struct {
	GRPC *GRPCServerArguments `river:"grpc,block,optional"`
	HTTP *HTTPServerArguments `river:"http,block,optional"`

	// Output configures where to send received data. Required.
	Output *otelcol.ConsumerArguments `river:"output,block"`
}

var _ receiver.Arguments = Arguments{}

// Convert implements receiver.Arguments.
func (args Arguments) Convert() (otelconfig.Receiver, error) {
	return &otlpreceiver.Config{
		ReceiverSettings: otelconfig.NewReceiverSettings(otelconfig.NewComponentID("otlp")),
		Protocols: otlpreceiver.Protocols{
			GRPC: (*otelcol.GRPCServerArguments)(args.GRPC).Convert(),
			HTTP: (*otelcol.HTTPServerArguments)(args.HTTP).Convert(),
		},
	}, nil
}

// Extensions implements receiver.Arguments.
func (args Arguments) Extensions() map[otelconfig.ComponentID]otelcomponent.Extension {
	return nil
}

// Exporters implements receiver.Arguments.
func (args Arguments) Exporters() map[otelconfig.DataType]map[otelconfig.ComponentID]otelcomponent.Exporter {
	return nil
}

// NextConsumers implements receiver.Arguments.
func (args Arguments) NextConsumers() *otelcol.ConsumerArguments {
	return args.Output
}

type (
	// GRPCServerArguments is used to configure otelcol.receiver.otlp with
	// component-specific defaults.
	GRPCServerArguments otelcol.GRPCServerArguments

	// HTTPServerArguments is used to configure otelcol.receiver.otlp with
	// component-specific defaults.
	HTTPServerArguments otelcol.HTTPServerArguments
)

var (
	_ river.Unmarshaler = (*GRPCServerArguments)(nil)
	_ river.Unmarshaler = (*HTTPServerArguments)(nil)
)

// Default server settings.
var (
	DefaultGRPCServerArguments = GRPCServerArguments{
		Endpoint:  "0.0.0.0:4317",
		Transport: "tcp",

		ReadBufferSize: 512 * units.Kibibyte,
		// We almost write 0 bytes, so no need to tune WriteBufferSize.
	}

	DefaultHTTPServerArguments = HTTPServerArguments{
		Endpoint: "0.0.0.0:4318",
	}
)

// UnmarshalRiver implements river.Unmarshaler and supplies defaults.
func (args *GRPCServerArguments) UnmarshalRiver(f func(interface{}) error) error {
	*args = DefaultGRPCServerArguments
	type arguments GRPCServerArguments
	return f((*arguments)(args))
}

// UnmarshalRiver implements river.Unmarshaler and supplies defaults.
func (args *HTTPServerArguments) UnmarshalRiver(f func(interface{}) error) error {
	*args = DefaultHTTPServerArguments
	type arguments HTTPServerArguments
	return f((*arguments)(args))
}
