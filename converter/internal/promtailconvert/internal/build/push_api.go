package build

import (
	"github.com/grafana/agent/component/common/loki"
	fnet "github.com/grafana/agent/component/common/net"
	"github.com/grafana/agent/component/common/relabel"
	"github.com/grafana/agent/component/loki/source/api"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/loki/clients/pkg/promtail/scrapeconfig"
)

func (s *ScrapeConfigBuilder) AppendPushAPI() {
	if s.cfg.PushConfig == nil {
		return
	}
	s.diags.AddAll(common.ValidateWeaveWorksServerCfg(s.cfg.PushConfig.Server))
	args := toLokiApiArguments(s.cfg.PushConfig, s.getOrNewProcessStageReceivers())
	override := func(val interface{}) interface{} {
		switch val.(type) {
		case relabel.Rules:
			return common.CustomTokenizer{Expr: s.getOrNewDiscoveryRelabelRules()}
		default:
			return val
		}
	}
	compLabel := common.LabelForParts(s.globalCtx.LabelPrefix, s.cfg.JobName)
	s.f.Body().AppendBlock(common.NewBlockWithOverrideFn(
		[]string{"loki", "source", "api"},
		compLabel,
		args,
		override,
	))
}

func toLokiApiArguments(config *scrapeconfig.PushTargetConfig, forwardTo []loki.LogsReceiver) api.Arguments {
	return api.Arguments{
		ForwardTo:            forwardTo,
		RelabelRules:         make(relabel.Rules, 0),
		Labels:               convertPromLabels(config.Labels),
		UseIncomingTimestamp: config.KeepTimestamp,
		Server: &fnet.ServerConfig{
			HTTP: &fnet.HTTPConfig{
				ListenAddress:      config.Server.HTTPListenAddress,
				ListenPort:         config.Server.HTTPListenPort,
				ConnLimit:          config.Server.HTTPConnLimit,
				ServerReadTimeout:  config.Server.HTTPServerReadTimeout,
				ServerWriteTimeout: config.Server.HTTPServerWriteTimeout,
				ServerIdleTimeout:  config.Server.HTTPServerIdleTimeout,
			},
			GRPC: &fnet.GRPCConfig{
				ListenAddress:              config.Server.GRPCListenAddress,
				ListenPort:                 config.Server.GRPCListenPort,
				ConnLimit:                  config.Server.GRPCConnLimit,
				MaxConnectionAge:           config.Server.GRPCServerMaxConnectionAge,
				MaxConnectionAgeGrace:      config.Server.GRPCServerMaxConnectionAgeGrace,
				MaxConnectionIdle:          config.Server.GRPCServerMaxConnectionIdle,
				ServerMaxRecvMsg:           config.Server.GPRCServerMaxRecvMsgSize,
				ServerMaxSendMsg:           config.Server.GRPCServerMaxSendMsgSize,
				ServerMaxConcurrentStreams: config.Server.GPRCServerMaxConcurrentStreams,
			},
			GracefulShutdownTimeout: config.Server.ServerGracefulShutdownTimeout,
		},
	}
}
