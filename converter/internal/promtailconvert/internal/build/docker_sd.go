package build

import (
	"fmt"

	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/converter/internal/prometheusconvert"
)

func (s *ScrapeConfigBuilder) AppendDockerSDs() {
	if len(s.cfg.DockerSDConfigs) == 0 {
		return
	}
	for i, sd := range s.cfg.DockerSDConfigs {
		s.diags.AddAll(prometheusconvert.ValidateHttpClientConfig(&sd.HTTPClientConfig))
		compName := fmt.Sprintf("%s_%d", s.cfg.JobName, i)

		args := prometheusconvert.ToDiscoveryDocker(sd)
		s.f.Body().AppendBlock(common.NewBlockWithOverride(
			[]string{"discovery", "docker"},
			compName,
			args,
		))
		s.allTargetsExps = append(s.allTargetsExps, "discovery.docker."+compName+".targets")
	}
}
