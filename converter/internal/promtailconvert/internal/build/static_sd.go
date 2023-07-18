package build

import (
	"github.com/grafana/agent/converter/diag"
)

func (s *ScrapeConfigBuilder) AppendStaticSDs() {
	if len(s.cfg.ServiceDiscoveryConfig.StaticConfigs) == 0 {
		return
	}

	var allStaticTargets []map[string]string
	for _, sd := range s.cfg.ServiceDiscoveryConfig.StaticConfigs {
		allStaticTargets = append(allStaticTargets, convertPromLabels(sd.Labels))
	}

	targetsExpr, err := toRiverExpression(allStaticTargets)
	if err != nil {
		s.diags.Add(
			diag.SeverityLevelCritical,
			"failed to write static SD targets as valid River expression: "+err.Error(),
		)
	}

	s.allTargetsExps = append(s.allTargetsExps, targetsExpr)
}
