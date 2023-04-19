package pushreceiver

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	otelreceiver "go.opentelemetry.io/collector/receiver"
)

const (
	// The receiver type that PushReceiverFactory produces
	TypeStr = "push_receiver"
)

// Factory is a factory that sneakily exposes a Traces consumer for use within the agent.
type Factory struct {
	otelreceiver.Factory
	//TODO: Should I name the Config member if its name is the same as the type - Config?
	Config Config
}

// push_receiver sneakily exposes a Traces consumer for use within the agent.
type Config struct {
	Consumer consumer.Traces
}

// NewFactory creates a new push receiver factory.
func NewFactory() otelreceiver.Factory {
	return otelreceiver.NewFactory(
		TypeStr,
		func() component.Config { return &Config{} },
		//TODO: Delete the WithMetrics and WithLogs? They are not supported and throw errors anyway.
		otelreceiver.WithMetrics(createMetricsReceiver, component.StabilityLevelUndefined),
		otelreceiver.WithLogs(createLogsReceiver, component.StabilityLevelUndefined),
		otelreceiver.WithTraces(createTracesReceiver, component.StabilityLevelUndefined),
	)
}

// CreateTracesReceiver creates a stub receiver while also sneakily keeping a reference to the provided Traces consumer.
func createTracesReceiver(_ context.Context, _ otelreceiver.CreateSettings,
	cfg component.Config, c consumer.Traces) (otelreceiver.Traces, error) {

	oCfg := cfg.(*Config)

	oCfg.Consumer = c

	return &receiver{}, nil
}

// CreateMetricsReceiver returns an error because metrics are not supported by push receiver.
func createMetricsReceiver(ctx context.Context, set otelreceiver.CreateSettings,
	cfg component.Config, nextConsumer consumer.Metrics) (otelreceiver.Metrics, error) {

	return nil, component.ErrDataTypeIsNotSupported
}

// CreateLogsReceiver returns an error because logs are not supported by push receiver.
func createLogsReceiver(ctx context.Context, set otelreceiver.CreateSettings,
	cfg component.Config, nextConsumer consumer.Logs) (otelreceiver.Logs, error) {

	return nil, component.ErrDataTypeIsNotSupported
}
