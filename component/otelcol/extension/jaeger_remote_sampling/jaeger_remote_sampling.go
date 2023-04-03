package jaeger_remote_sampling

import (
	"fmt"
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/otelcol"
	"github.com/grafana/agent/component/otelcol/extension"
	"github.com/grafana/agent/component/otelcol/extension/jaeger_remote_sampling/internal/jaegerremotesampling"
	"github.com/grafana/agent/pkg/river"
	otelcomponent "go.opentelemetry.io/collector/component"
	otelextension "go.opentelemetry.io/collector/extension"
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

type (
	// GRPCServerArguments is used to configure otelcol.extension.jaeger_remote_sampling with
	// component-specific defaults.
	GRPCServerArguments otelcol.GRPCServerArguments

	// HTTPServerArguments is used to configure otelcol.extension.jaeger_remote_sampling with
	// component-specific defaults.
	HTTPServerArguments otelcol.HTTPServerArguments
)

var (
	_ river.Unmarshaler = (*GRPCServerArguments)(nil)
	_ river.Unmarshaler = (*HTTPServerArguments)(nil)
)

// Default server settings.
var (
	DefaultGRPCServerArguments = GRPCServerArguments{
		Endpoint:  "0.0.0.0:14250",
		Transport: "tcp",
	}

	DefaultHTTPServerArguments = HTTPServerArguments{
		Endpoint: "0.0.0.0:5778",
	}
)

// Arguments configures the otelcol.extension.jaegerremotesampling component.
type Arguments struct {
	GRPC *GRPCServerArguments `river:"grpc,block,optional"`
	HTTP *HTTPServerArguments `river:"http,block,optional"`

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

// Convert implements extension.Arguments.
func (args Arguments) Convert() (otelcomponent.Config, error) {
	return &jaegerremotesampling.Config{
		HTTPServerSettings: (*otelcol.HTTPServerArguments)(args.HTTP).Convert(),
		GRPCServerSettings: (*otelcol.GRPCServerArguments)(args.GRPC).Convert(),
		Source: jaegerremotesampling.Source{
			Remote:         args.Source.Remote.Convert(),
			File:           args.Source.File,
			ReloadInterval: args.Source.ReloadInterval,
			Contents:       args.Source.Content,
		},
	}, nil
}

// Extensions implements extension.Arguments.
func (args Arguments) Extensions() map[otelcomponent.ID]otelextension.Extension {
	return nil
}

// Exporters implements extension.Arguments.
func (args Arguments) Exporters() map[otelcomponent.DataType]map[otelcomponent.ID]otelcomponent.Component {
	return nil
}

// UnmarshalRiver applies defaults to args before unmarshaling.
func (a *Arguments) UnmarshalRiver(f func(interface{}) error) error {
	type args Arguments
	err := f((*args)(a))
	if err != nil {
		return err
	}

	if a.GRPC == nil && a.HTTP == nil {
		return fmt.Errorf("http or grpc must be configured to serve the sampling document")
	}

	return nil
}

func (a *ArgumentsSource) UnmarshalRiver(f func(interface{}) error) error {
	type args ArgumentsSource
	err := f((*args)(a))
	if err != nil {
		return err
	}

	// remote config, local file and contents are all mutually exclusive
	sourcesSet := 0
	if a.Content != "" {
		sourcesSet++
	}
	if a.File != "" {
		sourcesSet++
	}
	if a.Remote != nil {
		sourcesSet++
	}

	if sourcesSet == 0 {
		return fmt.Errorf("one of contents, file or remote must be configured")
	}
	if sourcesSet > 1 {
		return fmt.Errorf("only one of contents, file or remote can be configured")
	}

	return nil
}

// UnmarshalRiver implements river.Unmarshaler and supplies defaults.
func (args *GRPCServerArguments) UnmarshalRiver(f func(interface{}) error) error {
	*args = DefaultGRPCServerArguments
	type arguments GRPCServerArguments
	return f((*arguments)(args))
}

// UnmarshalRiver implements river.Unmarshaler and supplies defaults.
func (args *HTTPServerArguments) UnmarshalRiver(f func(interface{}) error) error {
	*args = DefaultHTTPServerArguments
	type arguments HTTPServerArguments
	return f((*arguments)(args))
}
