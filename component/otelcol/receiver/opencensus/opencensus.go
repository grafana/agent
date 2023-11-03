// Package opencensus provides an otelcol.receiver.opencensus component.
package opencensus

import (
	"github.com/alecthomas/units"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/otelcol"
	"github.com/grafana/agent/component/otelcol/receiver"
	otel_service "github.com/grafana/agent/service/otel"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/opencensusreceiver"
	otelcomponent "go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/extension"
)

func init() {
	component.Register(component.Registration{
		Name:          "otelcol.receiver.opencensus",
		Args:          Arguments{},
		NeedsServices: []string{otel_service.ServiceName},

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

	// DebugMetrics configures component internal metrics. Optional.
	DebugMetrics otelcol.DebugMetricsArguments `river:"debug_metrics,block,optional"`

	// Output configures where to send received data. Required.
	Output *otelcol.ConsumerArguments `river:"output,block"`
}

var _ receiver.Arguments = Arguments{}

// Default server settings.
var DefaultArguments = Arguments{
	GRPC: otelcol.GRPCServerArguments{
		Endpoint:  "0.0.0.0:4317",
		Transport: "tcp",

		ReadBufferSize: 512 * units.Kibibyte,
		// We almost write 0 bytes, so no need to tune WriteBufferSize.
	},
}

// SetToDefault implements river.Defaulter.
func (args *Arguments) SetToDefault() {
	*args = DefaultArguments
}

// Convert implements receiver.Arguments.
func (args Arguments) Convert() (otelcomponent.Config, error) {
	return &opencensusreceiver.Config{
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

// DebugMetricsConfig implements receiver.Arguments.
func (args Arguments) DebugMetricsConfig() otelcol.DebugMetricsArguments {
	return args.DebugMetrics
}
