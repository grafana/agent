package promsdprocessor

import (
	"context"
	"fmt"

	prom_config "github.com/prometheus/prometheus/config"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/consumer"
	"gopkg.in/yaml.v2"
)

// TypeStr is the unique identifier for the Prometheus SD processor.
const TypeStr = "prom_sd_processor"

const (
	// OperationTypeInsert inserts a new k/v if it isn't already present
	OperationTypeInsert = "insert"
	// OperationTypeUpdate only modifies an existing k/v
	OperationTypeUpdate = "update"
	// OperationTypeUpsert does both of above
	OperationTypeUpsert = "upsert"

	podAssociationIPLabel       = "ip"
	podAssociationOTelIPLabel   = "net.host.ip"
	podAssociationk8sIPLabel    = "k8s.pod.ip"
	podAssociationHostnameLabel = "hostname"
	podAssociationConnectionIP  = "connection"
)

// Config holds the configuration for the Prometheus SD processor.
type Config struct {
	config.ProcessorSettings `mapstructure:",squash"`
	ScrapeConfigs            []interface{} `mapstructure:"scrape_configs"`
	OperationType            string        `mapstructure:"operation_type"`
	PodAssociations          []string      `mapstructure:"pod_associations"`
}

// NewFactory returns a new factory for the Attributes processor.
func NewFactory() component.ProcessorFactory {
	return component.NewProcessorFactory(
		TypeStr,
		createDefaultConfig,
		component.WithTracesProcessor(createTraceProcessor, component.StabilityLevelUndefined),
	)
}

func createDefaultConfig() config.Processor {
	return &Config{
		ProcessorSettings: config.NewProcessorSettings(config.NewComponentIDWithName(TypeStr, TypeStr)),
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

	return newTraceProcessor(nextConsumer, oCfg.OperationType, oCfg.PodAssociations, scrapeConfigs)
}
