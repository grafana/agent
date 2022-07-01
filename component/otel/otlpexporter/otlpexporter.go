package otlpexporter

import (
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/otel"
	"github.com/hashicorp/hcl/v2"
	"github.com/rfratto/gohcl"

	otelcomponent "go.opentelemetry.io/collector/component"
	otelconfig "go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/configcompression"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
	"go.opentelemetry.io/collector/exporter/otlpexporter"
)

func init() {
	component.Register(component.Registration{
		Name:    "otel.exporter_otlp",
		Args:    Arguments{},
		Exports: otel.ConsumerExports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			fact := otlpexporter.NewFactory()
			return otel.NewFlowExporter(opts, fact, args.(Arguments))
		},
	})
}

type (
	// Arguments configures the OTLP exporter.
	Arguments struct {
		Timeout        time.Duration               `hcl:"timeout,optional"`
		QueueSetttings *otel.ExporterQueueSettings `hcl:"sending_queue,block"`
		RetrySettings  *otel.ExporterRetrySettings `hcl:"retry_on_failure,block"`
		ClientSettings *otel.GRPCClientSettings    `hcl:"client,block"`

		// TODO(rfratto): ClientSettings should be squashed, if that were supported
		// by gohcl.
	}
)

var (
	_ otel.ExporterArguments = (*Arguments)(nil)
	_ gohcl.Decoder          = (*Arguments)(nil)
)

// Default values
var (
	DefaultArguments = Arguments{
		Timeout:        5 * time.Second,
		QueueSetttings: &otel.DefaultExporterQueueSettings,
		RetrySettings:  &otel.DefaultExporterRetrySettings,
		ClientSettings: &otel.GRPCClientSettings{
			Headers:         make(map[string]string),
			Compression:     configcompression.Gzip,
			WriteBufferSize: 512 * 1024,
		},
	}
)

// DecodeHCL implements gohcl.Decoder.
func (args *Arguments) DecodeHCL(body hcl.Body, ctx *hcl.EvalContext) error {
	*args = DefaultArguments
	type arguments Arguments
	return gohcl.DecodeBody(body, ctx, (*arguments)(args))
}

// Convert transforms OTLPReceiverArguments into the upstream otlpreceiver
// Config.
func (args Arguments) Convert() otelconfig.Exporter {
	return &otlpexporter.Config{
		ExporterSettings: otelconfig.NewExporterSettings(otelconfig.NewComponentID("otlp")),
		TimeoutSettings: exporterhelper.TimeoutSettings{
			Timeout: args.Timeout,
		},
		QueueSettings:      args.QueueSetttings.Convert(),
		RetrySettings:      args.RetrySettings.Convert(),
		GRPCClientSettings: args.ClientSettings.Convert(),
	}
}

func (args Arguments) Extensions() map[otelconfig.ComponentID]otelcomponent.Extension {
	return nil
}

func (args Arguments) Exporters() map[otelconfig.DataType]map[otelconfig.ComponentID]otelcomponent.Exporter {
	return nil
}
