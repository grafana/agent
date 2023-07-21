package build

import (
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/converter/internal/prometheusconvert"
)

func (s *ScrapeConfigBuilder) AppendDockerSDs() {
	if len(s.cfg.DockerSDConfigs) == 0 {
		return
	}
	for i, sd := range s.cfg.DockerSDConfigs {
		s.diags.AddAll(prometheusconvert.ValidateHttpClientConfig(&sd.HTTPClientConfig))
		args := prometheusconvert.ToDiscoveryDocker(sd)
		compLabel := common.GetLabelWithPrefix(s.globalCtx.LabelPrefix, s.cfg.JobName, i)
		s.f.Body().AppendBlock(common.NewBlockWithOverride(
			[]string{"discovery", "docker"},
			compLabel,
			args,
		))
		s.allTargetsExps = append(s.allTargetsExps, "discovery.docker."+compLabel+".targets")
	}
}
