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
	debugStreamCallback func(computeDataFunc func() string)
	logsMarshaler       plog.Marshaler
	metricsMarshaler    pmetric.Marshaler
	tracesMarshaler     ptrace.Marshaler
	isActive            bool
}

var _ otelcol.Consumer = (*Consumer)(nil)

func New() *Consumer {
	return &Consumer{
		debugStreamCallback: func(computeDataFunc func() string) {},
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
	if c.isActive {
		data, _ := c.tracesMarshaler.MarshalTraces(td)
		c.debugStreamCallback(func() string { return string(data) })
	}
	return nil
}

// ConsumeMetrics implements otelcol.ConsumeMetrics.
func (c *Consumer) ConsumeMetrics(ctx context.Context, md pmetric.Metrics) error {
	if c.isActive {
		data, _ := c.metricsMarshaler.MarshalMetrics(md)
		c.debugStreamCallback(func() string { return string(data) })
	}
	return nil
}

// ConsumeLogs implements otelcol.ConsumeLogs.
func (c *Consumer) ConsumeLogs(ctx context.Context, md plog.Logs) error {
	if c.isActive {
		data, _ := c.logsMarshaler.MarshalLogs(md)
		c.debugStreamCallback(func() string { return string(data) })
	}
	return nil
}

func (c *Consumer) HookDebugStream(active bool, debugStreamCallback func(computeDataFunc func() string)) {
	c.debugStreamCallback = debugStreamCallback
	c.isActive = active
}
