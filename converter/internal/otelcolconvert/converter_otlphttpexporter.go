package otelcolconvert

import (
	"fmt"

	"github.com/alecthomas/units"
	"github.com/grafana/agent/component/otelcol"
	"github.com/grafana/agent/component/otelcol/exporter/otlphttp"
	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/converter/internal/common"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configcompression"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/exporter/otlphttpexporter"
)

func init() {
	converters = append(converters, otlphttpExporterConverter{})
}

type otlphttpExporterConverter struct{}

func (otlphttpExporterConverter) Factory() component.Factory { return otlphttpexporter.NewFactory() }

func (otlphttpExporterConverter) InputComponentName() string { return "otelcol.exporter.otlphttp" }

func (otlphttpExporterConverter) ConvertAndAppend(state *state, id component.InstanceID, cfg component.Config) diag.Diagnostics {
	var diags diag.Diagnostics

	label := state.FlowComponentLabel()

	// NOTE(rfratto): the error from toOtelcolExporterOTLPHTTP is non-fatal, so
	// we can continue best-effort conversion even if it fails.
	args, err := toOtelcolExporterOTLPHTTP(cfg.(*otlphttpexporter.Config))
	if err != nil {
		diags.Add(
			diag.SeverityLevelError,
			fmt.Sprintf("failed to fully convert %s: %s", stringifyInstanceID(id), err),
		)
	}
	block := common.NewBlockWithOverride([]string{"otelcol", "exporter", "otlphttp"}, label, args)

	diags.Add(
		diag.SeverityLevelInfo,
		fmt.Sprintf("Converted %s into %s", stringifyInstanceID(id), stringifyBlock(block)),
	)

	state.Body().AppendBlock(block)
	return diags
}

func toOtelcolExporterOTLPHTTP(cfg *otlphttpexporter.Config) (*otlphttp.Arguments, error) {
	// NOTE(rfratto): the error from toHTTPClientArguments is non-fatal, so we
	// can continue best-effort conversion even if it fails.
	clientArgs, err := toHTTPClientArguments(cfg.HTTPClientSettings)

	return &otlphttp.Arguments{
		Client: otlphttp.HTTPClientArguments(clientArgs),
		Queue:  toQueueArguments(cfg.QueueSettings),
		Retry:  toRetryArguments(cfg.RetrySettings),

		DebugMetrics: common.DefaultValue[otlphttp.Arguments]().DebugMetrics,

		TracesEndpoint:  cfg.TracesEndpoint,
		MetricsEndpoint: cfg.MetricsEndpoint,
		LogsEndpoint:    cfg.LogsEndpoint,
	}, err
}

func toHTTPClientArguments(cfg confighttp.HTTPClientSettings) (otelcol.HTTPClientArguments, error) {
	// NOTE(rfratto): the error from toCompressionType is non-fatal, so we can
	// continue best-effort conversion even if it fails.
	compression, err := toCompressionType(cfg.Compression)

	return otelcol.HTTPClientArguments{
		Endpoint: cfg.Endpoint,

		Compression: compression,

		TLS: toTLSClientArguments(cfg.TLSSetting),

		ReadBufferSize:      units.Base2Bytes(cfg.ReadBufferSize),
		WriteBufferSize:     units.Base2Bytes(cfg.WriteBufferSize),
		Timeout:             cfg.Timeout,
		Headers:             toHeadersMap(cfg.Headers),
		MaxIdleConns:        cfg.MaxIdleConns,
		MaxIdleConnsPerHost: cfg.MaxIdleConnsPerHost,
		MaxConnsPerHost:     cfg.MaxConnsPerHost,
		IdleConnTimeout:     cfg.IdleConnTimeout,
		DisableKeepAlives:   cfg.DisableKeepAlives,

		// TODO(rfratto): auth support
	}, err
}

func toCompressionType(cfg configcompression.CompressionType) (otelcol.CompressionType, error) {
	switch cfg {
	case configcompression.Gzip:
		return otelcol.CompressionTypeGzip, nil
	case configcompression.Zlib:
		return otelcol.CompressionTypeZlib, nil
	case configcompression.Deflate:
		return otelcol.CompressionTypeDeflate, nil
	case configcompression.Snappy:
		return otelcol.CompressionTypeSnappy, nil
	case configcompression.Zstd:
		return otelcol.CompressionTypeZstd, nil
	case configcompression.CompressionType("none"):
		return otelcol.CompressionTypeNone, nil
	case configcompression.CompressionType(""):
		return otelcol.CompressionTypeEmpty, nil
	default:
		return otelcol.CompressionTypeNone, fmt.Errorf("unrecognized compression type %q; compression is disabled", cfg)
	}
}
