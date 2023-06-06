// Package otlp provides an otelcol.exporter.otlp component.
package otlp

import (
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/otelcol"
	"github.com/grafana/agent/component/otelcol/exporter"
	otelcomponent "go.opentelemetry.io/collector/component"
	otelconfig "go.opentelemetry.io/collector/config"
	otelpexporterhelper "go.opentelemetry.io/collector/exporter/exporterhelper"
	"go.opentelemetry.io/collector/exporter/otlpexporter"
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
	_ exporter.Arguments = Arguments{}
)

// DefaultArguments holds default values for Arguments.
var DefaultArguments = Arguments{
	Timeout: otelcol.DefaultTimeout,
	Queue:   otelcol.DefaultQueueArguments,
	Retry:   otelcol.DefaultRetryArguments,
	Client:  DefaultGRPCClientArguments,
}

// SetToDefault implements river.Defaulter.
func (args *Arguments) SetToDefault() {
	*args = DefaultArguments
}

// Convert implements exporter.Arguments.
func (args Arguments) Convert() (otelconfig.Exporter, error) {
	return &otlpexporter.Config{
		ExporterSettings: otelconfig.NewExporterSettings(otelconfig.NewComponentID("otlp")),
		TimeoutSettings: otelpexporterhelper.TimeoutSettings{
			Timeout: args.Timeout,
		},
		QueueSettings:      *args.Queue.Convert(),
		RetrySettings:      *args.Retry.Convert(),
		GRPCClientSettings: *(*otelcol.GRPCClientArguments)(&args.Client).Convert(),
	}, nil
}

// Extensions implements exporter.Arguments.
func (args Arguments) Extensions() map[otelconfig.ComponentID]otelcomponent.Extension {
	return (*otelcol.GRPCClientArguments)(&args.Client).Extensions()
}

// Exporters implements exporter.Arguments.
func (args Arguments) Exporters() map[otelconfig.DataType]map[otelconfig.ComponentID]otelcomponent.Exporter {
	return nil
}

// GRPCClientArguments is used to configure otelcol.exporter.otlp with
// component-specific defaults.
type GRPCClientArguments otelcol.GRPCClientArguments

// DefaultGRPCClientArguments holds component-specific default settings for
// GRPCClientArguments.
var DefaultGRPCClientArguments = GRPCClientArguments{
	Headers:         map[string]string{},
	Compression:     otelcol.CompressionTypeGzip,
	WriteBufferSize: 512 * 1024,
}

// SetToDefault implements river.Defaulter.
func (args *GRPCClientArguments) SetToDefault() {
	*args = DefaultGRPCClientArguments
}
