package component

import (
	"time"

	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/converter/internal/prometheusconvert/build"
	"github.com/grafana/agent/internal/component/discovery"
	"github.com/grafana/agent/internal/component/discovery/ionos"
	prom_ionos "github.com/prometheus/prometheus/discovery/ionos"
)

func appendDiscoveryIonos(pb *build.PrometheusBlocks, label string, sdConfig *prom_ionos.SDConfig) discovery.Exports {
	discoveryIonosArgs := toDiscoveryIonos(sdConfig)
	name := []string{"discovery", "ionos"}
	block := common.NewBlockWithOverride(name, label, discoveryIonosArgs)
	pb.DiscoveryBlocks = append(pb.DiscoveryBlocks, build.NewPrometheusBlock(block, name, label, "", ""))
	return common.NewDiscoveryExports("discovery.ionos." + label + ".targets")
}

func ValidateDiscoveryIonos(sdConfig *prom_ionos.SDConfig) diag.Diagnostics {
	return common.ValidateHttpClientConfig(&sdConfig.HTTPClientConfig)
}

func toDiscoveryIonos(sdConfig *prom_ionos.SDConfig) *ionos.Arguments {
	if sdConfig == nil {
		return nil
	}

	return &ionos.Arguments{
		DatacenterID:     sdConfig.DatacenterID,
		RefreshInterval:  time.Duration(sdConfig.RefreshInterval),
		Port:             sdConfig.Port,
		HTTPClientConfig: *common.ToHttpClientConfig(&sdConfig.HTTPClientConfig),
	}
}
