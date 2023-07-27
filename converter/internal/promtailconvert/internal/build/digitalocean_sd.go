package build

import (
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/converter/internal/prometheusconvert"
)

func (s *ScrapeConfigBuilder) AppendDigitalOceanSDs() {
	if len(s.cfg.ServiceDiscoveryConfig.DigitalOceanSDConfigs) == 0 {
		return
	}
	for i, sd := range s.cfg.ServiceDiscoveryConfig.DigitalOceanSDConfigs {
		s.diags.AddAll(prometheusconvert.ValidateDiscoveryDigitalOcean(sd))
		args := prometheusconvert.ToDiscoveryDigitalOcean(sd)
		compLabel := common.LabelWithIndex(i, s.globalCtx.LabelPrefix, s.cfg.JobName)
		s.f.Body().AppendBlock(common.NewBlockWithOverride(
			[]string{"discovery", "digitalocean"},
			compLabel,
			args,
		))
		s.allTargetsExps = append(s.allTargetsExps, "discovery.digitalocean."+compLabel+".targets")
	}
}
