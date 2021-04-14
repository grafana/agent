package automaticloggingprocessor

import (
	"context"
	"fmt"

	"github.com/grafana/agent/pkg/loki"
	"github.com/grafana/agent/pkg/tempo/contextkeys"
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

	LokiName string `mapstructure:"loki_name"`
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
	ctx context.Context,
	cp component.ProcessorCreateParams,
	cfg configmodels.Processor,
	nextConsumer consumer.TracesConsumer,
) (component.TracesProcessor, error) {
	oCfg := cfg.(*Config)

	loki := ctx.Value(contextkeys.Loki).(*loki.Loki)
	if loki == nil {
		return nil, fmt.Errorf("key %s does not contain a Loki instance", contextkeys.Loki)
	}
	lokiInstance := loki.Instance(oCfg.LokiName)
	if lokiInstance == nil {
		return nil, fmt.Errorf("loki instance %s not found", oCfg.LokiName)
	}

	lokiChan := lokiInstance.Promtail().Client().Chan()
	return newTraceProcessor(nextConsumer, oCfg, lokiChan)
}
