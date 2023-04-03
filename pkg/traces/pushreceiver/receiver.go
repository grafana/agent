package pushreceiver

import (
	"context"

	"go.opentelemetry.io/collector/component"
)

type receiver struct{}

func (r *receiver) Start(_ context.Context, _ component.Host) error {
	return nil
}

func (r *receiver) Shutdown(_ context.Context) error {
	return nil
}
