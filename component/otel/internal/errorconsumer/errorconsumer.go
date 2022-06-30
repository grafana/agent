// Package errorconsumer exposes consumers which always fail.
package errorconsumer

import (
	"context"
	"fmt"

	otelconsumer "go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/model/pdata"
)

var (
	Metrics otelconsumer.Metrics = errorConsumer{}
	Logs    otelconsumer.Logs    = errorConsumer{}
	Traces  otelconsumer.Traces  = errorConsumer{}
)

type errorConsumer struct{}

func (errorConsumer) Capabilities() otelconsumer.Capabilities {
	return otelconsumer.Capabilities{MutatesData: false}
}

func (errorConsumer) ConsumeMetrics(context.Context, pdata.Metrics) error {
	return fmt.Errorf("component does not accept metrics")
}

func (errorConsumer) ConsumeLogs(context.Context, pdata.Logs) error {
	return fmt.Errorf("component does not accept logs")
}

func (errorConsumer) ConsumeTraces(context.Context, pdata.Traces) error {
	return fmt.Errorf("component does not accept traces")
}
