package pushreceiver

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/model/pdata"
)

type receiver struct{}

func (r *receiver) Start(_ context.Context, _ component.Host) error {
	return nil
}

func (r *receiver) Shutdown(_ context.Context) error {
	return nil
}

func (r *receiver) ConsumeTraces(_ context.Context, _ pdata.Traces) error {
	return nil
}

func newPushReceiver() (component.TracesReceiver, error) {
	return &receiver{}, nil
}
