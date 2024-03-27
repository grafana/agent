package otelcolconvert

import (
	"fmt"

	"github.com/grafana/agent/internal/component/common/loki"
	"github.com/grafana/agent/internal/component/loki/source/api"
	"github.com/grafana/agent/internal/component/otelcol"
	otel_loki "github.com/grafana/agent/internal/component/otelcol/receiver/loki"
	"github.com/grafana/agent/internal/converter/diag"
	"github.com/grafana/agent/internal/converter/internal/common"
	"github.com/grafana/agent/internal/converter/internal/promtailconvert/build"
	"github.com/grafana/dskit/server"
	"github.com/grafana/loki/clients/pkg/promtail/scrapeconfig"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/lokireceiver"
	"go.opentelemetry.io/collector/component"
)

func init() {
	converters = append(converters, lokiReceiverConverter{})
}

type lokiReceiverConverter struct{}

func (lokiReceiverConverter) Factory() component.Factory { return lokireceiver.NewFactory() }

func (lokiReceiverConverter) InputComponentName() string { return "" }

func (lokiReceiverConverter) ConvertAndAppend(state *State, id component.InstanceID, cfg component.Config) diag.Diagnostics {
	var diags diag.Diagnostics

	label := state.FlowComponentLabel()

	otelArgs := toOtelcolReceiverLoki(state, id)
	otelBlock := common.NewBlockWithOverride([]string{"otelcol", "receiver", "loki"}, label, otelArgs)

	logsReceivers := []loki.LogsReceiver{common.ConvertLogsReceiver{
		Expr: StringifyBlock(otelBlock) + ".receiver",
	}}
	apiArgs := toLokiSourceApi(logsReceivers, cfg.(*lokireceiver.Config))
	apiBlock := common.NewBlockWithOverride([]string{"loki", "source", "api"}, label, apiArgs)

	diags.Add(
		diag.SeverityLevelInfo,
		fmt.Sprintf("Converted %s into %s", StringifyInstanceID(id), StringifyBlock(otelBlock)),
	)

	diags.Add(
		diag.SeverityLevelInfo,
		fmt.Sprintf("Converted %s into %s", StringifyInstanceID(id), StringifyBlock(apiBlock)),
	)

	// Do this at the end in reverse order so the pipeline looks a little better
	state.Body().AppendBlock(apiBlock)
	state.Body().AppendBlock(otelBlock)
	return diags
}

func toOtelcolReceiverLoki(state *State, id component.InstanceID) *otel_loki.Arguments {
	var (
		nextLogs = state.Next(id, component.DataTypeLogs)
	)

	return &otel_loki.Arguments{
		Output: &otelcol.ConsumerArguments{
			Logs: ToTokenizedConsumers(nextLogs),
		},
	}
}

func toLokiSourceApi(logsReceivers []loki.LogsReceiver, cfg *lokireceiver.Config) *api.Arguments {
	ptc := &scrapeconfig.PushTargetConfig{
		Server:        toServer(&cfg.Protocols),
		Labels:        nil,
		KeepTimestamp: cfg.KeepTimestamp,
	}

	args := build.ToLokiApiArguments(ptc, logsReceivers)
	return &args
}

func toServer(cfg *lokireceiver.Protocols) server.Config {
	return server.Config{
		HTTPListenAddress:               "",
		HTTPListenPort:                  0,
		HTTPConnLimit:                   0,
		HTTPServerReadTimeout:           0,
		HTTPServerWriteTimeout:          0,
		HTTPServerIdleTimeout:           0,
		GRPCListenAddress:               "",
		GRPCListenPort:                  0,
		GRPCConnLimit:                   0,
		GRPCServerMaxConnectionAge:      0,
		GRPCServerMaxConnectionAgeGrace: 0,
		GRPCServerMaxRecvMsgSize:        0,
		GRPCServerMaxSendMsgSize:        0,
		GRPCServerMaxConcurrentStreams:  uint(cfg.GRPC.MaxConcurrentStreams),
		GRPCServerMaxConnectionIdle:     0,
		ServerGracefulShutdownTimeout:   cfg.GRPC.NetAddr.DialerConfig.Timeout,

		// PROMTAIL CONVERTER IGNORES ALL OF THESE
		//
		// HTTPListenNetwork:                        "",
		// HTTPServerReadHeaderTimeout:              0,
		// GRPCListenNetwork:                        "",
		// CipherSuites:                             "",
		// MinVersion:                               "",
		// HTTPTLSConfig:                            server.TLSConfig{},
		// GRPCTLSConfig:                            server.TLSConfig{},
		// RegisterInstrumentation:                  false,
		// ReportGRPCCodesInInstrumentationLabel:    false,
		// ReportHTTP4XXCodesInInstrumentationLabel: false,
		// ExcludeRequestInLog:                      false,
		// DisableRequestSuccessLog:                 false,
		// HTTPLogClosedConnectionsWithoutResponse:  false,
		// GRPCOptions:                              []grpc.ServerOption{},
		// GRPCMiddleware:                           []grpc.UnaryServerInterceptor{},
		// GRPCStreamMiddleware:                     []grpc.StreamServerInterceptor{},
		// HTTPMiddleware:                           []middleware.Interface{},
		// Router:                                   &mux.Router{},
		// DoNotAddDefaultHTTPMiddleware:            false,
		// RouteHTTPToGRPC:                          false,
		// GRPCServerTime:                           0,
		// GRPCServerTimeout:                        cfg.GRPC.NetAddr.DialerConfig.Timeout,
		// GRPCServerMinTimeBetweenPings:            0,
		// GRPCServerPingWithoutStreamAllowed:       false,
		// GRPCServerNumWorkers:                     0,
		// LogFormat:                                "",
		// LogLevel:                                 log.Level{},
		// Log:                                      nil,
		// LogSourceIPs:                             false,
		// LogSourceIPsHeader:                       "",
		// LogSourceIPsRegex:                        "",
		// LogRequestHeaders:                        false,
		// LogRequestAtInfoLevel:                    false,
		// LogRequestExcludeHeadersList:             "",
		// SignalHandler:                            nil,
		// Registerer:                               nil,
		// Gatherer:                                 nil,
		// PathPrefix:                               "",
		// GrpcMethodLimiter:                        nil,
	}
}
