package otelcolconvert

import (
	"fmt"
	"time"

	"github.com/alecthomas/units"
	"github.com/grafana/agent/component/otelcol"
	"github.com/grafana/agent/component/otelcol/exporter/otlphttp"
	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/converter/internal/common"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/exporter/otlphttpexporter"
)

func init() {
	converters = append(converters, otlpHTTPExporterConverter{})
}

type otlpHTTPExporterConverter struct{}

func (otlpHTTPExporterConverter) Factory() component.Factory {
	return otlphttpexporter.NewFactory()
}

func (otlpHTTPExporterConverter) InputComponentName() string {
	return "otelcol.exporter.otlphttp"
}

func (otlpHTTPExporterConverter) ConvertAndAppend(state *state, id component.InstanceID, cfg component.Config) diag.Diagnostics {
	var diags diag.Diagnostics

	label := state.FlowComponentLabel()

	args := toOtelcolExporterOTLPHTTP(cfg.(*otlphttpexporter.Config))
	block := common.NewBlockWithOverride([]string{"otelcol", "exporter", "otlphttp"}, label, args)

	diags.Add(
		diag.SeverityLevelInfo,
		fmt.Sprintf("Converted %s into %s", stringifyInstanceID(id), stringifyBlock(block)),
	)

	state.Body().AppendBlock(block)
	return diags
}

func toOtelcolExporterOTLPHTTP(cfg *otlphttpexporter.Config) *otlphttp.Arguments {
	return &otlphttp.Arguments{
		Client:       otlphttp.HTTPClientArguments(toHTTPClientArguments(cfg.HTTPClientSettings)),
		Queue:        toQueueArguments(cfg.QueueSettings),
		Retry:        toRetryArguments(cfg.RetrySettings),
		DebugMetrics: common.DefaultValue[otlphttp.Arguments]().DebugMetrics,
	}
}

func toHTTPClientArguments(cfg confighttp.HTTPClientSettings) otelcol.HTTPClientArguments {
	var mic *int
	var ict *time.Duration
	defaults := confighttp.NewDefaultHTTPClientSettings()
	if mic = cfg.MaxIdleConns; mic == nil {
		mic = defaults.MaxIdleConns
	}
	if ict = cfg.IdleConnTimeout; ict == nil {
		ict = defaults.IdleConnTimeout
	}
	return otelcol.HTTPClientArguments{
		Endpoint:        cfg.Endpoint,
		Compression:     otelcol.CompressionType(cfg.Compression),
		TLS:             toTLSClientArguments(cfg.TLSSetting),
		ReadBufferSize:  units.Base2Bytes(cfg.ReadBufferSize),
		WriteBufferSize: units.Base2Bytes(cfg.WriteBufferSize),

		Timeout:             cfg.Timeout,
		Headers:             toHeadersMap(cfg.Headers),
		MaxIdleConns:        mic,
		MaxIdleConnsPerHost: cfg.MaxIdleConnsPerHost,
		MaxConnsPerHost:     cfg.MaxConnsPerHost,
		IdleConnTimeout:     ict,
		DisableKeepAlives:   cfg.DisableKeepAlives,

		// TODO(@tpaschalis): auth extension
	}
}
