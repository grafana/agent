package prometheusconvert

import (
	"time"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/discovery/ionos"
	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/converter/internal/common"
	prom_ionos "github.com/prometheus/prometheus/discovery/ionos"
)

func appendDiscoveryIonos(pb *prometheusBlocks, label string, sdConfig *prom_ionos.SDConfig) discovery.Exports {
	discoveryIonosArgs := toDiscoveryIonos(sdConfig)
	name := []string{"discovery", "ionos"}
	block := common.NewBlockWithOverride(name, label, discoveryIonosArgs)
	pb.discoveryBlocks = append(pb.discoveryBlocks, newPrometheusBlock(block, name, label, "", ""))
	return NewDiscoveryExports("discovery.ionos." + label + ".targets")
}

func validateDiscoveryIonos(sdConfig *prom_ionos.SDConfig) diag.Diagnostics {
	return ValidateHttpClientConfig(&sdConfig.HTTPClientConfig)
}

func toDiscoveryIonos(sdConfig *prom_ionos.SDConfig) *ionos.Arguments {
	if sdConfig == nil {
		return nil
	}

	return &ionos.Arguments{
		DatacenterID:     sdConfig.DatacenterID,
		RefreshInterval:  time.Duration(sdConfig.RefreshInterval),
		Port:             sdConfig.Port,
		HTTPClientConfig: *ToHttpClientConfig(&sdConfig.HTTPClientConfig),
	}
}
