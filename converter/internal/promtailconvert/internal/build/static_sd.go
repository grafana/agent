package build

import (
	"github.com/grafana/agent/converter/diag"
	"golang.org/x/exp/maps"
)

func (s *ScrapeConfigBuilder) AppendStaticSDs() {
	if len(s.cfg.ServiceDiscoveryConfig.StaticConfigs) == 0 {
		return
	}

	var allStaticTargets []map[string]string
	for _, sd := range s.cfg.ServiceDiscoveryConfig.StaticConfigs {
		labels := convertPromLabels(sd.Labels)
		for _, target := range sd.Targets {
			mergedTarget := map[string]string{}
			maps.Copy(mergedTarget, labels)
			maps.Copy(mergedTarget, convertPromLabels(target))
			allStaticTargets = append(allStaticTargets, mergedTarget)
		}
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
