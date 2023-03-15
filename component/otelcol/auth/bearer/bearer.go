// Package bearer provides an otelcol.auth.bearer component.
package bearer

import (
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/otelcol/auth"
	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/bearertokenauthextension"
	otelcomponent "go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configopaque"
	otelextension "go.opentelemetry.io/collector/extension"
)

func init() {
	component.Register(component.Registration{
		Name:    "otelcol.auth.bearer",
		Args:    Arguments{},
		Exports: auth.Exports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			fact := bearertokenauthextension.NewFactory()
			return auth.New(opts, fact, args.(Arguments))
		},
	})
}

// Arguments configures the otelcol.auth.bearer component.
type Arguments struct {
	Scheme   string              `river:"scheme,attr"`
	Token    configopaque.String `river:"token,attr"`
	Filename string              `river:"filename,attr"`
}

var _ auth.Arguments = Arguments{}

// Convert implements auth.Arguments.
func (args Arguments) Convert() otelcomponent.Config {
	return &bearertokenauthextension.Config{
		Scheme:      args.Scheme,
		BearerToken: args.Token,
		Filename:    args.Filename,
	}
}

// Extensions implements auth.Arguments.
func (args Arguments) Extensions() map[otelcomponent.ID]otelextension.Extension {
	return nil
}

// Exporters implements auth.Arguments.
func (args Arguments) Exporters() map[otelcomponent.DataType]map[otelcomponent.ID]otelcomponent.Component {
	return nil
}
