package pushreceiver

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver/receiverhelper"
)

const (
	typeStr = "push_receiver"
)

func NewFactory() component.ReceiverFactory {
	f := &pushReceiverFactory{}

	return receiverhelper.NewFactory(
		typeStr,
		createDefaultConfig,
		receiverhelper.WithTraces(f.createTracesReceiver),
	)
}

func createDefaultConfig() config.Receiver {
	return nil
}

type pushReceiverFactory struct {
	receiver consumer.Traces
}

func (f *pushReceiverFactory) createTracesReceiver(
	_ context.Context,
	_ component.ReceiverCreateSettings,
	_ config.Receiver,
	_ consumer.Traces,
) (component.TracesReceiver, error) {
	r, err := newPushReceiver()
	f.receiver = r.(consumer.Traces)

	return r, err
}
