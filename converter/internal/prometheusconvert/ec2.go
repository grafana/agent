package prometheusconvert

import (
	"reflect"
	"time"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/discovery/aws"
	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/pkg/river/rivertypes"
	prom_config "github.com/prometheus/common/config"
	prom_aws "github.com/prometheus/prometheus/discovery/aws"
)

func appendDiscoveryEC2(pb *prometheusBlocks, label string, sdConfig *prom_aws.EC2SDConfig) discovery.Exports {
	discoveryec2Args := toDiscoveryEC2(sdConfig)
	block := common.NewBlockWithOverride([]string{"discovery", "ec2"}, label, discoveryec2Args)
	pb.discoveryBlocks = append(pb.discoveryBlocks, block)
	return newDiscoverExports("discovery.ec2." + label + ".targets")
}

func validateDiscoveryEC2(sdConfig *prom_aws.EC2SDConfig) diag.Diagnostics {
	var diags diag.Diagnostics

	if sdConfig.HTTPClientConfig.BasicAuth != nil {
		diags.Add(diag.SeverityLevelError, "unsupported basic_auth for ec2_sd_configs")
	}

	if sdConfig.HTTPClientConfig.Authorization != nil {
		diags.Add(diag.SeverityLevelError, "unsupported authorization for ec2_sd_configs")
	}

	if sdConfig.HTTPClientConfig.OAuth2 != nil {
		diags.Add(diag.SeverityLevelError, "unsupported oauth2 for ec2_sd_configs")
	}

	if !reflect.DeepEqual(sdConfig.HTTPClientConfig.BearerToken, prom_config.DefaultHTTPClientConfig.BearerToken) {
		diags.Add(diag.SeverityLevelError, "unsupported bearer_token for ec2_sd_configs")
	}

	if !reflect.DeepEqual(sdConfig.HTTPClientConfig.BearerTokenFile, prom_config.DefaultHTTPClientConfig.BearerTokenFile) {
		diags.Add(diag.SeverityLevelError, "unsupported bearer_token_file for ec2_sd_configs")
	}

	if !reflect.DeepEqual(sdConfig.HTTPClientConfig.FollowRedirects, prom_config.DefaultHTTPClientConfig.FollowRedirects) {
		diags.Add(diag.SeverityLevelError, "unsupported follow_redirects for ec2_sd_configs")
	}

	if !reflect.DeepEqual(sdConfig.HTTPClientConfig.EnableHTTP2, prom_config.DefaultHTTPClientConfig.EnableHTTP2) {
		diags.Add(diag.SeverityLevelError, "unsupported enable_http2 for ec2_sd_configs")
	}

	if !reflect.DeepEqual(sdConfig.HTTPClientConfig.ProxyConfig, prom_config.DefaultHTTPClientConfig.ProxyConfig) {
		diags.Add(diag.SeverityLevelError, "unsupported proxy for ec2_sd_configs")
	}

	// Do a last check in case any of the specific checks missed anything.
	if len(diags) == 0 && !reflect.DeepEqual(sdConfig.HTTPClientConfig, prom_config.DefaultHTTPClientConfig) {
		diags.Add(diag.SeverityLevelError, "unsupported http_client_config for ec2_sd_configs")
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
