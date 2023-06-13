package prometheusconvert

import (
	"time"

	"github.com/grafana/agent/component/common/config"
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/discovery/digitalocean"
	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/pkg/river/rivertypes"
	"github.com/grafana/agent/pkg/river/token/builder"
	promdigitalocean "github.com/prometheus/prometheus/discovery/digitalocean"
)

func appendDiscoveryDigitalOcean(f *builder.File, jobName string, sdConfig *promdigitalocean.SDConfig) (discovery.Exports, diag.Diagnostics) {
	discoveryDigitalOceanArgs, diags := toDiscoveryDigitalOcean(sdConfig)
	common.AppendBlockWithOverride(f, []string{"discovery", "digitalocean"}, jobName, discoveryDigitalOceanArgs)
	return discovery.Exports{
		// This target map will utilize a RiverTokenize that results in this
		// component export rather than the standard discovery.Target RiverTokenize.
		Targets: []discovery.Target{map[string]string{"__expr__": "discovery.digitalocean." + jobName + ".targets"}},
	}, diags
}

func toDiscoveryDigitalOcean(sdConfig *promdigitalocean.SDConfig) (*digitalocean.Arguments, diag.Diagnostics) {
	if sdConfig == nil {
		return nil, nil
	}

	return &digitalocean.Arguments{
		RefreshInterval: time.Duration(sdConfig.RefreshInterval),
		Port:            sdConfig.Port,
		BearerToken:     rivertypes.Secret(sdConfig.HTTPClientConfig.BearerToken),
		BearerTokenFile: sdConfig.HTTPClientConfig.BearerTokenFile,
		ProxyURL:        config.URL(sdConfig.HTTPClientConfig.ProxyURL),
		FollowRedirects: sdConfig.HTTPClientConfig.FollowRedirects,
		EnableHTTP2:     sdConfig.HTTPClientConfig.EnableHTTP2,
	}, validateDiscoveryDigitalOcean(sdConfig)
}

func validateDiscoveryDigitalOcean(sdConfig *promdigitalocean.SDConfig) diag.Diagnostics {
	return validateHttpClientConfig(&sdConfig.HTTPClientConfig)
}
