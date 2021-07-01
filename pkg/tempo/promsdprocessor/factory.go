package promsdprocessor

import (
	"context"
	"fmt"

	prom_config "github.com/prometheus/prometheus/config"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor/processorhelper"
	"gopkg.in/yaml.v2"
)

// TypeStr is the unique identifier for the Prometheus SD processor.
const TypeStr = "prom_sd_processor"

// Config holds the configuration for the Prometheus SD processor.
type Config struct {
	config.ProcessorSettings `mapstructure:",squash"`
	ScrapeConfigs            []interface{} `mapstructure:"scrape_configs"`
}

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

	out, err := yaml.Marshal(oCfg.ScrapeConfigs)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal scrapeConfigs interface{} to yaml: %w", err)
	}

	scrapeConfigs := make([]*prom_config.ScrapeConfig, 0)
	err = yaml.Unmarshal(out, &scrapeConfigs)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal bytes to []*config.ScrapeConfig: %w", err)
	}

	return newTraceProcessor(nextConsumer, scrapeConfigs)
}
