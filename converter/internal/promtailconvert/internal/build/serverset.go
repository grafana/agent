package build

import (
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/converter/internal/prometheusconvert"
)

func (s *ScrapeConfigBuilder) AppendServersetSDs() {
	if len(s.cfg.ServiceDiscoveryConfig.ServersetSDConfigs) == 0 {
		return
	}
	for i, sd := range s.cfg.ServiceDiscoveryConfig.ServersetSDConfigs {
		s.diags.AddAll(prometheusconvert.ValidateDiscoveryServerset(sd))
		args := prometheusconvert.ToDiscoveryServerset(sd)
		compLabel := common.LabelWithIndex(i, s.globalCtx.LabelPrefix, s.cfg.JobName)
		s.f.Body().AppendBlock(common.NewBlockWithOverride(
			[]string{"discovery", "serverset"},
			compLabel,
			args,
		))
		s.allTargetsExps = append(s.allTargetsExps, "discovery.serverset."+compLabel+".targets")
	}
}
