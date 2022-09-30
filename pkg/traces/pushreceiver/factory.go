package pushreceiver

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/consumer"
)

const (
	//TypeStr for push receiver.
	TypeStr = "push_receiver"
)

// Type returns the receiver type that PushReceiverFactory produces
func (f Factory) Type() config.Type {
	return TypeStr
}

// NewFactory creates a new push receiver factory.
func NewFactory() component.ReceiverFactory {
	return &Factory{}
}

// CreateDefaultConfig creates a default push receiver config.
func (f *Factory) CreateDefaultConfig() config.Receiver {
	s := config.NewReceiverSettings(config.NewComponentIDWithName(TypeStr, TypeStr))
	return &s
}

// Factory is a factory that sneakily exposes a Traces consumer for use within the agent.
type Factory struct {
	component.Factory
	Consumer consumer.Traces
}

// MetricsReceiverStability implements component.ReceiverFactory.
func (f *Factory) MetricsReceiverStability() component.StabilityLevel {
	return component.StabilityLevelUndefined
}

// LogsReceiverStability implements component.ReceiverFactory.
func (f *Factory) LogsReceiverStability() component.StabilityLevel {
	return component.StabilityLevelUndefined
}

// TracesReceiverStability implements component.ReceiverFactory.
func (f *Factory) TracesReceiverStability() component.StabilityLevel {
	return component.StabilityLevelUndefined
}

// CreateTracesReceiver creates a stub receiver while also sneakily keeping a reference to the provided Traces consumer.
func (f *Factory) CreateTracesReceiver(
	_ context.Context,
	_ component.ReceiverCreateSettings,
	_ config.Receiver,
	c consumer.Traces,
) (component.TracesReceiver, error) {

	r, err := newPushReceiver()
	f.Consumer = c

	return r, err
}

// CreateMetricsReceiver returns an error because metrics are not supported by push receiver.
func (f *Factory) CreateMetricsReceiver(ctx context.Context, set component.ReceiverCreateSettings,
	cfg config.Receiver, nextConsumer consumer.Metrics) (component.MetricsReceiver, error) {

	return nil, component.ErrDataTypeIsNotSupported
}

// CreateLogsReceiver returns an error because logs are not supported by push receiver.
func (f *Factory) CreateLogsReceiver(ctx context.Context, set component.ReceiverCreateSettings,
	cfg config.Receiver, nextConsumer consumer.Logs) (component.LogsReceiver, error) {

	return nil, component.ErrDataTypeIsNotSupported
}
