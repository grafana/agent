package gcpricingestimatorprocessor

import (
	"context"
	"sync"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor"
)

const (
	// Grafana Cloud Pricing Estimator
	typeStr   = "gcpricingestimator"
	stability = component.StabilityLevelDevelopment
)

var processorCapabilities = consumer.Capabilities{MutatesData: false}

type factory struct {
	lock sync.Mutex
}

func NewFactory() processor.Factory {
	return processor.NewFactory(
		typeStr,
		createDefaultConfig,
		processor.WithMetrics(createMetricsProcessor, stability),
		processor.WithTraces(createTracesProcessor, stability),
		processor.WithLogs(createLogsProcessor, stability),
	)
}

func createDefaultConfig() component.Config {
	return &Config{}
}

func createMetricsProcessor(_ context.Context, params processor.CreateSettings, cfg component.Config, next consumer.Metrics) (processor.Metrics, error) {
	p, err := newProcessor(params.Logger, cfg.(*Config))
	if err != nil {
		return nil, err
	}
	p.metricsConsumer = next
	return p, nil
}

func createTracesProcessor(_ context.Context, params processor.CreateSettings, cfg component.Config, next consumer.Traces) (processor.Traces, error) {
	p, err := newProcessor(params.Logger, cfg.(*Config))
	if err != nil {
		return nil, err
	}
	p.tracesConsumer = next
	return p, nil
}

func createLogsProcessor(_ context.Context, params processor.CreateSettings, cfg component.Config, next consumer.Logs) (processor.Logs, error) {
	p, err := newProcessor(params.Logger, cfg.(*Config))
	if err != nil {
		return nil, err
	}
	p.logsConsumer = next
	return p, nil
}
