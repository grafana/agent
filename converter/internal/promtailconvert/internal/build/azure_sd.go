package build

import (
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/converter/internal/prometheusconvert"
)

func (s *ScrapeConfigBuilder) AppendAzureSDs() {
	if len(s.cfg.ServiceDiscoveryConfig.AzureSDConfigs) == 0 {
		return
	}
	for i, sd := range s.cfg.ServiceDiscoveryConfig.AzureSDConfigs {
		s.diags.AddAll(prometheusconvert.ValidateDiscoveryAzure(sd))
		args := prometheusconvert.ToDiscoveryAzure(sd)
		compLabel := common.LabelWithIndex(i, s.globalCtx.LabelPrefix, s.cfg.JobName)
		s.f.Body().AppendBlock(common.NewBlockWithOverride(
			[]string{"discovery", "azure"},
			compLabel,
			args,
		))
		s.allTargetsExps = append(s.allTargetsExps, "discovery.azure."+compLabel+".targets")
	}
}
