package prometheusconvert

import (
	"time"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/discovery/aws"
	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/river/rivertypes"
	prom_config "github.com/prometheus/common/config"
	prom_aws "github.com/prometheus/prometheus/discovery/aws"
)

func appendDiscoveryLightsail(pb *prometheusBlocks, label string, sdConfig *prom_aws.LightsailSDConfig) discovery.Exports {
	discoverylightsailArgs := toDiscoveryLightsail(sdConfig)
	name := []string{"discovery", "lightsail"}
	block := common.NewBlockWithOverride(name, label, discoverylightsailArgs)
	pb.discoveryBlocks = append(pb.discoveryBlocks, newPrometheusBlock(block, name, label, "", ""))
	return NewDiscoveryExports("discovery.lightsail." + label + ".targets")
}

func validateDiscoveryLightsail(sdConfig *prom_aws.LightsailSDConfig) diag.Diagnostics {
	var diags diag.Diagnostics

	diags.AddAll(common.UnsupportedNotEquals(sdConfig.HTTPClientConfig.BasicAuth, nil, "lightsail_sd_configs basic_auth"))
	diags.AddAll(common.UnsupportedNotEquals(sdConfig.HTTPClientConfig.Authorization, nil, "lightsail_sd_configs authorization"))
	diags.AddAll(common.UnsupportedNotEquals(sdConfig.HTTPClientConfig.OAuth2, nil, "lightsail_sd_configs oauth2"))
	diags.AddAll(common.UnsupportedNotDeepEquals(sdConfig.HTTPClientConfig.BearerToken, prom_config.DefaultHTTPClientConfig.BearerToken, "lightsail_sd_configs bearer_token"))
	diags.AddAll(common.UnsupportedNotDeepEquals(sdConfig.HTTPClientConfig.BearerTokenFile, prom_config.DefaultHTTPClientConfig.BearerTokenFile, "lightsail_sd_configs bearer_token_file"))
	diags.AddAll(common.UnsupportedNotDeepEquals(sdConfig.HTTPClientConfig.FollowRedirects, prom_config.DefaultHTTPClientConfig.FollowRedirects, "lightsail_sd_configs follow_redirects"))
	diags.AddAll(common.UnsupportedNotDeepEquals(sdConfig.HTTPClientConfig.EnableHTTP2, prom_config.DefaultHTTPClientConfig.EnableHTTP2, "lightsail_sd_configs enable_http2"))
	diags.AddAll(common.UnsupportedNotDeepEquals(sdConfig.HTTPClientConfig.ProxyConfig, prom_config.DefaultHTTPClientConfig.ProxyConfig, "lightsail_sd_configs proxy"))

	// Do a last check in case any of the specific checks missed anything.
	if len(diags) == 0 {
		diags.AddAll(common.UnsupportedNotDeepEquals(sdConfig.HTTPClientConfig, prom_config.DefaultHTTPClientConfig, "lightsail_sd_configs http_client_config"))
	}

	return diags
}

func toDiscoveryLightsail(sdConfig *prom_aws.LightsailSDConfig) *aws.LightsailArguments {
	if sdConfig == nil {
		return nil
	}

	return &aws.LightsailArguments{
		Endpoint:        sdConfig.Endpoint,
		Region:          sdConfig.Region,
		AccessKey:       sdConfig.AccessKey,
		SecretKey:       rivertypes.Secret(sdConfig.SecretKey),
		Profile:         sdConfig.Profile,
		RoleARN:         sdConfig.RoleARN,
		RefreshInterval: time.Duration(sdConfig.RefreshInterval),
		Port:            sdConfig.Port,
	}
}
