package build

import (
	"github.com/grafana/agent/component/common/relabel"
	"github.com/grafana/agent/component/loki/source/gcplog"
	"github.com/grafana/agent/component/loki/source/gcplog/gcptypes"
	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/converter/internal/common"
)

func (s *ScrapeConfigBuilder) AppendGCPLog() {
	if s.cfg.GcplogConfig == nil {
		return
	}

	var (
		pushConfig *gcptypes.PushConfig = nil
		pullConfig *gcptypes.PullConfig = nil
	)

	cfg := s.cfg.GcplogConfig
	switch cfg.SubscriptionType {
	case "", "pull":
		pullConfig = &gcptypes.PullConfig{
			ProjectID:            cfg.ProjectID,
			Subscription:         cfg.Subscription,
			Labels:               convertPromLabels(cfg.Labels),
			UseIncomingTimestamp: cfg.UseIncomingTimestamp,
			UseFullLine:          cfg.UseFullLine,
		}
	case "push":
		s.diags.AddAll(common.ValidateWeaveWorksServerCfg(cfg.Server))
		flowServer := common.WeaveWorksServerToFlowServer(cfg.Server)
		pushConfig = &gcptypes.PushConfig{
			Server:               flowServer,
			PushTimeout:          cfg.PushTimeout,
			Labels:               convertPromLabels(cfg.Labels),
			UseIncomingTimestamp: cfg.UseIncomingTimestamp,
			UseFullLine:          cfg.UseFullLine,
		}
	default:
		s.diags.Add(diag.SeverityLevelError, "gcplog.subscription_type must be one of 'pull' or 'push'")
	}

	args := gcplog.Arguments{
		PullTarget:   pullConfig,
		PushTarget:   pushConfig,
		ForwardTo:    s.getOrNewProcessStageReceivers(),
		RelabelRules: make(relabel.Rules, 0),
	}

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
		[]string{"loki", "source", "gcplog"},
		compLabel,
		args,
		override,
	))
}
