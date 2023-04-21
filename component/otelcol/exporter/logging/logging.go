// Package logging provides an otelcol.exporter.logging component.
package logging

import (
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/otelcol"
	"github.com/grafana/agent/component/otelcol/exporter"
	"github.com/grafana/agent/pkg/river"
	otelcomponent "go.opentelemetry.io/collector/component"
	otelconfig "go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/configtelemetry"
	loggingexporter "go.opentelemetry.io/collector/exporter/loggingexporter"
)

func init() {
	component.Register(component.Registration{
		Name:    "otelcol.exporter.logging",
		Args:    Arguments{},
		Exports: otelcol.ConsumerExports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			fact := loggingexporter.NewFactory()
			return exporter.New(opts, fact, args.(Arguments))
		},
	})
}

// Arguments configures the otelcol.exporter.logging component.
type Arguments struct {
	Verbosity          configtelemetry.Level `river:"verbosity,attr,optional"`
	SamplingInitial    int                   `river:"sampling_initial,attr,optional"`
	SamplingThereafter int                   `river:"sampling_thereafter,attr,optional"`
}

var (
	_ river.Unmarshaler  = (*Arguments)(nil)
	_ exporter.Arguments = Arguments{}
)

// DefaultArguments holds default values for Arguments.
var DefaultArguments = Arguments{
	Verbosity:          configtelemetry.LevelNormal,
	SamplingInitial:    2,
	SamplingThereafter: 500,
}

// UnmarshalRiver implements river.Unmarshaler.
func (args *Arguments) UnmarshalRiver(f func(interface{}) error) error {
	*args = DefaultArguments
	type arguments Arguments
	return f((*arguments)(args))
}

// Convert implements exporter.Arguments.
func (args Arguments) Convert() (otelconfig.Exporter, error) {
	return &loggingexporter.Config{
		ExporterSettings:   otelconfig.NewExporterSettings(otelconfig.NewComponentID("logging")),
		Verbosity:          args.Verbosity,
		SamplingInitial:    args.SamplingInitial,
		SamplingThereafter: args.SamplingInitial,
	}, nil
}

// Extensions implements exporter.Arguments.
func (args Arguments) Extensions() map[otelconfig.ComponentID]otelcomponent.Extension {
	return nil
}

// Exporters implements exporter.Arguments.
func (args Arguments) Exporters() map[otelconfig.DataType]map[otelconfig.ComponentID]otelcomponent.Exporter {
	return nil
}
