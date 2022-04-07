package exporters

import (
	"context"
	"errors"

	"github.com/grafana/agent/pkg/integrations/v2/app_o11y_receiver/models"
	"github.com/grafana/agent/pkg/traces/pushreceiver"
)

// TracesExporter will send traces to a traces instance
type TracesExporter struct {
	factory *pushreceiver.Factory
}

// NewTracesExporter creates a trace exporter for the app o11y receiver.
func NewTracesExporter(pushReceiverFactory *pushreceiver.Factory) AppO11yReceiverExporter {
	return &TracesExporter{pushReceiverFactory}
}

// Name of the exporter, for logging purposes
func (te *TracesExporter) Name() string {
	return "traces exporter"
}

// Export implements the AppDataExporter interface
func (te *TracesExporter) Export(ctx context.Context, payload models.Payload) error {
	if payload.Traces == nil {
		return nil
	}
	if te.factory.Consumer != nil {
		return te.factory.Consumer.ConsumeTraces(ctx, payload.Traces.Traces)
	}
	return errors.New("push receiver factory consumer not initialized")
}
