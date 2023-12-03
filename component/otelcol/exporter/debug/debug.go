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

// Arguments configures the otelcol.exporter.debug component.
type exporterArguments struct {
	// Verbosity          configtelemetry.Level `river:"verbosity,attr,optional"`
	Verbosity          configtelemetry.Level `river:"verbosity, attr, optional"`
	SamplingInitial    int                   `river:"sampling_initial,attr,optional"`
	SamplingThereafter int                   `river:"sampling_thereafter,attr,optional"`

	// DebugMetrics configures component internal metrics. Optional.
	DebugMetrics otelcol.DebugMetricsArguments `river:"debug_metrics,block,optional"`
}

type Arguments struct {
	// Verbosity          configtelemetry.Level `river:"verbosity,attr,optional"`
	Verbosity          string `river:"verbosity, attr, optional"`
	SamplingInitial    int    `river:"sampling_initial,attr,optional"`
	SamplingThereafter int    `river:"sampling_thereafter,attr,optional"`

	// DebugMetrics configures component internal metrics. Optional.
	// DebugMetrics otelcol.DebugMetricsArguments `river:"debug_metrics,block,optional"`
}

func (args Arguments) convertToExporter() (exporterArguments, error) {
	const exporterVerbosity = map[string]configtelemetry.Level{
		"basic":    configtelemetry.LevelBasic,
		"normal":   configtelemetry.LevelNormal,
		"detailed": configtelemetry.LevelDetailed,
	}

	if _, ok := exporterVerbosity[args.Verbosity]; !ok {
		return exporterArguments{}, fmt.Errorf("Invalid verbosity %q", args.Verbosity)
	}

	e := &exporterArguments{
		Verbosity:          args.Verbosity,
		SamplingInitial:    args.SamplingInitial,
		SamplingThereafter: args.SamplingThereafter,
	}

	return *e, nil
}

var _ exporter.Arguments = Arguments{}

// DefaultArguments holds default values for Arguments.
var DefaultArguments = Arguments{
	Verbosity:          "basic",
	SamplingInitial:    2,
	SamplingThereafter: 500,
}

// SetToDefault implements river.Defaulter.
func (args *Arguments) SetToDefault() {
	*args = DefaultArguments
}

// Convert implements exporter.Arguments.
func (args Arguments) Convert() (otelcomponent.Config, error) {
	exporterArgs, err := args.convertToExporter()
	if err != nil {
		return nil, fmt.Errorf("Error in conversion to config arguments, %v", err)
	}

	return &debugexporter.Config{
		Verbosity:          exporterArgs.Verbosity,
		SamplingInitial:    exporterArgs.SamplingInitial,
		SamplingThereafter: exporterArgs.SamplingInitial,
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

// DebugMetricsConfig implements receiver.Arguments.
func (args Arguments) DebugMetricsConfig() otelcol.DebugMetricsArguments {
	exporterArgs, _ := args.convertToExporter()
	return exporterArgs.DebugMetrics
}
