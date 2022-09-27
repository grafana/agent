package otelcol

import (
	otelconsumer "go.opentelemetry.io/collector/consumer"
)

// Consumer is a compbined OpenTelemetry Collector consumer which can consume
// any telemetry signal.
type Consumer interface {
	otelconsumer.Traces
	otelconsumer.Metrics
	otelconsumer.Logs
}

// ConsumerExports is a common Exports type for Flow components which are
// otelcol processors or otelcol exporters.
type ConsumerExports struct {
	Input Consumer `river:"input,attr"`
}
