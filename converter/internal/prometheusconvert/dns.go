package prometheusconvert

import (
	"time"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/discovery/dns"
	"github.com/grafana/agent/converter/internal/common"
	prom_dns "github.com/prometheus/prometheus/discovery/dns"
)

func appendDiscoveryDns(pb *prometheusBlocks, label string, sdConfig *prom_dns.SDConfig) discovery.Exports {
	discoveryDnsArgs := toDiscoveryDns(sdConfig)
	pb.discoveryBlocks = common.AppendNewBlockWithOverride(pb.discoveryBlocks, []string{"discovery", "dns"}, label, discoveryDnsArgs)
	return newDiscoverExports("discovery.dns." + label + ".targets")
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
