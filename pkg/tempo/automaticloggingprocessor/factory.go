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

	LoggingConfig *AutomaticLoggingConfig `mapstructure:"automatic_logging"`
}

// AutomaticLoggingConfig holds config information for automatic logging
type AutomaticLoggingConfig struct { // jpe moar options
	LokiName          string   `mapstructure:"loki_name" yaml:"loki_name"`
	EnableSpans       bool     `mapstructure:"enable_spans" yaml:"enable_spans"`
	EnableRoots       bool     `mapstructure:"enable_roots" yaml:"enable_roots"`
	EnableProcesses   bool     `mapstructure:"enable_processes" yaml:"enable_processes"`
	SpanAttributes    []string `mapstructure:"span_attributes" yaml:"span_attributes"`
	ProcessAttributes []string `mapstructure:"process_attributes" yaml:"process_attributes"`
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
	oCfg := cfg.(*Config)

	return newTraceProcessor(nextConsumer, oCfg.LoggingConfig)
}
