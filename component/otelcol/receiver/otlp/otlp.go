// Package otlp provides an otelcol.receiver.otlp component.
package otlp

import (
	"fmt"
	net_url "net/url"

	"github.com/alecthomas/units"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/otelcol"
	"github.com/grafana/agent/component/otelcol/receiver"
	otel_service "github.com/grafana/agent/service/otel"
	otelcomponent "go.opentelemetry.io/collector/component"
	otelextension "go.opentelemetry.io/collector/extension"
	"go.opentelemetry.io/collector/receiver/otlpreceiver"
)

func init() {
	component.Register(component.Registration{
		Name:          "otelcol.receiver.otlp",
		Args:          Arguments{},
		NeedsServices: []string{otel_service.ServiceName},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			fact := otlpreceiver.NewFactory()
			return receiver.New(opts, fact, args.(Arguments))
		},
	})
}

// Arguments configures the otelcol.receiver.otlp component.
type Arguments struct {
	GRPC *GRPCServerArguments `river:"grpc,block,optional"`
	HTTP *HTTPConfigArguments `river:"http,block,optional"`

	// DebugMetrics configures component internal metrics. Optional.
	DebugMetrics otelcol.DebugMetricsArguments `river:"debug_metrics,block,optional"`

	// Output configures where to send received data. Required.
	Output *otelcol.ConsumerArguments `river:"output,block"`
}

type HTTPConfigArguments struct {
	HTTPServerArguments *otelcol.HTTPServerArguments `river:",squash"`

	// The URL path to receive traces on. If omitted "/v1/traces" will be used.
	TracesURLPath string `river:"traces_url_path,attr,optional"`

	// The URL path to receive metrics on. If omitted "/v1/metrics" will be used.
	MetricsURLPath string `river:"metrics_url_path,attr,optional"`

	// The URL path to receive logs on. If omitted "/v1/logs" will be used.
	LogsURLPath string `river:"logs_url_path,attr,optional"`
}

// Convert converts args into the upstream type.
func (args *HTTPConfigArguments) Convert() *otlpreceiver.HTTPConfig {
	if args == nil {
		return nil
	}

	return &otlpreceiver.HTTPConfig{
		HTTPServerSettings: args.HTTPServerArguments.Convert(),
		TracesURLPath:      args.TracesURLPath,
		MetricsURLPath:     args.MetricsURLPath,
		LogsURLPath:        args.LogsURLPath,
	}
}

var _ receiver.Arguments = Arguments{}

// Convert implements receiver.Arguments.
func (args Arguments) Convert() (otelcomponent.Config, error) {
	return &otlpreceiver.Config{
		Protocols: otlpreceiver.Protocols{
			GRPC: (*otelcol.GRPCServerArguments)(args.GRPC).Convert(),
			HTTP: args.HTTP.Convert(),
		},
	}, nil
}

// Extensions implements receiver.Arguments.
func (args Arguments) Extensions() map[otelcomponent.ID]otelextension.Extension {
	return nil
}

// Exporters implements receiver.Arguments.
func (args Arguments) Exporters() map[otelcomponent.DataType]map[otelcomponent.ID]otelcomponent.Component {
	return nil
}

// NextConsumers implements receiver.Arguments.
func (args Arguments) NextConsumers() *otelcol.ConsumerArguments {
	return args.Output
}

type (
	// GRPCServerArguments is used to configure otelcol.receiver.otlp with
	// component-specific defaults.
	GRPCServerArguments otelcol.GRPCServerArguments
)

// Default server settings.
var (
	DefaultGRPCServerArguments = GRPCServerArguments{
		Endpoint:  "0.0.0.0:4317",
		Transport: "tcp",

		ReadBufferSize: 512 * units.Kibibyte,
		// We almost write 0 bytes, so no need to tune WriteBufferSize.
	}

	DefaultHTTPConfigArguments = HTTPConfigArguments{
		HTTPServerArguments: &otelcol.HTTPServerArguments{
			Endpoint: "0.0.0.0:4318",
		},
		MetricsURLPath: "/v1/metrics",
		LogsURLPath:    "/v1/logs",
		TracesURLPath:  "/v1/traces",
	}
)

// Validate implements river.Validator.
func (args *Arguments) Validate() error {
	if args.HTTP != nil {
		if err := validateURL(args.HTTP.TracesURLPath, "traces_url_path"); err != nil {
			return err
		}
		if err := validateURL(args.HTTP.LogsURLPath, "logs_url_path"); err != nil {
			return err
		}
		if err := validateURL(args.HTTP.MetricsURLPath, "metrics_url_path"); err != nil {
			return err
		}
	}
	return nil
}

func validateURL(url string, urlName string) error {
	if url == "" {
		return fmt.Errorf("%s cannot be empty", urlName)
	}
	if _, err := net_url.Parse(url); err != nil {
		return fmt.Errorf("invalid %s: %w", urlName, err)
	}
	return nil
}

// SetToDefault implements river.Defaulter.
func (args *GRPCServerArguments) SetToDefault() {
	*args = DefaultGRPCServerArguments
}

// SetToDefault implements river.Defaulter.
func (args *HTTPConfigArguments) SetToDefault() {
	*args = DefaultHTTPConfigArguments
}

// DebugMetricsConfig implements receiver.Arguments.
func (args Arguments) DebugMetricsConfig() otelcol.DebugMetricsArguments {
	return args.DebugMetrics
}
