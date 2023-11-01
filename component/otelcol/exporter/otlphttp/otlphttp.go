// Package otlphttp provides an otelcol.exporter.otlphttp component.
package otlphttp

import (
	"errors"
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/otelcol"
	"github.com/grafana/agent/component/otelcol/exporter"
	otel_service "github.com/grafana/agent/service/otel"
	otelcomponent "go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/exporter/otlphttpexporter"
	otelextension "go.opentelemetry.io/collector/extension"
)

func init() {
	component.Register(component.Registration{
		Name:          "otelcol.exporter.otlphttp",
		Args:          Arguments{},
		Exports:       otelcol.ConsumerExports{},
		NeedsServices: []string{otel_service.ServiceName},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			fact := otlphttpexporter.NewFactory()
			return exporter.New(opts, fact, args.(Arguments))
		},
	})
}

// Arguments configures the otelcol.exporter.otlphttp component.
type Arguments struct {
	Client HTTPClientArguments    `river:"client,block"`
	Queue  otelcol.QueueArguments `river:"sending_queue,block,optional"`
	Retry  otelcol.RetryArguments `river:"retry_on_failure,block,optional"`

	// DebugMetrics configures component internal metrics. Optional.
	DebugMetrics otelcol.DebugMetricsArguments `river:"debug_metrics,block,optional"`

	// The URLs to send metrics/logs/traces to. If omitted the exporter will
	// use Client.Endpoint by appending "/v1/metrics", "/v1/logs" or
	// "/v1/traces", respectively. If set, these settings override
	// Client.Endpoint for the corresponding signal.
	TracesEndpoint  string `river:"traces_endpoint,attr,optional"`
	MetricsEndpoint string `river:"metrics_endpoint,attr,optional"`
	LogsEndpoint    string `river:"logs_endpoint,attr,optional"`
}

var _ exporter.Arguments = Arguments{}

// DefaultArguments holds default values for Arguments.
var DefaultArguments = Arguments{
	Queue:  otelcol.DefaultQueueArguments,
	Retry:  otelcol.DefaultRetryArguments,
	Client: DefaultHTTPClientArguments,
}

// SetToDefault implements river.Defaulter.
func (args *Arguments) SetToDefault() {
	*args = DefaultArguments
}

// Convert implements exporter.Arguments.
func (args Arguments) Convert() (otelcomponent.Config, error) {
	return &otlphttpexporter.Config{
		HTTPClientSettings: *(*otelcol.HTTPClientArguments)(&args.Client).Convert(),
		QueueSettings:      *args.Queue.Convert(),
		RetrySettings:      *args.Retry.Convert(),
		TracesEndpoint:     args.TracesEndpoint,
		MetricsEndpoint:    args.MetricsEndpoint,
		LogsEndpoint:       args.LogsEndpoint,
	}, nil
}

// Extensions implements exporter.Arguments.
func (args Arguments) Extensions() map[otelcomponent.ID]otelextension.Extension {
	return (*otelcol.HTTPClientArguments)(&args.Client).Extensions()
}

// Exporters implements exporter.Arguments.
func (args Arguments) Exporters() map[otelcomponent.DataType]map[otelcomponent.ID]otelcomponent.Component {
	return nil
}

// DebugMetricsConfig implements receiver.Arguments.
func (args Arguments) DebugMetricsConfig() otelcol.DebugMetricsArguments {
	return args.DebugMetrics
}

// Validate implements river.Validator.
func (args *Arguments) Validate() error {
	if args.Client.Endpoint == "" && args.TracesEndpoint == "" && args.MetricsEndpoint == "" && args.LogsEndpoint == "" {
		return errors.New("at least one endpoint must be specified")
	}
	return nil
}

// HTTPClientArguments is used to configure otelcol.exporter.otlphttp with
// component-specific defaults.
type HTTPClientArguments otelcol.HTTPClientArguments

// Default server settings.
var (
	DefaultMaxIddleConns       = 100
	DefaultIdleConnTimeout     = 90 * time.Second
	DefaultHTTPClientArguments = HTTPClientArguments{
		MaxIdleConns:    &DefaultMaxIddleConns,
		IdleConnTimeout: &DefaultIdleConnTimeout,

		Timeout:         30 * time.Second,
		Headers:         map[string]string{},
		Compression:     otelcol.CompressionTypeGzip,
		ReadBufferSize:  0,
		WriteBufferSize: 512 * 1024,
	}
)

// SetToDefault implements river.Defaulter.
func (args *HTTPClientArguments) SetToDefault() {
	*args = DefaultHTTPClientArguments
}
