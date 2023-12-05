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

type Arguments struct {
	Verbosity          string `river:"verbosity, attr, optional"`
	SamplingInitial    int    `river:"sampling_initial,attr, optional"`
	SamplingThereafter int    `river:"sampling_thereafter,attr, optional"`
}

func (args Arguments) convertVerbosity() (configtelemetry.Level, error) {
	var verbosity configtelemetry.Level
	switch args.Verbosity {
	case "basic":
		verbosity = configtelemetry.LevelBasic
	case "normal":
		verbosity = configtelemetry.LevelNormal
	case "detailed":
		verbosity = configtelemetry.LevelDetailed
	default:
		// Invalid verbosity
		// debugexporter only supports basic, normal and detailed levels
		return verbosity, fmt.Errorf("invalid verbosity %q", args.Verbosity)
	}

	return verbosity, nil
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
	verbosity, err := args.convertVerbosity()
	if err != nil {
		return nil, fmt.Errorf("error in conversion to config arguments, %v", err)
	}

	return &debugexporter.Config{
		Verbosity:          verbosity,
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

// DebugMetricsConfig implements receiver.Arguments.
func (args Arguments) DebugMetricsConfig() otelcol.DebugMetricsArguments {
	var debugMetrics otelcol.DebugMetricsArguments
	return debugMetrics
}
