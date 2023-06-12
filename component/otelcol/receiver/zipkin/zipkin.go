// Package zipkin provides an otelcol.receiver.zipkin component.
package zipkin

import (
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/otelcol"
	"github.com/grafana/agent/component/otelcol/receiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/zipkinreceiver"
	otelcomponent "go.opentelemetry.io/collector/component"
	otelconfig "go.opentelemetry.io/collector/config"
)

func init() {
	component.Register(component.Registration{
		Name: "otelcol.receiver.zipkin",
		Args: Arguments{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			fact := zipkinreceiver.NewFactory()
			return receiver.New(opts, fact, args.(Arguments))
		},
	})
}

// Arguments configures the otelcol.receiver.zipkin component.
type Arguments struct {
	ParseStringTags bool `river:"parse_string_tags,attr,optional"`

	HTTPServer otelcol.HTTPServerArguments `river:",squash"`

	// Output configures where to send received data. Required.
	Output *otelcol.ConsumerArguments `river:"output,block"`
}

var (
	_ receiver.Arguments = Arguments{}
)

// DefaultArguments holds default settings for otelcol.receiver.zipkin.
var DefaultArguments = Arguments{
	HTTPServer: otelcol.HTTPServerArguments{
		Endpoint: "0.0.0.0:9411",
	},
}

// SetToDefault implements river.Defaulter.
func (args *Arguments) SetToDefault() {
	*args = DefaultArguments
}

// Convert implements receiver.Arguments.
func (args Arguments) Convert() (otelconfig.Receiver, error) {
	return &zipkinreceiver.Config{
		ReceiverSettings: otelconfig.NewReceiverSettings(otelconfig.NewComponentID("zipkin")),

		ParseStringTags:    args.ParseStringTags,
		HTTPServerSettings: *args.HTTPServer.Convert(),
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
