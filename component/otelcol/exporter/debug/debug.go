package debug

import (
	"fmt"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/otelcol"
	"github.com/grafana/agent/component/otelcol/exporter"
	otelcomponent "go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configtelemetry"
	debugexporter "go.opentelemetry.io/collector/exporter/debugexporter"
	otelextension "go.opentelemetry.io/collector/extension"
)

func init() {
	component.Register(component.Registration{
		Name:    "otelcol.exporter.debug",
		Args:    Arguments{},
		Exports: otelcol.ConsumerExports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			fact := debugexporter.NewFactory()
			return exporter.New(opts, fact, args.(Arguments), exporter.TypeAll)
		},
	})
}

type Verbosity struct {
  Type string 
}

func (v Verbosity) convert() configtelemetry.Level {
  if (v.Type == "basic") {
    return configtelemetry.LevelBasic
  } else if (v.Type == "normal") {
    return configtelemetry.LevelNormal    
  } else if (v.Tyep == "detailed") {
    return configtelemetry.LevelDetailed
  } else {
    return fmt.Errorf("invalid type in verbosity %v", v)
  }

}

// Arguments configures the otelcol.exporter.debug component.
type Arguments struct {
	// Verbosity          configtelemetry.Level `river:"verbosity,attr,optional"`
  Verbosity string                         `river:"verbosity, attr, optional"`
	SamplingInitial    int                   `river:"sampling_initial,attr,optional"`
	SamplingThereafter int                   `river:"sampling_thereafter,attr,optional"`

  // DebugMetrics configures component internal metrics. Optional.
  _ DebugMetrics otelcol.DebugMetricsArguments `river: ";"`
}

var _ exporter.Arguments = Arguments{}

// DefaultArguments holds default values for Arguments.
var DefaultArguments = Arguments{
	Verbosity:          configtelemetry.LevelBasic,
	SamplingInitial:    2,
	SamplingThereafter: 500,
}

// SetToDefault implements river.Defaulter.
func (args *Arguments) SetToDefault() {
	*args = DefaultArguments
}

// Convert implements exporter.Arguments.
func (args Arguments) Convert() (otelcomponent.Config, error) {
	return &debugexporter.Config{
		Verbosity:          args.Verbosity,
		SamplingInitial:    args.SamplingInitial,
		SamplingThereafter: args.SamplingInitial,
	}, nil
}

// Extensions implements exporter.Arguments.
func (args Arguments) Extensions() map[otelcomponent.ID]otelextension.Extension {
	return nil
}

// Exporters implements exporter.Arguments.
func (args Arguments) Exporters() map[otelcomponent.DataType]map[otelcomponent.ID]otelcomponent.Component {
	return nil
}
