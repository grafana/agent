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
// TODO: Update the docs with the new arguments
// TODO: Should we keep the "filename" attribute? Or omit it and have users use a separate flow component to read a file?
type Arguments struct {
	Scheme string `river:"scheme,attr,optional"`
	Token  string `river:"token,attr"`
	// Filename string `river:"filename,attr"`
}

var _ auth.Arguments = Arguments{}

// DefaultArguments holds default settings for Arguments.
var DefaultArguments = Arguments{
	Scheme: "Bearer",
}

// UnmarshalRiver implements river.Unmarshaler.
func (args *Arguments) UnmarshalRiver(f func(interface{}) error) error {
	*args = DefaultArguments

	type arguments Arguments
	if err := f((*arguments)(args)); err != nil {
		return err
	}

	return nil
}

// Convert implements auth.Arguments.
func (args Arguments) Convert() (otelcomponent.Config, error) {
	//TODO: I think it's no point in explicitly setting the default scheme to "bearer"? We can just leave it empty?
	return &bearertokenauthextension.Config{
		Scheme:      args.Scheme,
		BearerToken: configopaque.String(args.Token),
		//TODO: Delete this or enable it?
		// Filename:    args.Filename,
	}, nil
}

// Extensions implements auth.Arguments.
func (args Arguments) Extensions() map[otelcomponent.ID]otelextension.Extension {
	return nil
}

// Exporters implements auth.Arguments.
func (args Arguments) Exporters() map[otelcomponent.DataType]map[otelcomponent.ID]otelcomponent.Component {
	return nil
}
