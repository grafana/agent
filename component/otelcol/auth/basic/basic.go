// Package basic provides an otelcol.auth.basic component.
package basic

import (
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/otelcol/auth"
	"github.com/grafana/agent/pkg/flow/rivertypes"
	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/basicauthextension"
	otelcomponent "go.opentelemetry.io/collector/component"
)

func init() {
	component.Register(component.Registration{
		Name:    "otelcol.auth.basic",
		Args:    Arguments{},
		Exports: auth.Exports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			fact := basicauthextension.NewFactory()
			return auth.New(opts, fact, args.(Arguments))
		},
	})
}

// Arguments configures the otelcol.auth.basic component.
type Arguments struct {
	// TODO(rfratto): should we support htpasswd?

	Username string            `river:"username,attr"`
	Password rivertypes.Secret `river:"password,attr"`
}

var _ auth.Arguments = Arguments{}

// Convert implements auth.Arguments.
func (args Arguments) Convert() otelcomponent.Config {
	return &basicauthextension.Config{
		ExtensionSettings: otelcomponent.NewExtensionConfigSettings(otelcomponent.NewID("basic")),
		ClientAuth: &basicauthextension.ClientAuthSettings{
			Username: args.Username,
			Password: string(args.Password),
		},
	}
}

// Extensions implements auth.Arguments.
func (args Arguments) Extensions() map[otelcomponent.ID]otelcomponent.Extension {
	return nil
}

// Exporters implements auth.Arguments.
func (args Arguments) Exporters() map[otelcomponent.DataType]map[otelcomponent.ID]otelcomponent.Exporter {
	return nil
}
