package build

import (
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/converter/internal/prometheusconvert"
)

func (s *ScrapeConfigBuilder) AppendConsulSDs() {
	if len(s.cfg.ServiceDiscoveryConfig.ConsulSDConfigs) == 0 {
		return
	}

	for i, sd := range s.cfg.ServiceDiscoveryConfig.ConsulSDConfigs {
		args := prometheusconvert.ToDiscoveryConsul(sd)
		compLabel := common.LabelWithIndex(i, s.globalCtx.LabelPrefix, s.cfg.JobName)
		s.f.Body().AppendBlock(common.NewBlockWithOverride(
			[]string{"discovery", "consul"},
			compLabel,
			args,
		))
		s.allTargetsExps = append(s.allTargetsExps, "discovery.consul."+compLabel+".targets")
	}
}
