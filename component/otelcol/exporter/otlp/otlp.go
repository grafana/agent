// Package otlp provides an otelcol.exporter.otlp component.
package otlp

import (
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/otelcol"
	"github.com/grafana/agent/component/otelcol/exporter"
	"github.com/grafana/agent/pkg/river"
	otelcomponent "go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configopaque"
	otelpexporterhelper "go.opentelemetry.io/collector/exporter/exporterhelper"
	"go.opentelemetry.io/collector/exporter/otlpexporter"
	otelextension "go.opentelemetry.io/collector/extension"
)

func init() {
	component.Register(component.Registration{
		Name:    "otelcol.exporter.otlp",
		Args:    Arguments{},
		Exports: otelcol.ConsumerExports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			fact := otlpexporter.NewFactory()
			return exporter.New(opts, fact, args.(Arguments))
		},
	})
}

// Arguments configures the otelcol.exporter.otlp component.
type Arguments struct {
	Timeout time.Duration `river:"timeout,attr,optional"`

	Queue otelcol.QueueArguments `river:"sending_queue,block,optional"`
	Retry otelcol.RetryArguments `river:"retry_on_failure,block,optional"`

	Client GRPCClientArguments `river:"client,block"`
}

var (
	_ river.Unmarshaler  = (*Arguments)(nil)
	_ exporter.Arguments = Arguments{}
)

// DefaultArguments holds default values for Arguments.
var DefaultArguments = Arguments{
	Timeout: otelcol.DefaultTimeout,
	Queue:   otelcol.DefaultQueueArguments,
	Retry:   otelcol.DefaultRetryArguments,
	Client:  DefaultGRPCClientArguments,
}

// UnmarshalRiver implements river.Unmarshaler.
func (args *Arguments) UnmarshalRiver(f func(interface{}) error) error {
	*args = DefaultArguments
	type arguments Arguments
	return f((*arguments)(args))
}

// Convert implements exporter.Arguments.
func (args Arguments) Convert() otelcomponent.Config {
	return &otlpexporter.Config{
		TimeoutSettings: otelpexporterhelper.TimeoutSettings{
			Timeout: args.Timeout,
		},
		QueueSettings:      *args.Queue.Convert(),
		RetrySettings:      *args.Retry.Convert(),
		GRPCClientSettings: *(*otelcol.GRPCClientArguments)(&args.Client).Convert(),
	}
}

// Extensions implements exporter.Arguments.
func (args Arguments) Extensions() map[otelcomponent.ID]otelextension.Extension {
	return (*otelcol.GRPCClientArguments)(&args.Client).Extensions()
}

// Exporters implements exporter.Arguments.
func (args Arguments) Exporters() map[otelcomponent.DataType]map[otelcomponent.ID]otelcomponent.Component {
	return nil
}

// GRPCClientArguments is used to configure otelcol.exporter.otlp with
// component-specific defaults.
type GRPCClientArguments otelcol.GRPCClientArguments

var _ river.Unmarshaler = (*GRPCClientArguments)(nil)

// DefaultGRPCClientArguments holds component-specific default settings for
// GRPCClientArguments.
var DefaultGRPCClientArguments = GRPCClientArguments{
	Headers:         map[string]configopaque.String{},
	Compression:     otelcol.CompressionTypeGzip,
	WriteBufferSize: 512 * 1024,
}

// UnmarshalRiver implements river.Unmarshaler and supplies defaults.
func (args *GRPCClientArguments) UnmarshalRiver(f func(interface{}) error) error {
	*args = DefaultGRPCClientArguments
	type arguments GRPCClientArguments
	return f((*arguments)(args))
}
