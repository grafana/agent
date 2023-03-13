package jaegerremotesampling

import (
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/otelcol"
	"github.com/grafana/agent/component/otelcol/extension"
	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/jaegerremotesampling"
	otelcomponent "go.opentelemetry.io/collector/component"
	otelconfig "go.opentelemetry.io/collector/config"
)

func init() {
	component.Register(component.Registration{
		Name: "otelcol.extension.jaegerremotesampling",
		Args: Arguments{},
		// Exports: nil, jpe - export sampling json?

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			fact := jaegerremotesampling.NewFactory()

			return extension.New(opts, fact, args.(Arguments))
		},
	})
}

// Arguments configures the otelcol.auth.basic component.
type Arguments struct {
	// https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/extension/jaegerremotesampling/config.go#L42
	GRPC *otelcol.GRPCServerArguments `river:"grpc,block,optional"`
	HTTP *otelcol.HTTPServerArguments `river:"http,block,optional"`

	Source ArgumentsSource `river:"source,block"`
}

type ArgumentsSource struct {
	Remote         *otelcol.GRPCClientArguments `river:"remote,block,optional"`
	File           string                       `river:"file,attr,optional"`
	ReloadInterval time.Duration                `river:"reload_interval,attr,optional"`
}

var _ extension.Arguments = Arguments{}

// Convert implements extension.Arguments.
func (args Arguments) Convert() otelconfig.Extension {
	return &jaegerremotesampling.Config{
		HTTPServerSettings: args.HTTP.Convert(),
		GRPCServerSettings: args.GRPC.Convert(),
		Source: jaegerremotesampling.Source{
			Remote:         args.Source.Remote.Convert(),
			File:           args.Source.File,
			ReloadInterval: args.Source.ReloadInterval,
		},
	}
}

// Extensions implements extension.Arguments.
func (args Arguments) Extensions() map[otelconfig.ComponentID]otelcomponent.Extension {
	return nil
}

// Exporters implements extension.Arguments.
func (args Arguments) Exporters() map[otelconfig.DataType]map[otelconfig.ComponentID]otelcomponent.Exporter {
	return nil
}
