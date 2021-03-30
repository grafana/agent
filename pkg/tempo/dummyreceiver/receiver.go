package dummyreceiver

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.uber.org/zap"
)

type dummyReceiver struct{}

// New creates a dummy receiver.
func newDummyReceiver(_ *zap.Logger, _ *Config, _ consumer.MetricsConsumer) *dummyReceiver {
	return &dummyReceiver{}
}

// Start implements the Component interface.
func (r *dummyReceiver) Start(_ context.Context, _ component.Host) error {
	return nil
}

// Shutdown implements the Component interface.
func (r *dummyReceiver) Shutdown(context.Context) error {
	return nil
}
