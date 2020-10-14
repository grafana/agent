package promsdprocessor

import (
	"context"
	"fmt"

	"github.com/prometheus/prometheus/config"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configmodels"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor/processorhelper"
	"gopkg.in/yaml.v2"
)

const TypeStr = "prom_sd_processor"

type Config struct {
	configmodels.ProcessorSettings `mapstructure:",squash"`
	ScrapeConfigs                  []interface{} `mapstructure:"scrape_configs"`
}

// NewFactory returns a new factory for the Attributes processor.
func NewFactory() component.ProcessorFactory {
	return processorhelper.NewFactory(
		TypeStr,
		createDefaultConfig,
		processorhelper.WithTraces(createTraceProcessor))
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
	_ component.ProcessorCreateParams,
	cfg configmodels.Processor,
	nextConsumer consumer.TraceConsumer,
) (component.TraceProcessor, error) {
	oCfg := cfg.(*Config)

	out, err := yaml.Marshal(oCfg.ScrapeConfigs)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal scrapeConfigs interface{} to yaml: %w", err)
	}

	scrapeConfigs := make([]*config.ScrapeConfig, 0)
	err = yaml.Unmarshal(out, &scrapeConfigs)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal bytes to []*config.ScrapeConfig: %w", err)
	}

	return newTraceProcessor(nextConsumer, scrapeConfigs)
}
