package component

import (
	"time"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/discovery/aws"
	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/converter/internal/prometheusconvert/build"
	"github.com/grafana/river/rivertypes"
	prom_config "github.com/prometheus/common/config"
	prom_aws "github.com/prometheus/prometheus/discovery/aws"
)

func appendDiscoveryLightsail(pb *build.PrometheusBlocks, label string, sdConfig *prom_aws.LightsailSDConfig) discovery.Exports {
	discoverylightsailArgs := toDiscoveryLightsail(sdConfig)
	name := []string{"discovery", "lightsail"}
	block := common.NewBlockWithOverride(name, label, discoverylightsailArgs)
	pb.DiscoveryBlocks = append(pb.DiscoveryBlocks, build.NewPrometheusBlock(block, name, label, "", ""))
	return common.NewDiscoveryExports("discovery.lightsail." + label + ".targets")
}

func ValidateDiscoveryLightsail(sdConfig *prom_aws.LightsailSDConfig) diag.Diagnostics {
	var diags diag.Diagnostics

	var nilBasicAuth *prom_config.BasicAuth
	var nilAuthorization *prom_config.Authorization
	var nilOAuth2 *prom_config.OAuth2

	diags.AddAll(common.ValidateSupported(common.NotEquals, sdConfig.HTTPClientConfig.BasicAuth, nilBasicAuth, "lightsail_sd_configs basic_auth", ""))
	diags.AddAll(common.ValidateSupported(common.NotEquals, sdConfig.HTTPClientConfig.Authorization, nilAuthorization, "lightsail_sd_configs authorization", ""))
	diags.AddAll(common.ValidateSupported(common.NotEquals, sdConfig.HTTPClientConfig.OAuth2, nilOAuth2, "lightsail_sd_configs oauth2", ""))
	diags.AddAll(common.ValidateSupported(common.NotDeepEquals, sdConfig.HTTPClientConfig.BearerToken, prom_config.DefaultHTTPClientConfig.BearerToken, "lightsail_sd_configs bearer_token", ""))
	diags.AddAll(common.ValidateSupported(common.NotDeepEquals, sdConfig.HTTPClientConfig.BearerTokenFile, prom_config.DefaultHTTPClientConfig.BearerTokenFile, "lightsail_sd_configs bearer_token_file", ""))
	diags.AddAll(common.ValidateSupported(common.NotDeepEquals, sdConfig.HTTPClientConfig.FollowRedirects, prom_config.DefaultHTTPClientConfig.FollowRedirects, "lightsail_sd_configs follow_redirects", ""))
	diags.AddAll(common.ValidateSupported(common.NotDeepEquals, sdConfig.HTTPClientConfig.EnableHTTP2, prom_config.DefaultHTTPClientConfig.EnableHTTP2, "lightsail_sd_configs enable_http2", ""))
	diags.AddAll(common.ValidateSupported(common.NotDeepEquals, sdConfig.HTTPClientConfig.ProxyConfig, prom_config.DefaultHTTPClientConfig.ProxyConfig, "lightsail_sd_configs proxy", ""))

	// Do a last check in case any of the specific checks missed anything.
	if len(diags) == 0 {
		diags.AddAll(common.ValidateSupported(common.NotDeepEquals, sdConfig.HTTPClientConfig, prom_config.DefaultHTTPClientConfig, "lightsail_sd_configs http_client_config", ""))
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
