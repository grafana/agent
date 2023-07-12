package prometheusconvert

import (
	"time"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/discovery/gce"
	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/converter/internal/common"
	prom_gce "github.com/prometheus/prometheus/discovery/gce"
)

func appendDiscoveryGCE(pb *prometheusBlocks, label string, sdConfig *prom_gce.SDConfig) discovery.Exports {
	discoveryGCEArgs := toDiscoveryGCE(sdConfig)
	name := []string{"discovery", "gce"}
	block := common.NewBlockWithOverride(name, label, discoveryGCEArgs)
	pb.discoveryBlocks = append(pb.discoveryBlocks, newPrometheusBlock(block, name, label, "", ""))
	return newDiscoverExports("discovery.gce." + label + ".targets")
}

func validateDiscoveryGce(sdConfig *prom_gce.SDConfig) diag.Diagnostics {
	return make(diag.Diagnostics, 0)
}

func toDiscoveryGCE(sdConfig *prom_gce.SDConfig) *gce.Arguments {
	if sdConfig == nil {
		return nil
	}

	return &gce.Arguments{
		Project:         sdConfig.Project,
		Zone:            sdConfig.Zone,
		Filter:          sdConfig.Filter,
		RefreshInterval: time.Duration(sdConfig.RefreshInterval),
		Port:            sdConfig.Port,
		TagSeparator:    sdConfig.TagSeparator,
	}
}
