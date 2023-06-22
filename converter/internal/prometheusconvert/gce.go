package prometheusconvert

import (
	"time"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/discovery/gce"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/pkg/river/token/builder"
	promgce "github.com/prometheus/prometheus/discovery/gce"
)

func appendDiscoveryGCE(f *builder.File, label string, sdConfig *promgce.SDConfig) discovery.Exports {
	discoveryGCEArgs := toDiscoveryGCE(sdConfig)
	common.AppendBlockWithOverride(f, []string{"discovery", "gce"}, label, discoveryGCEArgs)
	return newDiscoverExports("discovery.gce." + label + ".targets")
}

func toDiscoveryGCE(sdConfig *promgce.SDConfig) *gce.Arguments {
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
