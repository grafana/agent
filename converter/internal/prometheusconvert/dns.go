package prometheusconvert

import (
	"time"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/discovery/dns"
	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/converter/internal/common"
	prom_dns "github.com/prometheus/prometheus/discovery/dns"
)

func appendDiscoveryDns(pb *prometheusBlocks, label string, sdConfig *prom_dns.SDConfig) discovery.Exports {
	discoveryDnsArgs := toDiscoveryDns(sdConfig)
	name := []string{"discovery", "dns"}
	block := common.NewBlockWithOverride(name, label, discoveryDnsArgs)
	pb.discoveryBlocks = append(pb.discoveryBlocks, newPrometheusBlock(block, name, label))
	return newDiscoverExports("discovery.dns." + label + ".targets")
}

func validateDiscoveryDns(sdConfig *prom_dns.SDConfig) diag.Diagnostics {
	return make(diag.Diagnostics, 0)
}

func toDiscoveryDns(sdConfig *prom_dns.SDConfig) *dns.Arguments {
	if sdConfig == nil {
		return nil
	}

	return &dns.Arguments{
		Names:           sdConfig.Names,
		RefreshInterval: time.Duration(sdConfig.RefreshInterval),
		Type:            sdConfig.Type,
		Port:            sdConfig.Port,
	}
}
