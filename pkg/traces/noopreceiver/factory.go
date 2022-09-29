package noopreceiver

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/consumer"
)

const (
	// TypeStr for noop receiver.
	TypeStr = "noop"
)

// NewFactory creates noop receiver factory.
func NewFactory() component.ReceiverFactory {
	return component.NewReceiverFactory(
		TypeStr,
		createDefaultConfig,
		component.WithMetricsReceiver(createMetricsReceiver, component.StabilityLevelUndefined),
	)
}

// Config defines configuration for noop receiver.
type Config struct {
	config.Receiver `mapstructure:",squash"` // squash ensures fields are correctly decoded in embedded struct.
}

func createDefaultConfig() config.Receiver {
	s := config.NewReceiverSettings(config.NewComponentIDWithName(TypeStr, TypeStr))
	return &s
}

// noop receiver is used in the metrics pipeline so we need to
// implement a metrics receiver.
func createMetricsReceiver(
	_ context.Context,
	_ component.ReceiverCreateSettings,
	_ config.Receiver,
	_ consumer.Metrics,
) (component.MetricsReceiver, error) {

	return newNoopReceiver(nil, nil, nil), nil
}
