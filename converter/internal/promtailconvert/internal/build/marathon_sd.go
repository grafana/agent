package build

import (
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/converter/internal/prometheusconvert"
)

func (s *ScrapeConfigBuilder) AppendMarathonSD() {
	if len(s.cfg.ServiceDiscoveryConfig.MarathonSDConfigs) == 0 {
		return
	}
	for i, sd := range s.cfg.ServiceDiscoveryConfig.MarathonSDConfigs {
		args := prometheusconvert.ToDiscoveryMarathon(sd)
		compLabel := common.LabelWithIndex(i, s.globalCtx.LabelPrefix, s.cfg.JobName)
		s.f.Body().AppendBlock(common.NewBlockWithOverride([]string{"discovery", "marathon"}, compLabel, args))
		s.allTargetsExps = append(s.allTargetsExps, "discovery.marathon."+compLabel+".targets")
	}
}
