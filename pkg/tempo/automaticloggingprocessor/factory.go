package automaticloggingprocessor

import (
	"context"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor/processorhelper"
)

// TypeStr is the unique identifier for the Automatic Logging processor.
const TypeStr = "automatic_logging"

// Config holds the configuration for the Automatic Logging processor.
type Config struct {
	config.ProcessorSettings `mapstructure:",squash"`

	LoggingConfig *AutomaticLoggingConfig `mapstructure:"automatic_logging"`
}

// AutomaticLoggingConfig holds config information for automatic logging
type AutomaticLoggingConfig struct {
	Backend           string         `mapstructure:"backend" yaml:"backend"`
	LokiName          string         `mapstructure:"loki_name" yaml:"loki_name"`
	Spans             bool           `mapstructure:"spans" yaml:"spans"`
	Roots             bool           `mapstructure:"roots" yaml:"roots"`
	Processes         bool           `mapstructure:"processes" yaml:"processes"`
	SpanAttributes    []string       `mapstructure:"span_attributes" yaml:"span_attributes"`
	ProcessAttributes []string       `mapstructure:"process_attributes" yaml:"process_attributes"`
	Overrides         OverrideConfig `mapstructure:"overrides" yaml:"overrides"`
	Timeout           time.Duration  `mapstructure:"timeout" yaml:"timeout"`
}

// OverrideConfig contains overrides for various strings
type OverrideConfig struct {
	LokiTag     string `mapstructure:"loki_tag" yaml:"loki_tag"`
	ServiceKey  string `mapstructure:"service_key" yaml:"service_key"`
	SpanNameKey string `mapstructure:"span_name_key" yaml:"span_name_key"`
	StatusKey   string `mapstructure:"status_key" yaml:"status_key"`
	DurationKey string `mapstructure:"duration_key" yaml:"duration_key"`
	TraceIDKey  string `mapstructure:"trace_id_key" yaml:"trace_id_key"`
}

const (
	// BackendLoki is the backend config value for sending logs to a Loki pipeline
	BackendLoki = "loki"
	// BackendStdout is the backend config value for sending logs to stdout
	BackendStdout = "stdout"
)

// NewFactory returns a new factory for the Attributes processor.
func NewFactory() component.ProcessorFactory {
	return processorhelper.NewFactory(
		TypeStr,
		createDefaultConfig,
		processorhelper.WithTraces(createTraceProcessor),
	)
}

func createDefaultConfig() config.Processor {
	return &Config{
		ProcessorSettings: config.NewProcessorSettings(config.NewIDWithName(TypeStr, TypeStr)),
	}
}

func createTraceProcessor(
	_ context.Context,
	cp component.ProcessorCreateSettings,
	cfg config.Processor,
	nextConsumer consumer.Traces,
) (component.TracesProcessor, error) {
	oCfg := cfg.(*Config)

	return newTraceProcessor(nextConsumer, oCfg.LoggingConfig)
}
