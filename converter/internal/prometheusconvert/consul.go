package prometheusconvert

import (
	"time"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/discovery/consul"
	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/pkg/river/rivertypes"
	prom_consul "github.com/prometheus/prometheus/discovery/consul"
)

func appendDiscoveryConsul(pb *prometheusBlocks, label string, sdConfig *prom_consul.SDConfig) (discovery.Exports, diag.Diagnostics) {
	discoveryConsulArgs, diags := toDiscoveryConsul(sdConfig)
	block := common.NewBlockWithOverride([]string{"discovery", "consul"}, label, discoveryConsulArgs)
	pb.discoveryBlocks = append(pb.discoveryBlocks, block)
	return newDiscoverExports("discovery.consul." + label + ".targets"), diags
}

func toDiscoveryConsul(sdConfig *prom_consul.SDConfig) (*consul.Arguments, diag.Diagnostics) {
	if sdConfig == nil {
		return nil, nil
	}

	return &consul.Arguments{
		Server:           sdConfig.Server,
		Token:            rivertypes.Secret(sdConfig.Token),
		Datacenter:       sdConfig.Datacenter,
		Namespace:        sdConfig.Namespace,
		Partition:        sdConfig.Partition,
		TagSeparator:     sdConfig.TagSeparator,
		Scheme:           sdConfig.Scheme,
		Username:         sdConfig.Username,
		Password:         rivertypes.Secret(sdConfig.Password),
		AllowStale:       sdConfig.AllowStale,
		Services:         sdConfig.Services,
		ServiceTags:      sdConfig.ServiceTags,
		NodeMeta:         sdConfig.NodeMeta,
		RefreshInterval:  time.Duration(sdConfig.RefreshInterval),
		HTTPClientConfig: *toHttpClientConfig(&sdConfig.HTTPClientConfig),
	}, validateDiscoveryConsul(sdConfig)
}

func validateDiscoveryConsul(sdConfig *prom_consul.SDConfig) diag.Diagnostics {
	return validateHttpClientConfig(&sdConfig.HTTPClientConfig)
}
