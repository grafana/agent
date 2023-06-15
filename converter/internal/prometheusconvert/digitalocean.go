package prometheusconvert

import (
	"reflect"
	"time"

	"github.com/grafana/agent/component/common/config"
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/discovery/digitalocean"
	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/pkg/river/rivertypes"
	"github.com/grafana/agent/pkg/river/token/builder"
	promconfig "github.com/prometheus/common/config"
	promdigitalocean "github.com/prometheus/prometheus/discovery/digitalocean"
)

func appendDiscoveryDigitalOcean(f *builder.File, jobName string, sdConfig *promdigitalocean.SDConfig) (discovery.Exports, diag.Diagnostics) {
	discoveryDigitalOceanArgs, diags := toDiscoveryDigitalOcean(sdConfig)
	common.AppendBlockWithOverride(f, []string{"discovery", "digitalocean"}, jobName, discoveryDigitalOceanArgs)
	return newDiscoverExports("discovery.digitalocean." + jobName + ".targets"), diags
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
	var diags diag.Diagnostics

	if sdConfig.HTTPClientConfig.BasicAuth != nil {
		diags.Add(diag.SeverityLevelWarn, "unsupported basic_auth for digitalocean_sd_configs")
	}

	if sdConfig.HTTPClientConfig.Authorization != nil {
		diags.Add(diag.SeverityLevelWarn, "unsupported authorization for digitalocean_sd_configs")
	}

	if sdConfig.HTTPClientConfig.OAuth2 != nil {
		diags.Add(diag.SeverityLevelWarn, "unsupported oauth2 for digitalocean_sd_configs")
	}

	if !reflect.DeepEqual(promconfig.TLSConfig{}, sdConfig.HTTPClientConfig.TLSConfig) {
		diags.Add(diag.SeverityLevelWarn, "unsupported oauth2 for digitalocean_sd_configs")
	}

	newDiags := validateHttpClientConfig(&sdConfig.HTTPClientConfig)

	diags = append(diags, newDiags...)
	return diags
}
