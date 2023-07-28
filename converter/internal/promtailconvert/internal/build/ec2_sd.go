package build

import (
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/converter/internal/prometheusconvert"
)

func (s *ScrapeConfigBuilder) AppendEC2SDs() {
	if len(s.cfg.ServiceDiscoveryConfig.EC2SDConfigs) == 0 {
		return
	}
	for i, sd := range s.cfg.ServiceDiscoveryConfig.EC2SDConfigs {
		s.diags.AddAll(prometheusconvert.ValidateDiscoveryEC2(sd))
		args := prometheusconvert.ToDiscoveryEC2(sd)
		compLabel := common.LabelWithIndex(i, s.globalCtx.LabelPrefix, s.cfg.JobName)
		s.f.Body().AppendBlock(common.NewBlockWithOverride(
			[]string{"discovery", "ec2"},
			compLabel,
			args,
		))
		s.allTargetsExps = append(s.allTargetsExps, "discovery.ec2."+compLabel+".targets")
	}
}
