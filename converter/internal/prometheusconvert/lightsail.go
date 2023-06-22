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
	prom_config "github.com/prometheus/common/config"
	prom_aws "github.com/prometheus/prometheus/discovery/aws"
)

func appendDiscoveryLightsail(f *builder.File, label string, sdConfig *prom_aws.LightsailSDConfig) (discovery.Exports, diag.Diagnostics) {
	discoverylightsailArgs, diags := toDiscoveryLightsail(sdConfig)
	common.AppendBlockWithOverride(f, []string{"discovery", "lightsail"}, label, discoverylightsailArgs)
	return newDiscoverExports("discovery.lightsail." + label + ".targets"), diags
}

func toDiscoveryLightsail(sdConfig *prom_aws.LightsailSDConfig) (*aws.LightsailArguments, diag.Diagnostics) {
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

func validateDiscoveryLightsail(sdConfig *prom_aws.LightsailSDConfig) diag.Diagnostics {
	var diags diag.Diagnostics

	if sdConfig.HTTPClientConfig.BasicAuth != nil {
		diags.Add(diag.SeverityLevelWarn, "unsupported basic_auth for lightsail_sd_configs")
	}

	if sdConfig.HTTPClientConfig.Authorization != nil {
		diags.Add(diag.SeverityLevelWarn, "unsupported authorization for lightsail_sd_configs")
	}

	if sdConfig.HTTPClientConfig.OAuth2 != nil {
		diags.Add(diag.SeverityLevelWarn, "unsupported oauth2 for lightsail_sd_configs")
	}

	if !reflect.DeepEqual(prom_config.TLSConfig{}, sdConfig.HTTPClientConfig.TLSConfig) {
		diags.Add(diag.SeverityLevelWarn, "unsupported oauth2 for lightsail_sd_configs")
	}

	newDiags := validateHttpClientConfig(&sdConfig.HTTPClientConfig)

	diags = append(diags, newDiags...)
	return diags
}
