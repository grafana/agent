package servicegraphprocessor

import (
	"context"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor/processorhelper"
)

const (
	// TypeStr is the unique identifier for the Prometheus service graph exporter.
	TypeStr = "service_graphs"

	defaultWait     = time.Second * 10
	defaultMaxEdges = 10_000
)

// Config holds the configuration for the Prometheus SD processor.
type Config struct {
	config.ProcessorSettings `mapstructure:",squash"`

	wait     time.Duration `mapstructure:"wait"`
	maxEdges int           `mapstructure:"max_edges"`
}

// NewFactory returns a new factory for the Service Graphs processor.
func NewFactory() component.ProcessorFactory {
	return processorhelper.NewFactory(
		TypeStr,
		createDefaultConfig,
		processorhelper.WithTraces(createTracesProcessor),
	)
}

func createDefaultConfig() config.Processor {
	return &Config{
		ProcessorSettings: config.NewProcessorSettings(config.NewIDWithName(TypeStr, TypeStr)),
	}
}

func createTracesProcessor(
	_ context.Context,
	_ component.ProcessorCreateSettings,
	cfg config.Processor,
	nextConsumer consumer.Traces,
) (component.TracesProcessor, error) {
	eCfg := cfg.(*Config)

	return newProcessor(nextConsumer, eCfg)
}
