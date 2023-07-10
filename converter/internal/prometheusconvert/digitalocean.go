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
	prom_config "github.com/prometheus/common/config"
	prom_digitalocean "github.com/prometheus/prometheus/discovery/digitalocean"
)

func appendDiscoveryDigitalOcean(pb *prometheusBlocks, label string, sdConfig *prom_digitalocean.SDConfig) discovery.Exports {
	discoveryDigitalOceanArgs := toDiscoveryDigitalOcean(sdConfig)
	block := common.NewBlockWithOverride([]string{"discovery", "digitalocean"}, label, discoveryDigitalOceanArgs)
	pb.discoveryBlocks = append(pb.discoveryBlocks, block)
	return newDiscoverExports("discovery.digitalocean." + label + ".targets")
}

func validateDiscoveryDigitalOcean(sdConfig *prom_digitalocean.SDConfig) diag.Diagnostics {
	var diags diag.Diagnostics

	if sdConfig.HTTPClientConfig.BasicAuth != nil {
		diags.Add(diag.SeverityLevelError, "unsupported basic_auth for digitalocean_sd_configs")
	}

	if sdConfig.HTTPClientConfig.Authorization != nil {
		diags.Add(diag.SeverityLevelError, "unsupported authorization for digitalocean_sd_configs")
	}

	if sdConfig.HTTPClientConfig.OAuth2 != nil {
		diags.Add(diag.SeverityLevelError, "unsupported oauth2 for digitalocean_sd_configs")
	}

	if !reflect.DeepEqual(prom_config.TLSConfig{}, sdConfig.HTTPClientConfig.TLSConfig) {
		diags.Add(diag.SeverityLevelError, "unsupported oauth2 for digitalocean_sd_configs")
	}

	newDiags := validateHttpClientConfig(&sdConfig.HTTPClientConfig)

	diags = append(diags, newDiags...)
	return diags
}

func toDiscoveryDigitalOcean(sdConfig *prom_digitalocean.SDConfig) *digitalocean.Arguments {
	if sdConfig == nil {
		return nil
	}

	return &digitalocean.Arguments{
		RefreshInterval: time.Duration(sdConfig.RefreshInterval),
		Port:            sdConfig.Port,
		BearerToken:     rivertypes.Secret(sdConfig.HTTPClientConfig.BearerToken),
		BearerTokenFile: sdConfig.HTTPClientConfig.BearerTokenFile,
		ProxyURL:        config.URL(sdConfig.HTTPClientConfig.ProxyURL),
		FollowRedirects: sdConfig.HTTPClientConfig.FollowRedirects,
		EnableHTTP2:     sdConfig.HTTPClientConfig.EnableHTTP2,
	}
}
