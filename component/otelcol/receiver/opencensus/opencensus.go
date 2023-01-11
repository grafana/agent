// Package opencensus provides an otelcol.receiver.opencensus component.
package opencensus

import (
	"github.com/alecthomas/units"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/otelcol"
	"github.com/grafana/agent/component/otelcol/receiver"
	"github.com/grafana/agent/pkg/river"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/opencensusreceiver"
	otelcomponent "go.opentelemetry.io/collector/component"
	otelconfig "go.opentelemetry.io/collector/config"
)

func init() {
	component.Register(component.Registration{
		Name: "otelcol.receiver.opencensus",
		Args: Arguments{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			fact := opencensusreceiver.NewFactory()
			return receiver.New(opts, fact, args.(Arguments))
		},
	})
}

// Arguments configures the otelcol.receiver.opencensus component.
type Arguments struct {
	CorsAllowedOrigins []string `river:"cors_allowed_origins,attr,optional"`

	GRPC *GRPCServerArguments `river:"grpc,block,optional"`

	// Output configures where to send received data. Required.
	Output *otelcol.ConsumerArguments `river:"output,block"`
}

var _ receiver.Arguments = Arguments{}

// Convert implements receiver.Arguments.
func (args Arguments) Convert() otelconfig.Receiver {
	return &opencensusreceiver.Config{
		ReceiverSettings: otelconfig.NewReceiverSettings(otelconfig.NewComponentID("opencensus")),

		CorsOrigins:        args.CorsAllowedOrigins,
		GRPCServerSettings: *(*otelcol.GRPCServerArguments)(args.GRPC).Convert(),
	}
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
)

var (
	_ river.Unmarshaler = (*GRPCServerArguments)(nil)
)

// Default server settings.
var (
	DefaultGRPCServerArguments = GRPCServerArguments{
		Endpoint:  "0.0.0.0:4317",
		Transport: "tcp",

		ReadBufferSize: 512 * units.Kibibyte,
		// We almost write 0 bytes, so no need to tune WriteBufferSize.
	}
)

// UnmarshalRiver implements river.Unmarshaler and supplies defaults.
func (args *GRPCServerArguments) UnmarshalRiver(f func(interface{}) error) error {
	*args = DefaultGRPCServerArguments
	type arguments GRPCServerArguments
	return f((*arguments)(args))
}
