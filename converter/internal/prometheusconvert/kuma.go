package prometheusconvert

import (
	"time"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/discovery/kuma"
	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/converter/internal/common"
	prom_kuma "github.com/prometheus/prometheus/discovery/xds"
)

func appendDiscoveryKuma(pb *prometheusBlocks, label string, sdConfig *prom_kuma.SDConfig) discovery.Exports {
	discoveryKumaArgs := ToDiscoveryKuma(sdConfig)
	name := []string{"discovery", "kuma"}
	block := common.NewBlockWithOverride(name, label, discoveryKumaArgs)
	pb.discoveryBlocks = append(pb.discoveryBlocks, newPrometheusBlock(block, name, label, "", ""))
	return NewDiscoveryExports("discovery.kuma." + label + ".targets")
}

func validateDiscoveryKuma(sdConfig *prom_kuma.SDConfig) diag.Diagnostics {
	return ValidateHttpClientConfig(&sdConfig.HTTPClientConfig)
}

func ToDiscoveryKuma(sdConfig *prom_kuma.SDConfig) *kuma.Arguments {
	if sdConfig == nil {
		return nil
	}

	return &kuma.Arguments{
		Server:          sdConfig.Server,
		RefreshInterval: time.Duration(sdConfig.RefreshInterval),
		FetchTimeout:    time.Duration(sdConfig.FetchTimeout),

		HTTPClientConfig: *ToHttpClientConfig(&sdConfig.HTTPClientConfig),
	}
}
