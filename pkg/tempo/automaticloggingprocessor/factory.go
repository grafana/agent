package automaticloggingprocessor

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configmodels"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor/processorhelper"
)

// TypeStr is the unique identifier for the Automatic Logging processor.
const TypeStr = "automatic_logging_processor"

// Config holds the configuration for the Automatic Logging processor.
type Config struct {
	configmodels.ProcessorSettings `mapstructure:",squash"`
}

// NewFactory returns a new factory for the Attributes processor.
func NewFactory() component.ProcessorFactory {
	return processorhelper.NewFactory(
		TypeStr,
		createDefaultConfig,
		processorhelper.WithTraces(createTraceProcessor),
	)
}

func createDefaultConfig() configmodels.Processor {
	return &Config{
		ProcessorSettings: configmodels.ProcessorSettings{
			TypeVal: TypeStr,
			NameVal: TypeStr,
		},
	}
}

func createTraceProcessor(
	_ context.Context,
	cp component.ProcessorCreateParams,
	cfg configmodels.Processor,
	nextConsumer consumer.TracesConsumer,
) (component.TracesProcessor, error) {

	return newTraceProcessor(nextConsumer)
}
