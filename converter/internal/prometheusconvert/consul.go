package prometheusconvert

import (
	"time"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/discovery/consul"
	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/pkg/river/rivertypes"
	"github.com/grafana/agent/pkg/river/token/builder"
	promconsul "github.com/prometheus/prometheus/discovery/consul"
)

func appendDiscoveryConsul(f *builder.File, label string, sdConfig *promconsul.SDConfig) (discovery.Exports, diag.Diagnostics) {
	discoveryConsulArgs, diags := toDiscoveryConsul(sdConfig)
	common.AppendBlockWithOverride(f, []string{"discovery", "consul"}, label, discoveryConsulArgs)
	return newDiscoverExports("discovery.consul." + label + ".targets"), diags
}

func toDiscoveryConsul(sdConfig *promconsul.SDConfig) (*consul.Arguments, diag.Diagnostics) {
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

func validateDiscoveryConsul(sdConfig *promconsul.SDConfig) diag.Diagnostics {
	return validateHttpClientConfig(&sdConfig.HTTPClientConfig)
}
