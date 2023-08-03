package build

import (
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/converter/internal/prometheusconvert"
)

func (s *ScrapeConfigBuilder) AppendFileSDs() {
	if len(s.cfg.ServiceDiscoveryConfig.FileSDConfigs) == 0 {
		return
	}
	for i, sd := range s.cfg.ServiceDiscoveryConfig.FileSDConfigs {
		args := prometheusconvert.ToDiscoveryFile(sd)
		compLabel := common.LabelWithIndex(i, s.globalCtx.LabelPrefix, s.cfg.JobName)
		s.f.Body().AppendBlock(common.NewBlockWithOverride([]string{"discovery", "file"}, compLabel, args))
		s.allTargetsExps = append(s.allTargetsExps, "discovery.file."+compLabel+".targets")
	}
}
