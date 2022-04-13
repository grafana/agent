package exporters

import (
	"context"
	"testing"

	"github.com/grafana/agent/pkg/traces/pushreceiver"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/model/pdata"
)

type mockTracesConsumer struct {
	consumed []pdata.Traces
}

func (c *mockTracesConsumer) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: false}
}

func (c *mockTracesConsumer) ConsumeTraces(ctx context.Context, td pdata.Traces) error {
	c.consumed = append(c.consumed, td)
	return nil
}

func Test_exportTraces_success(t *testing.T) {
	ctx := context.Background()
	factory := pushreceiver.NewFactory().(*pushreceiver.Factory)
	consumer := mockTracesConsumer{consumed: nil}
	_, err := factory.CreateTracesReceiver(ctx, component.ReceiverCreateSettings{}, nil, &consumer)
	require.NoError(t, err)
	exporter := NewTracesExporter(factory)
	payload := loadTestData(t)
	err = exporter.Export(ctx, payload)
	require.NoError(t, err)
	require.Len(t, consumer.consumed, 1)
}

func Test_exportTraces_noTracesInpayload(t *testing.T) {
	ctx := context.Background()
	factory := pushreceiver.NewFactory().(*pushreceiver.Factory)
	consumer := mockTracesConsumer{consumed: nil}
	_, err := factory.CreateTracesReceiver(ctx, component.ReceiverCreateSettings{}, nil, &consumer)
	require.NoError(t, err)
	exporter := NewTracesExporter(factory)
	payload := loadTestData(t)
	payload.Traces = nil
	err = exporter.Export(ctx, payload)
	require.NoError(t, err)
	require.Len(t, consumer.consumed, 0)
}

func Test_exportTraces_noConsumer(t *testing.T) {
	ctx := context.Background()
	factory := pushreceiver.NewFactory().(*pushreceiver.Factory)
	exporter := NewTracesExporter(factory)
	payload := loadTestData(t)
	err := exporter.Export(ctx, payload)
	require.Error(t, err, "push receiver factory consumer not initialized")
}
