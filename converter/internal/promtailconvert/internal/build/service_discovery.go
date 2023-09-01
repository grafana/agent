package build

import (
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

	for _, target := range targets {
		s.allTargetsExps = append(s.allTargetsExps, target["__expr__"])
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

	for _, sd := range cfg.ServiceDiscoveryConfig.ConsulAgentSDConfigs {
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

	for _, sd := range cfg.ServiceDiscoveryConfig.TritonSDConfigs {
		sdConfigs = append(sdConfigs, sd)
	}

	return sdConfigs
}
