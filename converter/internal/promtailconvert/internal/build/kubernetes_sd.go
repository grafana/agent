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
		compLabel := common.GetLabelWithPrefix(s.globalCtx.LabelPrefix, s.cfg.JobName, i)
		s.f.Body().AppendBlock(common.NewBlockWithOverride(
			[]string{"discovery", "kubernetes"},
			compLabel,
			args,
		))
		s.allTargetsExps = append(s.allTargetsExps, "discovery.kubernetes."+compLabel+".targets")
	}
}
