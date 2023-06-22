package prometheusconvert

import (
	"reflect"
	"time"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/discovery/aws"
	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/pkg/river/rivertypes"
	"github.com/grafana/agent/pkg/river/token/builder"
	promconfig "github.com/prometheus/common/config"
	promaws "github.com/prometheus/prometheus/discovery/aws"
)

func appendDiscoveryEC2(f *builder.File, label string, sdConfig *promaws.EC2SDConfig) (discovery.Exports, diag.Diagnostics) {
	discoveryec2Args, diags := toDiscoveryEC2(sdConfig)
	common.AppendBlockWithOverride(f, []string{"discovery", "ec2"}, label, discoveryec2Args)
	return newDiscoverExports("discovery.ec2." + label + ".targets"), diags
}

func toDiscoveryEC2(sdConfig *promaws.EC2SDConfig) (*aws.EC2Arguments, diag.Diagnostics) {
	if sdConfig == nil {
		return nil, nil
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
	}, validateDiscoveryEC2(sdConfig)
}

func validateDiscoveryEC2(sdConfig *promaws.EC2SDConfig) diag.Diagnostics {
	var diags diag.Diagnostics

	if sdConfig.HTTPClientConfig.BasicAuth != nil {
		diags.Add(diag.SeverityLevelWarn, "unsupported basic_auth for ec2_sd_configs")
	}

	if sdConfig.HTTPClientConfig.Authorization != nil {
		diags.Add(diag.SeverityLevelWarn, "unsupported authorization for ec2_sd_configs")
	}

	if sdConfig.HTTPClientConfig.OAuth2 != nil {
		diags.Add(diag.SeverityLevelWarn, "unsupported oauth2 for ec2_sd_configs")
	}

	if !reflect.DeepEqual(promconfig.TLSConfig{}, sdConfig.HTTPClientConfig.TLSConfig) {
		diags.Add(diag.SeverityLevelWarn, "unsupported oauth2 for ec2_sd_configs")
	}

	newDiags := validateHttpClientConfig(&sdConfig.HTTPClientConfig)

	diags = append(diags, newDiags...)
	return diags
}

func toEC2Filters(filtersConfig []*promaws.EC2Filter) []*aws.EC2Filter {
	filters := make([]*aws.EC2Filter, 0)

	for _, filter := range filtersConfig {
		filters = append(filters, &aws.EC2Filter{
			Name:   filter.Name,
			Values: filter.Values,
		})
	}

	return filters
}
