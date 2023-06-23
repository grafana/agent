package prometheusconvert

import (
	"time"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/discovery/gce"
	"github.com/grafana/agent/converter/internal/common"
	prom_gce "github.com/prometheus/prometheus/discovery/gce"
)

func appendDiscoveryGCE(pb *prometheusBlocks, label string, sdConfig *prom_gce.SDConfig) discovery.Exports {
	discoveryGCEArgs := toDiscoveryGCE(sdConfig)
	pb.discoveryBlocks = common.AppendNewBlockWithOverride(pb.discoveryBlocks, []string{"discovery", "gce"}, label, discoveryGCEArgs)
	return newDiscoverExports("discovery.gce." + label + ".targets")
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
