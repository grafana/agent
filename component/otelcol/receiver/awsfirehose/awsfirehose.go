// Package zipkin provides an otelcol.receiver.zipkin component.
package awsfirehose

import (
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/otelcol"
	"github.com/grafana/agent/component/otelcol/receiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/awsfirehosereceiver"
	otelcomponent "go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configopaque"
	otelextension "go.opentelemetry.io/collector/extension"
)

func init() {
	component.Register(component.Registration{
		Name: "otelcol.receiver.awsfirehose",
		Args: Arguments{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			fact := awsfirehosereceiver.NewFactory()
			return receiver.New(opts, fact, args.(Arguments))
		},
	})
}

// Arguments configures the otelcol.receiver.awsfirehose component.
type Arguments struct {
	// The type of record being received from the delivery stream.
	// Each unmarshaler handles a specific type,
	// so the field allows the receiver to use the correct one.
	RecordType string `river:"record_type,attr,optional"`

	// The access key to be checked on each request received.
	AccessKey string `river:"access_key,attr,optional"`

	HTTPServer otelcol.HTTPServerArguments `river:",squash"`

	// DebugMetrics configures component internal metrics. Optional.
	DebugMetrics otelcol.DebugMetricsArguments `river:"debug_metrics,block,optional"`

	// Output configures where to send received data. Required.
	Output *otelcol.ConsumerArguments `river:"output,block"`
}

var _ receiver.Arguments = Arguments{}

// DefaultArguments holds default settings for otelcol.receiver.awsfirehose.
var DefaultArguments = Arguments{
	RecordType: "cwmetrics",
	HTTPServer: otelcol.HTTPServerArguments{
		Endpoint: "0.0.0.0:4433",
	},
}

// SetToDefault implements river.Defaulter.
func (args *Arguments) SetToDefault() {
	*args = DefaultArguments
}

// Convert implements receiver.Arguments.
func (args Arguments) Convert() (otelcomponent.Config, error) {
	return &awsfirehosereceiver.Config{
		RecordType:         args.RecordType,
		AccessKey:          configopaque.String(args.AccessKey),
		HTTPServerSettings: *args.HTTPServer.Convert(),
	}, nil
}

// Extensions implements receiver.Arguments.
func (args Arguments) Extensions() map[otelcomponent.ID]otelextension.Extension {
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
