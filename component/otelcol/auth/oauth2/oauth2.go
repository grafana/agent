package oauth2

import (
	"net/url"
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/otelcol"
	"github.com/grafana/agent/component/otelcol/auth"
	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/oauth2clientauthextension"
	otelcomponent "go.opentelemetry.io/collector/component"
	otelconfig "go.opentelemetry.io/collector/config"
)

func init() {
	component.Register(component.Registration{
		Name:    "otelcol.auth.oauth2",
		Args:    Arguments{},
		Exports: auth.Exports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			fact := oauth2clientauthextension.NewFactory()
			return auth.New(opts, fact, args.(Arguments))
		},
	})
}

// Arguments configures the otelcol.auth.oauth2 component.
type Arguments struct {
	ClientID       string                     `river:"client_id,attr"`
	ClientSecret   string                     `river:"client_secret,attr"`
	TokenURL       string                     `river:"token_url,attr"`
	EndpointParams url.Values                 `river:"endpoint_params,attr,optional"`
	Scopes         []string                   `river:"scopes,attr,optional"`
	TLSSetting     otelcol.TLSClientArguments `river:"tls,block,optional"`
	Timeout        time.Duration              `river:"timeout,attr,optional"`
}

var _ auth.Arguments = Arguments{}

// Convert implements auth.Arguments.
func (args Arguments) Convert() otelconfig.Extension {
	return &oauth2clientauthextension.Config{
		ExtensionSettings: otelconfig.NewExtensionSettings(otelconfig.NewComponentID("oauth2")),
		ClientID:          args.ClientID,
		ClientSecret:      args.ClientSecret,
		TokenURL:          args.TokenURL,
		EndpointParams:    args.EndpointParams,
		Scopes:            args.Scopes,
		TLSSetting:        *args.TLSSetting.Convert(),
		Timeout:           args.Timeout,
	}
}

// Extensions implements auth.Arguments.
func (args Arguments) Extensions() map[otelconfig.ComponentID]otelcomponent.Extension {
	return nil
}

// Exporters implements auth.Arguments.
func (args Arguments) Exporters() map[otelconfig.DataType]map[otelconfig.ComponentID]otelcomponent.Exporter {
	return nil
}
