package build

import (
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/converter/internal/prometheusconvert"
)

func (s *ScrapeConfigBuilder) AppendKubernetesSDs() {
	if len(s.cfg.ServiceDiscoveryConfig.KubernetesSDConfigs) == 0 {
		return
	}

	for i, sd := range s.cfg.ServiceDiscoveryConfig.KubernetesSDConfigs {
		s.diags.AddAll(prometheusconvert.ValidateHttpClientConfig(&sd.HTTPClientConfig))
		args := prometheusconvert.ToDiscoveryKubernetes(sd)
		compLabel := common.GetLabelWithIndex(i, s.globalCtx.LabelPrefix, s.cfg.JobName)
		s.f.Body().AppendBlock(common.NewBlockWithOverride(
			[]string{"discovery", "kubernetes"},
			compLabel,
			args,
		))
		s.allTargetsExps = append(s.allTargetsExps, "discovery.kubernetes."+compLabel+".targets")
	}
}
