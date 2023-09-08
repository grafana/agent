package prometheusconvert

import (
	"time"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/discovery/triton"
	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/converter/internal/common"
	prom_triton "github.com/prometheus/prometheus/discovery/triton"
)

func appendDiscoveryTriton(pb *prometheusBlocks, label string, sdConfig *prom_triton.SDConfig) discovery.Exports {
	discoveryTritonArgs := ToDiscoveryTriton(sdConfig)
	name := []string{"discovery", "triton"}
	block := common.NewBlockWithOverride(name, label, discoveryTritonArgs)
	pb.discoveryBlocks = append(pb.discoveryBlocks, newPrometheusBlock(block, name, label, "", ""))
	return NewDiscoveryExports("discovery.triton." + label + ".targets")
}

func validateDiscoveryTriton(sdConfig *prom_triton.SDConfig) diag.Diagnostics {
	return nil
}

func ToDiscoveryTriton(sdConfig *prom_triton.SDConfig) *triton.Arguments {
	if sdConfig == nil {
		return nil
	}

	return &triton.Arguments{
		Account:         sdConfig.Account,
		Role:            sdConfig.Role,
		DNSSuffix:       sdConfig.DNSSuffix,
		Endpoint:        sdConfig.Endpoint,
		Groups:          sdConfig.Groups,
		Port:            sdConfig.Port,
		RefreshInterval: time.Duration(sdConfig.RefreshInterval),
		Version:         sdConfig.Version,
		TLSConfig:       *ToTLSConfig(&sdConfig.TLSConfig),
	}
}
