package gcpricingestimatorprocessor

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/processor"
	"go.uber.org/zap"
)

var (
	_ processor.Metrics = (*estimator)(nil)
)

type estimator struct {
	logger *zap.Logger
	config Config
	done   chan struct{}

	metricsConsumer consumer.Metrics
	tracesConsumer  consumer.Traces
	logsConsumer    consumer.Logs
}

func newProcessor(logger *zap.Logger, cfg *Config) (*estimator, error) {
	logger.Info("Building gcpricingestimator")
	return &estimator{
		logger: logger,
		config: *cfg,
	}, nil
}

// Start implements processor.Metrics
func (e *estimator) Start(ctx context.Context, host component.Host) error {
	e.logger.Info("Starting gcpricingestimator")
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-e.done:
				return
			}
		}
	}()
	return nil
}

// Shutdown implements processor.Metrics
func (e *estimator) Shutdown(ctx context.Context) error {
	e.logger.Info("Stopping gcpricingestimator")
	e.done <- struct{}{}
	return nil
}

// Capabilities implements processor.Metrics
func (*estimator) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: false}
}

func (e *estimator) ConsumeMetrics(ctx context.Context, metrics pmetric.Metrics) error {
	// TODO: count active series
	return e.metricsConsumer.ConsumeMetrics(ctx, metrics)
}

func (e *estimator) ConsumeLogs(ctx context.Context, logs plog.Logs) error {
	// TODO: get byte count of logs
	return e.logsConsumer.ConsumeLogs(ctx, logs)
}

func (e *estimator) ConsumeTraces(ctx context.Context, traces ptrace.Traces) error {
	// TODO: get byte count of traces
	return e.tracesConsumer.ConsumeTraces(ctx, traces)
}
