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

func appendDiscoveryEC2(pb *build.PrometheusBlocks, label string, sdConfig *prom_aws.EC2SDConfig) discovery.Exports {
	discoveryec2Args := toDiscoveryEC2(sdConfig)
	name := []string{"discovery", "ec2"}
	block := common.NewBlockWithOverride(name, label, discoveryec2Args)
	pb.DiscoveryBlocks = append(pb.DiscoveryBlocks, build.NewPrometheusBlock(block, name, label, "", ""))
	return common.NewDiscoveryExports("discovery.ec2." + label + ".targets")
}

func ValidateDiscoveryEC2(sdConfig *prom_aws.EC2SDConfig) diag.Diagnostics {
	var diags diag.Diagnostics

	var nilBasicAuth *prom_config.BasicAuth
	var nilAuthorization *prom_config.Authorization
	var nilOAuth2 *prom_config.OAuth2

	diags.AddAll(common.UnsupportedNotEquals(sdConfig.HTTPClientConfig.BasicAuth, nilBasicAuth, "ec2_sd_configs basic_auth"))
	diags.AddAll(common.UnsupportedNotEquals(sdConfig.HTTPClientConfig.Authorization, nilAuthorization, "ec2_sd_configs authorization"))
	diags.AddAll(common.UnsupportedNotEquals(sdConfig.HTTPClientConfig.OAuth2, nilOAuth2, "ec2_sd_configs oauth2"))
	diags.AddAll(common.UnsupportedNotDeepEquals(sdConfig.HTTPClientConfig.BearerToken, prom_config.DefaultHTTPClientConfig.BearerToken, "ec2_sd_configs bearer_token"))
	diags.AddAll(common.UnsupportedNotDeepEquals(sdConfig.HTTPClientConfig.BearerTokenFile, prom_config.DefaultHTTPClientConfig.BearerTokenFile, "ec2_sd_configs bearer_token_file"))
	diags.AddAll(common.UnsupportedNotDeepEquals(sdConfig.HTTPClientConfig.FollowRedirects, prom_config.DefaultHTTPClientConfig.FollowRedirects, "ec2_sd_configs follow_redirects"))
	diags.AddAll(common.UnsupportedNotDeepEquals(sdConfig.HTTPClientConfig.EnableHTTP2, prom_config.DefaultHTTPClientConfig.EnableHTTP2, "ec2_sd_configs enable_http2"))
	diags.AddAll(common.UnsupportedNotDeepEquals(sdConfig.HTTPClientConfig.ProxyConfig, prom_config.DefaultHTTPClientConfig.ProxyConfig, "ec2_sd_configs proxy"))

	// Do a last check in case any of the specific checks missed anything.
	if len(diags) == 0 {
		diags.AddAll(common.UnsupportedNotDeepEquals(sdConfig.HTTPClientConfig, prom_config.DefaultHTTPClientConfig, "ec2_sd_configs http_client_config"))
	}

	return diags
}

func toDiscoveryEC2(sdConfig *prom_aws.EC2SDConfig) *aws.EC2Arguments {
	if sdConfig == nil {
		return nil
	}

	return &aws.EC2Arguments{
		Endpoint:        sdConfig.Endpoint,
		Region:          sdConfig.Region,
		AccessKey:       sdConfig.AccessKey,
		SecretKey:       rivertypes.Secret(sdConfig.SecretKey),
		Profile:         sdConfig.Profile,
		RoleARN:         sdConfig.RoleARN,
		RefreshInterval: time.Duration(sdConfig.RefreshInterval),
		Port:            sdConfig.Port,
		Filters:         toEC2Filters(sdConfig.Filters),
	}
}

func toEC2Filters(filtersConfig []*prom_aws.EC2Filter) []*aws.EC2Filter {
	filters := make([]*aws.EC2Filter, 0)

	for _, filter := range filtersConfig {
		filters = append(filters, &aws.EC2Filter{
			Name:   filter.Name,
			Values: filter.Values,
		})
	}

	return filters
}
