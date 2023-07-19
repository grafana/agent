package build

import (
	"fmt"

	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/converter/internal/prometheusconvert"
)

func (s *ScrapeConfigBuilder) AppendDigitalOceanSDs() {
	if len(s.cfg.ServiceDiscoveryConfig.DigitalOceanSDConfigs) == 0 {
		return
	}
	for i, sd := range s.cfg.ServiceDiscoveryConfig.DigitalOceanSDConfigs {
		s.diags.AddAll(prometheusconvert.ValidateDiscoveryDigitalOcean(sd))
		compName := fmt.Sprintf("%s_%d", s.cfg.JobName, i)

		args := prometheusconvert.ToDiscoveryDigitalOcean(sd)
		s.f.Body().AppendBlock(common.NewBlockWithOverride(
			[]string{"discovery", "digitalocean"},
			compName,
			args,
		))
		s.allTargetsExps = append(s.allTargetsExps, "discovery.digitalocean."+compName+".targets")
	}
}
