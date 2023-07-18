package build

import (
	"fmt"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/converter/internal/prometheusconvert"
)

func (s *ScrapeConfigBuilder) AppendGCESDs() {
	if len(s.cfg.ServiceDiscoveryConfig.GCESDConfigs) == 0 {
		return
	}
	for i, sd := range s.cfg.ServiceDiscoveryConfig.GCESDConfigs {
		s.diags.AddAll(prometheusconvert.ValidateDiscoveryGCE(sd))
		args := prometheusconvert.ToDiscoveryGCE(sd)
		compLabel := fmt.Sprintf("%s_%d", s.cfg.JobName, i)
		s.f.Body().AppendBlock(common.NewBlockWithOverride(
			[]string{"discovery", "gce"},
			compLabel,
			args,
		))
		s.allTargetsExps = append(s.allTargetsExps, "discovery.gce."+compLabel+".targets")
	}
}
