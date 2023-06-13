package sigv4

import (
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/otelcol/auth"
	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/sigv4authextension"
	otelcomponent "go.opentelemetry.io/collector/component"
	otelconfig "go.opentelemetry.io/collector/config"
)

func init() {
	component.Register(component.Registration{
		Name:    "otelcol.auth.sigv4",
		Args:    Arguments{},
		Exports: auth.Exports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			fact := sigv4authextension.NewFactory()
			return auth.New(opts, fact, args.(Arguments))
		},
	})
}

// Arguments configures the otelcol.auth.sigv4 component.
type Arguments struct {
	Region     string     `river:"region,attr,optional"`
	Service    string     `river:"service,attr,optional"`
	AssumeRole AssumeRole `river:"assume_role,block,optional"`
}

var (
	_ auth.Arguments = Arguments{}
)

// Convert implements auth.Arguments.
func (args Arguments) Convert() (otelconfig.Extension, error) {
	res := sigv4authextension.Config{
		ExtensionSettings: otelconfig.NewExtensionSettings(otelconfig.NewComponentID("sigv4")),
		Region:            args.Region,
		Service:           args.Service,
		AssumeRole:        *args.AssumeRole.Convert(),
	}
	// sigv4authextension.Config has a private member called "credsProvider" which gets initialized when we call Validate().
	// If we don't call validate, the unit tests for this component will fail.
	if err := res.Validate(); err != nil {
		return nil, err
	}
	return &res, nil
}

// Validate implements river.Validator.
func (args Arguments) Validate() error {
	_, err := args.Convert()
	return err
}

// Extensions implements auth.Arguments.
func (args Arguments) Extensions() map[otelconfig.ComponentID]otelcomponent.Extension {
	return nil
}

// Exporters implements auth.Arguments.
func (args Arguments) Exporters() map[otelconfig.DataType]map[otelconfig.ComponentID]otelcomponent.Exporter {
	return nil
}

// AssumeRole replicates sigv4authextension.Config.AssumeRole
type AssumeRole struct {
	ARN         string `river:"arn,attr,optional"`
	SessionName string `river:"session_name,attr,optional"`
	STSRegion   string `river:"sts_region,attr,optional"`
}

// Convert converts args into the upstream type.
func (args *AssumeRole) Convert() *sigv4authextension.AssumeRole {
	if args == nil {
		return nil
	}

	return &sigv4authextension.AssumeRole{
		ARN:         args.ARN,
		SessionName: args.SessionName,
		STSRegion:   args.STSRegion,
	}
}
