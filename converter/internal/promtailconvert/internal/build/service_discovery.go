package build

import (
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/converter/internal/prometheusconvert"

	"github.com/grafana/loki/clients/pkg/promtail/scrapeconfig"
	prom_discover "github.com/prometheus/prometheus/discovery"
)

func (s *ScrapeConfigBuilder) AppendSDs() {
	sdConfigs := toDiscoveryConfig(s.cfg)
	if len(sdConfigs) == 0 {
		return
	}

	pb := prometheusconvert.NewPrometheusBlocks()
	targets := prometheusconvert.AppendServiceDiscoveryConfigs(pb, sdConfigs, common.LabelForParts(s.globalCtx.LabelPrefix, s.cfg.JobName))
	pb.AppendToFile(s.f)

	targetLiterals := make([]discovery.Target, 0)
	for _, target := range targets {
		if expr, ok := target["__expr__"]; ok {
			// use the expression if __expr__ is set
			s.allTargetsExps = append(s.allTargetsExps, expr)
		} else {
			// otherwise, use the target as a literal
			targetLiterals = append(targetLiterals, target)
		}
	}

	// write the target literals as a string if there are any
	if len(targetLiterals) != 0 {
		literalsStr, err := toRiverExpression(targetLiterals)
		if err != nil { // should not happen, unless we have a bug
			s.diags.Add(
				diag.SeverityLevelCritical,
				"failed to write static SD targets as valid River expression: "+err.Error(),
			)
		}
		s.allTargetsExps = append(s.allTargetsExps, literalsStr)
	}

	s.diags.AddAll(prometheusconvert.ValidateServiceDiscoveryConfigs(sdConfigs))
}

func toDiscoveryConfig(cfg *scrapeconfig.Config) prom_discover.Configs {
	sdConfigs := make(prom_discover.Configs, 0)

	for _, sd := range cfg.ServiceDiscoveryConfig.AzureSDConfigs {
		sdConfigs = append(sdConfigs, sd)
	}

	for _, sd := range cfg.ServiceDiscoveryConfig.ConsulSDConfigs {
		sdConfigs = append(sdConfigs, sd)
	}

	for _, sd := range cfg.ServiceDiscoveryConfig.DigitalOceanSDConfigs {
		sdConfigs = append(sdConfigs, sd)
	}

	for _, sd := range cfg.ServiceDiscoveryConfig.DockerSwarmSDConfigs {
		sdConfigs = append(sdConfigs, sd)
	}

	for _, sd := range cfg.DockerSDConfigs {
		sdConfigs = append(sdConfigs, sd)
	}

	for _, sd := range cfg.ServiceDiscoveryConfig.DNSSDConfigs {
		sdConfigs = append(sdConfigs, sd)
	}

	for _, sd := range cfg.ServiceDiscoveryConfig.EC2SDConfigs {
		sdConfigs = append(sdConfigs, sd)
	}

	for _, sd := range cfg.ServiceDiscoveryConfig.FileSDConfigs {
		sdConfigs = append(sdConfigs, sd)
	}

	for _, sd := range cfg.ServiceDiscoveryConfig.GCESDConfigs {
		sdConfigs = append(sdConfigs, sd)
	}

	for _, sd := range cfg.ServiceDiscoveryConfig.KubernetesSDConfigs {
		sdConfigs = append(sdConfigs, sd)
	}

	for _, sd := range cfg.ServiceDiscoveryConfig.MarathonSDConfigs {
		sdConfigs = append(sdConfigs, sd)
	}

	for _, sd := range cfg.ServiceDiscoveryConfig.NerveSDConfigs {
		sdConfigs = append(sdConfigs, sd)
	}

	for _, sd := range cfg.ServiceDiscoveryConfig.OpenstackSDConfigs {
		sdConfigs = append(sdConfigs, sd)
	}

	for _, sd := range cfg.ServiceDiscoveryConfig.ServersetSDConfigs {
		sdConfigs = append(sdConfigs, sd)
	}

	if len(cfg.ServiceDiscoveryConfig.StaticConfigs) != 0 {
		sdConfigs = append(sdConfigs, cfg.ServiceDiscoveryConfig.StaticConfigs)
	}

	for _, sd := range cfg.ServiceDiscoveryConfig.TritonSDConfigs {
		sdConfigs = append(sdConfigs, sd)
	}

	return sdConfigs
}
