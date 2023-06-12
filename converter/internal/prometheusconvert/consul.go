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

func appendDiscoveryConsul(f *builder.File, jobName string, sdConfig *promconsul.SDConfig) (discovery.Exports, diag.Diagnostics) {
	discoveryConsulArgs, diags := toDiscoveryConsul(sdConfig)
	common.AppendBlockWithOverride(f, []string{"discovery", "consul"}, jobName, discoveryConsulArgs)
	return discovery.Exports{
		// This target map will utilize a RiverTokenize that results in this
		// component export rather than the standard discovery.Target RiverTokenize.
		Targets: []discovery.Target{map[string]string{"discovery.consul." + jobName + ".targets": ""}},
	}, diags
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
	var diags diag.Diagnostics

	if sdConfig.HTTPClientConfig.NoProxy != "" {
		diags.Add(diag.SeverityLevelWarn, "unsupported consul service discovery config no_proxy was provided")
	}

	if sdConfig.HTTPClientConfig.ProxyFromEnvironment {
		diags.Add(diag.SeverityLevelWarn, "unsupported consul service discovery config proxy_from_environment was provided")
	}

	if len(sdConfig.HTTPClientConfig.ProxyConnectHeader) > 0 {
		diags.Add(diag.SeverityLevelWarn, "unsupported consul service discovery config proxy_connect_header was provided")
	}

	return diags
}
