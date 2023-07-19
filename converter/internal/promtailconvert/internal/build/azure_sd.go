package build

import (
	"fmt"

	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/converter/internal/prometheusconvert"
)

func (s *ScrapeConfigBuilder) AppendAzureSDs() {
	if len(s.cfg.ServiceDiscoveryConfig.AzureSDConfigs) == 0 {
		return
	}
	for i, sd := range s.cfg.ServiceDiscoveryConfig.AzureSDConfigs {
		s.diags.AddAll(prometheusconvert.ValidateDiscoveryAzure(sd))
		compName := fmt.Sprintf("%s_%d", s.cfg.JobName, i)

		args := prometheusconvert.ToDiscoveryAzure(sd)
		s.f.Body().AppendBlock(common.NewBlockWithOverride(
			[]string{"discovery", "azure"},
			compName,
			args,
		))
		s.allTargetsExps = append(s.allTargetsExps, "discovery.azure."+compName+".targets")
	}
}
