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
	"go.opentelemetry.io/collector/extension"
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

	GRPC otelcol.GRPCServerArguments `river:",squash"`

	// Output configures where to send received data. Required.
	Output *otelcol.ConsumerArguments `river:"output,block"`
}

var (
	_ receiver.Arguments = Arguments{}
	_ river.Unmarshaler  = (*Arguments)(nil)
)

// Default server settings.
var DefaultArguments = Arguments{
	GRPC: otelcol.GRPCServerArguments{
		Endpoint:  "0.0.0.0:4317",
		Transport: "tcp",

		ReadBufferSize: 512 * units.Kibibyte,
		// We almost write 0 bytes, so no need to tune WriteBufferSize.
	},
}

// UnmarshalRiver implements river.Unmarshaler and supplies defaults.
func (args *Arguments) UnmarshalRiver(f func(interface{}) error) error {
	*args = DefaultArguments

	type arguments Arguments
	return f((*arguments)(args))
}

// Convert implements receiver.Arguments.
func (args Arguments) Convert() (otelcomponent.Config, error) {
	return &opencensusreceiver.Config{
		//TODO: Should we replace this with something?
		// ReceiverSettings: otelcomponent.NewReceiverConfigSettings(otelcomponent.NewID("opencensus")),

		CorsOrigins:        args.CorsAllowedOrigins,
		GRPCServerSettings: *args.GRPC.Convert(),
	}, nil
}

// Extensions implements receiver.Arguments.
func (args Arguments) Extensions() map[otelcomponent.ID]extension.Extension {
	return nil
}

// Exporters implements receiver.Arguments.
func (args Arguments) Exporters() map[otelcomponent.DataType]map[otelcomponent.ID]otelcomponent.Component {
	return nil
}

// NextConsumers implements receiver.Arguments.
func (args Arguments) NextConsumers() *otelcol.ConsumerArguments {
	return args.Output
}
