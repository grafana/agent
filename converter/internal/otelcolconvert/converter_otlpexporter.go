package otelcolconvert

import (
	"fmt"

	"github.com/alecthomas/units"
	"github.com/grafana/agent/component/otelcol"
	"github.com/grafana/agent/component/otelcol/exporter/otlp"
	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/converter/internal/common"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configgrpc"
	"go.opentelemetry.io/collector/config/configopaque"
	"go.opentelemetry.io/collector/config/configtls"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
	"go.opentelemetry.io/collector/exporter/otlpexporter"
)

func init() {
	converters = append(converters, otlpExporterConverter{})
}

type otlpExporterConverter struct{}

func (otlpExporterConverter) Factory() component.Factory {
	return otlpexporter.NewFactory()
}

func (otlpExporterConverter) InputComponentName() string { return "otelcol.exporter.otlp" }

func (otlpExporterConverter) ConvertAndAppend(state *state, id component.InstanceID, cfg component.Config) diag.Diagnostics {
	var diags diag.Diagnostics

	label := state.FlowComponentLabel()

	args := toOtelcolExporterOTLP(cfg.(*otlpexporter.Config))
	block := common.NewBlockWithOverride([]string{"otelcol", "exporter", "otlp"}, label, args)

	diags.Add(
		diag.SeverityLevelInfo,
		fmt.Sprintf("Converted %s into %s", stringifyInstanceID(id), stringifyBlock(block)),
	)

	state.Body().AppendBlock(block)
	return diags
}

func toOtelcolExporterOTLP(cfg *otlpexporter.Config) *otlp.Arguments {
	return &otlp.Arguments{
		Timeout: cfg.Timeout,

		Queue: toQueueArguments(cfg.QueueSettings),
		Retry: toRetryArguments(cfg.RetrySettings),

		DebugMetrics: common.DefaultValue[otlp.Arguments]().DebugMetrics,

		Client: otlp.GRPCClientArguments(toGRPCClientArguments(cfg.GRPCClientSettings)),
	}
}

func toQueueArguments(cfg exporterhelper.QueueSettings) otelcol.QueueArguments {
	return otelcol.QueueArguments{
		Enabled:      cfg.Enabled,
		NumConsumers: cfg.NumConsumers,
		QueueSize:    cfg.QueueSize,
	}
}

func toRetryArguments(cfg exporterhelper.RetrySettings) otelcol.RetryArguments {
	return otelcol.RetryArguments{
		Enabled:             cfg.Enabled,
		InitialInterval:     cfg.InitialInterval,
		RandomizationFactor: cfg.RandomizationFactor,
		Multiplier:          cfg.Multiplier,
		MaxInterval:         cfg.MaxInterval,
		MaxElapsedTime:      cfg.MaxElapsedTime,
	}
}

func toGRPCClientArguments(cfg configgrpc.GRPCClientSettings) otelcol.GRPCClientArguments {
	return otelcol.GRPCClientArguments{
		Endpoint: cfg.Endpoint,

		Compression: otelcol.CompressionType(cfg.Compression),

		TLS:       toTLSClientArguments(cfg.TLSSetting),
		Keepalive: toKeepaliveClientArguments(cfg.Keepalive),

		ReadBufferSize:  units.Base2Bytes(cfg.ReadBufferSize),
		WriteBufferSize: units.Base2Bytes(cfg.WriteBufferSize),
		WaitForReady:    cfg.WaitForReady,
		Headers:         toHeadersMap(cfg.Headers),
		BalancerName:    cfg.BalancerName,
		Authority:       cfg.Authority,

		// TODO(rfratto): auth extension
	}
}

func toTLSClientArguments(cfg configtls.TLSClientSetting) otelcol.TLSClientArguments {
	return otelcol.TLSClientArguments{
		TLSSetting: toTLSSetting(cfg.TLSSetting),

		Insecure:           cfg.Insecure,
		InsecureSkipVerify: cfg.InsecureSkipVerify,
		ServerName:         cfg.ServerName,
	}
}

func toKeepaliveClientArguments(cfg *configgrpc.KeepaliveClientConfig) *otelcol.KeepaliveClientArguments {
	if cfg == nil {
		return nil
	}

	return &otelcol.KeepaliveClientArguments{
		PingWait:            cfg.Time,
		PingResponseTimeout: cfg.Timeout,
		PingWithoutStream:   cfg.PermitWithoutStream,
	}
}

func toHeadersMap(cfg map[string]configopaque.String) map[string]string {
	res := make(map[string]string, len(cfg))
	for k, v := range cfg {
		res[k] = string(v)
	}
	return res
}
