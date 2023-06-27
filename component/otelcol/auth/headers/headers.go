// Package headers provides an otelcol.auth.headers component.
package headers

import (
	"fmt"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/otelcol/auth"
	"github.com/grafana/agent/pkg/river/rivertypes"
	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/headerssetterextension"
	otelcomponent "go.opentelemetry.io/collector/component"
	otelconfig "go.opentelemetry.io/collector/config"
)

func init() {
	component.Register(component.Registration{
		Name:    "otelcol.auth.headers",
		Args:    Arguments{},
		Exports: auth.Exports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			fact := headerssetterextension.NewFactory()
			return auth.New(opts, fact, args.(Arguments))
		},
	})
}

// Arguments configures the otelcol.auth.headers component.
type Arguments struct {
	Headers []Header `river:"header,block,optional"`
}

var _ auth.Arguments = Arguments{}

// Convert implements auth.Arguments.
func (args Arguments) Convert() (otelconfig.Extension, error) {
	var upstreamHeaders []headerssetterextension.HeaderConfig
	for _, h := range args.Headers {
		upstreamHeader := headerssetterextension.HeaderConfig{
			Key: &h.Key,
		}

		if h.Value != nil {
			upstreamHeader.Value = &h.Value.Value
		}
		if h.FromContext != nil {
			upstreamHeader.FromContext = h.FromContext
		}

		upstreamHeaders = append(upstreamHeaders, upstreamHeader)
	}

	return &headerssetterextension.Config{
		ExtensionSettings: otelconfig.NewExtensionSettings(otelconfig.NewComponentID("headers")),
		HeadersConfig:     upstreamHeaders,
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

// Header is an individual Header to send along with requests.
type Header struct {
	Key         string                     `river:"key,attr"`
	Value       *rivertypes.OptionalSecret `river:"value,attr,optional"`
	FromContext *string                    `river:"from_context,attr,optional"`
}

// Validate implements river.Validator.
func (h *Header) Validate() error {
	switch {
	case h.Key == "":
		return fmt.Errorf("key must be set to a non-empty string")
	case h.FromContext == nil && h.Value == nil:
		return fmt.Errorf("either value or from_context must be provided")
	case h.FromContext != nil && h.Value != nil:
		return fmt.Errorf("either value or from_context must be provided, not both")
	}

	return nil
}
