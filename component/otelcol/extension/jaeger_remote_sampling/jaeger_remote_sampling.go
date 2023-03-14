package jaeger_remote_sampling

import (
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/otelcol"
	"github.com/grafana/agent/component/otelcol/extension"
	"github.com/grafana/agent/pkg/river"
	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/jaegerremotesampling"
	otelcomponent "go.opentelemetry.io/collector/component"
	otelconfig "go.opentelemetry.io/collector/config"
)

func init() {
	component.Register(component.Registration{
		Name: "otelcol.extension.jaeger_remote_sampling",
		Args: Arguments{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			fact := jaegerremotesampling.NewFactory()

			return extension.New(opts, fact, args.(Arguments))
		},
	})
}

// Arguments configures the otelcol.extension.jaegerremotesampling component.
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

var (
	_ extension.Arguments = Arguments{}
	_ river.Unmarshaler   = (*Arguments)(nil)
)

// DefaultArguments holds default settings for otelcol.receiver.zipkin.
var DefaultArguments = Arguments{
	HTTP: &otelcol.HTTPServerArguments{
		Endpoint: "0.0.0.0:5778",
	},
	GRPC: &otelcol.GRPCServerArguments{
		Endpoint:  "0.0.0.0:14250",
		Transport: "tcp",
	},
}

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

// UnmarshalRiver applies defaults to args before unmarshaling.
func (args *Arguments) UnmarshalRiver(f func(interface{}) error) error {
	*args = DefaultArguments

	type arguments Arguments
	return f((*arguments)(args))
}
