// Package basic provides an otelcol.auth.basic component.
package basic

import (
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/otelcol/auth"
	"github.com/grafana/agent/pkg/flow/rivertypes"
	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/basicauthextension"
	otelcomponent "go.opentelemetry.io/collector/component"
	otelconfig "go.opentelemetry.io/collector/config"
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
func (args Arguments) Convert() (otelconfig.Extension, error) {
	return &basicauthextension.Config{
		ExtensionSettings: otelconfig.NewExtensionSettings(otelconfig.NewComponentID("basic")),
		ClientAuth: &basicauthextension.ClientAuthSettings{
			Username: args.Username,
			Password: string(args.Password),
		},
	}, nil
}

// Extensions implements auth.Arguments.
func (args Arguments) Extensions() map[otelconfig.ComponentID]otelcomponent.Extension {
	return nil
}

// Exporters implements auth.Arguments.
func (args Arguments) Exporters() map[otelconfig.DataType]map[otelconfig.ComponentID]otelcomponent.Exporter {
	return nil
}
