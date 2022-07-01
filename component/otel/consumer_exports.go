package otel

import (
	"context"
	"sync"

	"github.com/grafana/agent/component"
	otelconsumer "go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/model/pdata"
)

func init() {
	component.RegisterGoStruct("Consumer", Consumer{})
}

// CombinedConsumer is a combined consumer of all telemetry types.
type CombinedConsumer interface {
	otelconsumer.Metrics
	otelconsumer.Logs
	otelconsumer.Traces
}

// Consumer is the registered Go struct used by components to pass around
// consumers.
type Consumer struct{ CombinedConsumer }

// ConsumerExports is a common Exports type for components which are processors
// or exporters.
type ConsumerExports struct {
	// Input is a collection of consumers that other components can use to send
	// telemetry data.
	Input *Consumer `hcl:"input,attr"`
}

// lazyCombinedConsumer is used by FlowExporter and FlowProcessor to expose
// their consumers as fast as possible even before components are constructed.
// Calls to process telemetry data will block while there is no active
// consumer.
type lazyCombinedConsumer struct {
	mut             sync.RWMutex
	metricsConsumer otelconsumer.Metrics
	logsConsumer    otelconsumer.Logs
	tracesConsumer  otelconsumer.Traces
}

var _ CombinedConsumer = (*lazyCombinedConsumer)(nil)

func newLazyCombinedConsumer() *lazyCombinedConsumer {
	return &lazyCombinedConsumer{}
}

func (c *lazyCombinedConsumer) Capabilities() otelconsumer.Capabilities {
	// TODO(rfratto): this is probably fairly inefficient over the upstream
	// collector and needs to be improved. As long as lazyCombinedConsumer is
	// always used in a context where MutatesData is constant (i.e., because the
	// consumers created by a component always have the same value for
	// MutatesData, we can pass the value here).
	return otelconsumer.Capabilities{
		MutatesData: true,
	}
}

func (c *lazyCombinedConsumer) ConsumeMetrics(ctx context.Context, md pdata.Metrics) error {
	c.mut.RLock()
	defer c.mut.RUnlock()
	if c.metricsConsumer == nil {
		// TODO(rfratto): should this log?
		return nil
	}
	return c.metricsConsumer.ConsumeMetrics(ctx, md)
}

func (c *lazyCombinedConsumer) ConsumeLogs(ctx context.Context, md pdata.Logs) error {
	c.mut.RLock()
	defer c.mut.RUnlock()
	if c.logsConsumer == nil {
		// TODO(rfratto): should this log?
		return nil
	}
	return c.logsConsumer.ConsumeLogs(ctx, md)
}

func (c *lazyCombinedConsumer) ConsumeTraces(ctx context.Context, md pdata.Traces) error {
	c.mut.RLock()
	defer c.mut.RUnlock()
	if c.tracesConsumer == nil {
		// TODO(rfratto): should this log?
		return nil
	}
	return c.tracesConsumer.ConsumeTraces(ctx, md)
}

func (c *lazyCombinedConsumer) SetConsumers(m otelconsumer.Metrics, l otelconsumer.Logs, t otelconsumer.Traces) {
	c.mut.Lock()
	defer c.mut.Unlock()

	c.metricsConsumer = m
	c.logsConsumer = l
	c.tracesConsumer = t
}
