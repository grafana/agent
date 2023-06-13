package prometheusconvert

import (
	"time"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/discovery/aws"
	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/pkg/river/rivertypes"
	"github.com/grafana/agent/pkg/river/token/builder"
	promaws "github.com/prometheus/prometheus/discovery/aws"
)

func appendDiscoveryEC2(f *builder.File, jobName string, sdConfig *promaws.EC2SDConfig) (discovery.Exports, diag.Diagnostics) {
	discoveryec2Args, diags := toDiscoveryEC2(sdConfig)
	common.AppendBlockWithOverride(f, []string{"discovery", "ec2"}, jobName, discoveryec2Args)
	return discovery.Exports{
		// This target map will utilize a RiverTokenize that results in this
		// component export rather than the standard discovery.Target RiverTokenize.
		Targets: []discovery.Target{map[string]string{"__expr__": "discovery.ec2." + jobName + ".targets"}},
	}, diags
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
	return validateHttpClientConfig(&sdConfig.HTTPClientConfig)
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
