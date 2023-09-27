package prometheusconvert

import (
	"time"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/discovery/linode"
	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/converter/internal/common"
	prom_linode "github.com/prometheus/prometheus/discovery/linode"
)

func appendDiscoveryLinode(pb *prometheusBlocks, label string, sdConfig *prom_linode.SDConfig) discovery.Exports {
	discoveryLinodeArgs := ToDiscoveryLinode(sdConfig)
	name := []string{"discovery", "linode"}
	block := common.NewBlockWithOverride(name, label, discoveryLinodeArgs)
	pb.discoveryBlocks = append(pb.discoveryBlocks, newPrometheusBlock(block, name, label, "", ""))
	return NewDiscoveryExports("discovery.linode." + label + ".targets")
}

func validateDiscoveryLinode(sdConfig *prom_linode.SDConfig) diag.Diagnostics {
	return ValidateHttpClientConfig(&sdConfig.HTTPClientConfig)
}

func ToDiscoveryLinode(sdConfig *prom_linode.SDConfig) *linode.Arguments {
	if sdConfig == nil {
		return nil
	}

	return &linode.Arguments{
		RefreshInterval:  time.Duration(sdConfig.RefreshInterval),
		Port:             sdConfig.Port,
		TagSeparator:     sdConfig.TagSeparator,
		HTTPClientConfig: *ToHttpClientConfig(&sdConfig.HTTPClientConfig),
	}
}
