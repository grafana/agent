package dummyreceiver

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configmodels"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver/receiverhelper"
)

const (
	// The value of "type" key in configuration.
	TypeStr = "dummy"
)

// NewFactory creates dummy receiver factory.
func NewFactory() component.ReceiverFactory {
	return receiverhelper.NewFactory(
		TypeStr,
		createDefaultConfig,
		receiverhelper.WithMetrics(createMetricsReceiver),
	)
}

// Config defines configuration for dummy receiver.
type Config struct {
	configmodels.Receiver `mapstructure:",squash"` // squash ensures fields are correctly decoded in embedded struct.
}

func createDefaultConfig() configmodels.Receiver {
	return &configmodels.ReceiverSettings{
		TypeVal: TypeStr,
		NameVal: TypeStr,
	}
}

// Dummy receiver is used in the metrics pipeline so we need to
// implement a metrics receiver.
func createMetricsReceiver(
	_ context.Context,
	_ component.ReceiverCreateParams,
	_ configmodels.Receiver,
	_ consumer.MetricsConsumer,
) (component.MetricsReceiver, error) {
	return newDummyReceiver(nil, nil, nil), nil
}
