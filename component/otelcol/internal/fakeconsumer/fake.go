package fakeconsumer

import (
	"context"

	"github.com/grafana/agent/component/otelcol"
	otelconsumer "go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

type Consumer struct {
	CapabilitiesFunc   func() otelconsumer.Capabilities
	ConsumeTracesFunc  func(context.Context, ptrace.Traces) error
	ConsumeMetricsFunc func(context.Context, pmetric.Metrics) error
	ConsumeLogsFunc    func(context.Context, plog.Logs) error
}

var _ otelcol.Consumer = (*Consumer)(nil)

func (c *Consumer) Capabilities() otelconsumer.Capabilities {
	if c.CapabilitiesFunc != nil {
		return c.CapabilitiesFunc()
	}

	// We don't know what the fake implementation will do, so return true just
	// in case it mutates data.
	return otelconsumer.Capabilities{MutatesData: true}
}

func (c *Consumer) ConsumeTraces(ctx context.Context, td ptrace.Traces) error {
	if c.ConsumeTracesFunc != nil {
		return c.ConsumeTracesFunc(ctx, td)
	}
	return nil
}

func (c *Consumer) ConsumeMetrics(ctx context.Context, md pmetric.Metrics) error {
	if c.ConsumeMetricsFunc != nil {
		return c.ConsumeMetricsFunc(ctx, md)
	}
	return nil
}

func (c *Consumer) ConsumeLogs(ctx context.Context, md plog.Logs) error {
	if c.ConsumeLogsFunc != nil {
		return c.ConsumeLogsFunc(ctx, md)
	}
	return nil
}
