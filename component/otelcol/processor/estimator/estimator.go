package estimator

import (
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/otelcol"
	"github.com/grafana/agent/component/otelcol/processor"
	"github.com/grafana/agent/component/otelcol/processor/estimator/gcpricingestimatorprocessor"
	otelcomponent "go.opentelemetry.io/collector/component"
	otelconfig "go.opentelemetry.io/collector/config"
)

func init() {
	component.Register(component.Registration{
		Name:      "otelcol.processor.gcpricingestimator",
		Singleton: false,
		Args:      Arguments{},
		Exports:   otelcol.ConsumerExports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {

			return nil, nil
		},
	})
}

var (
	_ processor.Arguments = Arguments{}
)

type Arguments struct {
	// Output configures where to send processed data. Required.
	Output *otelcol.ConsumerArguments `river:"output,block"`
}

// Convert implements processor.Arguments
func (Arguments) Convert() otelconfig.Processor {
	return &gcpricingestimatorprocessor.Config{}
}

// Exporters implements processor.Arguments
func (Arguments) Exporters() map[otelconfig.Type]map[otelconfig.ComponentID]otelcomponent.Exporter {
	panic("unimplemented")
}

// Extensions implements processor.Arguments
func (Arguments) Extensions() map[otelconfig.ComponentID]otelcomponent.Extension {
	panic("unimplemented")
}

// NextConsumers implements processor.Arguments
func (Arguments) NextConsumers() *otelcol.ConsumerArguments {
	panic("unimplemented")
}

type otlpEstimator struct{}
