package jaegerremotesampling

import (
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/otelcol"
	"github.com/grafana/agent/component/otelcol/auth"
	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/jaegerremotesampling"
	otelcomponent "go.opentelemetry.io/collector/component"
	otelconfig "go.opentelemetry.io/collector/config"
)

func init() {
	component.Register(component.Registration{
		Name: "otelcol.auth.basic",
		Args: Arguments{},
		// Exports: nil, jpe - export sampling json?

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			fact := jaegerremotesampling.NewFactory()

			return auth.New(opts, fact, args.(Arguments))
		},
	})
}

// Arguments configures the otelcol.auth.basic component.
type Arguments struct {
	// https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/extension/jaegerremotesampling/config.go#L42
	GRPC       otelcol.GRPCServerArguments `river:",squash"`
	ThriftHTTP otelcol.HTTPServerArguments `river:",squash"`

	Source ArgumentsSource `river:"source,block"`
}

type ArgumentsSource struct {
	Remote         otelcol.GRPCClientArguments `river:"remote,block,optional"`
	File           string                      `river:"file,attr,optional"`
	ReloadInterval time.Duration               `river:"reload_interval,attr,optional"`
}

var _ auth.Arguments = Arguments{} // jpe this is an auth thing. ditch?

// Convert implements auth.Arguments.
func (args Arguments) Convert() otelconfig.Extension {
	return &jaegerremotesampling.Config{} // jpe do
}

// Extensions implements auth.Arguments.
func (args Arguments) Extensions() map[otelconfig.ComponentID]otelcomponent.Extension {
	return nil
}

// Exporters implements auth.Arguments.
func (args Arguments) Exporters() map[otelconfig.DataType]map[otelconfig.ComponentID]otelcomponent.Exporter {
	return nil
}
