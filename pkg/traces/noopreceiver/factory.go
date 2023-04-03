package noopreceiver

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
)

const (
	// TypeStr for noop receiver.
	TypeStr = "noop"
)

// NewFactory creates noop receiver factory.
func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		TypeStr,
		createDefaultConfig,
		receiver.WithMetrics(createMetricsReceiver, component.StabilityLevelUndefined),
	)
}

// Config defines configuration for noop receiver.
type Config struct {
	component.Config `mapstructure:",squash"` // squash ensures fields are correctly decoded in embedded struct.
}

func createDefaultConfig() component.Config {
	return &Config{}
}

// noop receiver is used in the metrics pipeline so we need to
// implement a metrics receiver.
func createMetricsReceiver(
	_ context.Context,
	_ receiver.CreateSettings,
	_ component.Config,
	_ consumer.Metrics,
) (receiver.Metrics, error) {
	return newNoopReceiver(nil, nil, nil), nil
}
