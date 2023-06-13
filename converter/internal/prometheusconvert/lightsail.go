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

func appendDiscoveryLightsail(f *builder.File, jobName string, sdConfig *promaws.LightsailSDConfig) (discovery.Exports, diag.Diagnostics) {
	discoverylightsailArgs, diags := toDiscoveryLightsail(sdConfig)
	common.AppendBlockWithOverride(f, []string{"discovery", "lightsail"}, jobName, discoverylightsailArgs)
	return newDiscoverExports("discovery.lightsail." + jobName + ".targets"), diags
}

func toDiscoveryLightsail(sdConfig *promaws.LightsailSDConfig) (*aws.LightsailArguments, diag.Diagnostics) {
	if sdConfig == nil {
		return nil, nil
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
	}, validateDiscoveryLightsail(sdConfig)
}

func validateDiscoveryLightsail(sdConfig *promaws.LightsailSDConfig) diag.Diagnostics {
	return validateHttpClientConfig(&sdConfig.HTTPClientConfig)
}
