package jaeger_remote_sampling

import (
	"fmt"
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/otelcol"
	"github.com/grafana/agent/component/otelcol/extension"
	"github.com/grafana/agent/pkg/river"
	otelcomponent "go.opentelemetry.io/collector/component"
	otelconfig "go.opentelemetry.io/collector/config"
)

const (
	DefaultHTTPEndpoint   = "0.0.0.0:5778"
	DefaultGRPCEndpoint   = "0.0.0.0:14250"
	DefaaultGRPCTransport = "tcp"
)

func init() {
	component.Register(component.Registration{
		Name: "otelcol.extension.jaeger_remote_sampling",
		Args: Arguments{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			fact := NewFactory()

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
	Content        string                       `river:"content,attr,optional"`
	Remote         *otelcol.GRPCClientArguments `river:"remote,block,optional"`
	File           string                       `river:"file,attr,optional"`
	ReloadInterval time.Duration                `river:"reload_interval,attr,optional"`
}

var (
	_ extension.Arguments = Arguments{}
	_ river.Unmarshaler   = (*Arguments)(nil)
)

// DefaultArguments holds default settings for otelcol.receiver.zipkin.
var DefaultArguments = Arguments{}

// Convert implements extension.Arguments.
func (args Arguments) Convert() otelconfig.Extension {
	return &Config{
		HTTPServerSettings: args.HTTP.Convert(),
		GRPCServerSettings: args.GRPC.Convert(),
		Source: Source{
			Remote:         args.Source.Remote.Convert(),
			File:           args.Source.File,
			ReloadInterval: args.Source.ReloadInterval,
			Contents:       args.Source.Content,
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
func (a *Arguments) UnmarshalRiver(f func(interface{}) error) error {
	*a = DefaultArguments

	type args Arguments
	err := f((*args)(a))
	if err != nil {
		return err
	}

	// remote config, local file and contents are all mutually exclusive
	sourcesSet := 0
	if a.Source.Content != "" {
		sourcesSet++
	}
	if a.Source.File != "" {
		sourcesSet++
	}
	if a.Source.Remote != nil {
		sourcesSet++
	}

	if sourcesSet == 0 {
		return fmt.Errorf("one of contents, file or remote must be configured")
	}
	if sourcesSet > 1 {
		return fmt.Errorf("only one of contents, file or remote can be configured")
	}

	if a.GRPC == nil && a.HTTP == nil {
		return fmt.Errorf("http or grpc must be configured to serve the sampling document")
	}

	// if the block exists but required fields aren't set, use defaults
	if a.GRPC != nil {
		if a.GRPC.Endpoint == "" {
			a.GRPC.Endpoint = DefaultGRPCEndpoint
		}
		if a.GRPC.Transport == "" {
			a.GRPC.Transport = DefaaultGRPCTransport
		}
	}
	if a.HTTP != nil && a.HTTP.Endpoint == "" {
		a.HTTP.Endpoint = DefaultHTTPEndpoint
	}

	return nil
}
