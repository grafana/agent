package build

import (
	"time"

	"github.com/grafana/agent/component/discovery/consulagent"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/converter/internal/prometheusconvert"
	promtail_consulagent "github.com/grafana/loki/clients/pkg/promtail/discovery/consulagent"
	"github.com/grafana/river/rivertypes"
)

func (s *ScrapeConfigBuilder) AppendConsulAgentSDs() {
	if len(s.cfg.ServiceDiscoveryConfig.ConsulAgentSDConfigs) == 0 {
		return
	}

	for i, sd := range s.cfg.ServiceDiscoveryConfig.ConsulAgentSDConfigs {
		args := toDiscoveryAgentConsul(sd)
		compLabel := common.LabelWithIndex(i, s.globalCtx.LabelPrefix, s.cfg.JobName)
		s.f.Body().AppendBlock(common.NewBlockWithOverride(
			[]string{"discovery", "consulagent"},
			compLabel,
			args,
		))
		s.allTargetsExps = append(s.allTargetsExps, "discovery.consulagent."+compLabel+".targets")
	}
}

func toDiscoveryAgentConsul(sdConfig *promtail_consulagent.SDConfig) *consulagent.Arguments {
	if sdConfig == nil {
		return nil
	}
	return &consulagent.Arguments{
		RefreshInterval: time.Duration(sdConfig.RefreshInterval),
		Server:          sdConfig.Server,
		Token:           rivertypes.Secret(sdConfig.Token),
		Datacenter:      sdConfig.Datacenter,
		TagSeparator:    sdConfig.TagSeparator,
		Scheme:          sdConfig.Scheme,
		Username:        sdConfig.Username,
		Password:        rivertypes.Secret(sdConfig.Password),
		AllowStale:      sdConfig.AllowStale,
		Services:        sdConfig.Services,
		ServiceTags:     sdConfig.ServiceTags,
		NodeMeta:        sdConfig.NodeMeta,
		TLSConfig:       *prometheusconvert.ToTLSConfig(&sdConfig.TLSConfig),
	}
}
