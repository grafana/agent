package prometheusconvert

import (
	"time"

	"github.com/grafana/agent/component/common/config"
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/discovery/digitalocean"
	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/river/rivertypes"
	prom_config "github.com/prometheus/common/config"
	prom_digitalocean "github.com/prometheus/prometheus/discovery/digitalocean"
)

func appendDiscoveryDigitalOcean(pb *prometheusBlocks, label string, sdConfig *prom_digitalocean.SDConfig) discovery.Exports {
	discoveryDigitalOceanArgs := toDiscoveryDigitalOcean(sdConfig)
	name := []string{"discovery", "digitalocean"}
	block := common.NewBlockWithOverride(name, label, discoveryDigitalOceanArgs)
	pb.discoveryBlocks = append(pb.discoveryBlocks, newPrometheusBlock(block, name, label, "", ""))
	return NewDiscoveryExports("discovery.digitalocean." + label + ".targets")
}

func validateDiscoveryDigitalOcean(sdConfig *prom_digitalocean.SDConfig) diag.Diagnostics {
	var diags diag.Diagnostics

	diags.AddAll(common.UnsupportedNotEquals(sdConfig.HTTPClientConfig.BasicAuth, nil, "digitalocean_sd_configs basic_auth"))
	diags.AddAll(common.UnsupportedNotEquals(sdConfig.HTTPClientConfig.Authorization, nil, "digitalocean_sd_configs authorization"))
	diags.AddAll(common.UnsupportedNotEquals(sdConfig.HTTPClientConfig.OAuth2, nil, "digitalocean_sd_configs oauth2"))
	diags.AddAll(common.UnsupportedNotDeepEquals(prom_config.TLSConfig{}, sdConfig.HTTPClientConfig.TLSConfig, "digitalocean_sd_configs tls_config"))

	diags.AddAll(ValidateHttpClientConfig(&sdConfig.HTTPClientConfig))

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
