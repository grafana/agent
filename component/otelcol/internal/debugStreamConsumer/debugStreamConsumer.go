package debugstreamconsumer

import (
	"context"

	"github.com/grafana/agent/component/otelcol"
	otelconsumer "go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

type Consumer struct {
	debugStreamCallback func() func(string)
	logsMarshaler       plog.Marshaler
	metricsMarshaler    pmetric.Marshaler
	tracesMarshaler     ptrace.Marshaler
}

var _ otelcol.Consumer = (*Consumer)(nil)

func New(debugStreamCallback func() func(string)) *Consumer {
	return &Consumer{
		debugStreamCallback: debugStreamCallback,
		logsMarshaler:       NewTextLogsMarshaler(),
		metricsMarshaler:    NewTextMetricsMarshaler(),
		tracesMarshaler:     NewTextTracesMarshaler(),
	}
}

// Capabilities implements otelcol.Consumer.
func (c *Consumer) Capabilities() otelconsumer.Capabilities {
	// streaming data should not modify the value
	return otelconsumer.Capabilities{MutatesData: false}
}

// ConsumeTraces implements otelcol.ConsumeTraces.
func (c *Consumer) ConsumeTraces(ctx context.Context, td ptrace.Traces) error {
	if cb := c.debugStreamCallback(); cb != nil {
		data, _ := c.tracesMarshaler.MarshalTraces(td)
		cb(string(data))
	}
	return nil
}

// ConsumeMetrics implements otelcol.ConsumeMetrics.
func (c *Consumer) ConsumeMetrics(ctx context.Context, md pmetric.Metrics) error {
	if cb := c.debugStreamCallback(); cb != nil {
		data, _ := c.metricsMarshaler.MarshalMetrics(md)
		cb(string(data))
	}
	return nil
}

// ConsumeLogs implements otelcol.ConsumeLogs.
func (c *Consumer) ConsumeLogs(ctx context.Context, md plog.Logs) error {
	if cb := c.debugStreamCallback(); cb != nil {
		data, _ := c.logsMarshaler.MarshalLogs(md)
		cb(string(data))
	}
	return nil
}
