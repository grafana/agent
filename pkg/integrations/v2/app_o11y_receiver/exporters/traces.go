package exporters

import (
	"context"

	"github.com/grafana/agent/pkg/integrations/v2/app_o11y_receiver/models"
	"go.opentelemetry.io/collector/consumer"
)

// TracesConsumerGetter returns a traces consumer to push traces to
type TracesConsumerGetter func() (consumer.Traces, error)

// TracesExporter will send traces to a traces instance
type TracesExporter struct {
	getTracesConsumer TracesConsumerGetter
}

// NewTracesExporter creates a trace exporter for the app o11y receiver.
func NewTracesExporter(getTracesConsumer TracesConsumerGetter) AppO11yReceiverExporter {
	return &TracesExporter{getTracesConsumer}
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
	consumer, err := te.getTracesConsumer()
	if err != nil {
		return err
	}
	return consumer.ConsumeTraces(ctx, payload.Traces.Traces)
}
