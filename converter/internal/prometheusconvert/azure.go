package prometheusconvert

import (
	"time"

	"github.com/grafana/agent/component/common/config"
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/discovery/azure"
	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/pkg/river/rivertypes"
	"github.com/grafana/agent/pkg/river/token/builder"
	promconfig "github.com/prometheus/common/config"
	promazure "github.com/prometheus/prometheus/discovery/azure"
)

func appendDiscoveryAzure(f *builder.File, jobName string, sdConfig *promazure.SDConfig) (discovery.Exports, diag.Diagnostics) {
	discoveryAzureArgs, diags := toDiscoveryAzure(sdConfig)
	common.AppendBlockWithOverride(f, []string{"discovery", "azure"}, jobName, discoveryAzureArgs)
	return discovery.Exports{
		Targets: []discovery.Target{map[string]string{"discovery.azure." + jobName + ".targets": ""}},
	}, diags
}

func toDiscoveryAzure(sdConfig *promazure.SDConfig) (*azure.Arguments, diag.Diagnostics) {
	if sdConfig == nil {
		return nil, nil
	}

	return &azure.Arguments{
		Environment:     sdConfig.Environment,
		Port:            sdConfig.Port,
		SubscriptionID:  sdConfig.SubscriptionID,
		OAuth:           toDiscoveryAzureOauth2(sdConfig.HTTPClientConfig.OAuth2, sdConfig.TenantID),
		ManagedIdentity: toManagedIdentity(sdConfig),
		RefreshInterval: time.Duration(sdConfig.RefreshInterval),
		ResourceGroup:   sdConfig.ResourceGroup,
		ProxyURL:        config.URL(sdConfig.HTTPClientConfig.ProxyURL),
		FollowRedirects: sdConfig.HTTPClientConfig.FollowRedirects,
		EnableHTTP2:     sdConfig.HTTPClientConfig.EnableHTTP2,
		TLSConfig:       *toTLSConfig(&sdConfig.HTTPClientConfig.TLSConfig),
	}, validateDiscoveryAzure(sdConfig)
}

func validateDiscoveryAzure(sdConfig *promazure.SDConfig) diag.Diagnostics {
	var diags diag.Diagnostics

	if sdConfig.HTTPClientConfig.NoProxy != "" {
		diags.Add(diag.SeverityLevelWarn, "unsupported azure service discovery config no_proxy was provided")
	}

	if sdConfig.HTTPClientConfig.ProxyFromEnvironment {
		diags.Add(diag.SeverityLevelWarn, "unsupported azure service discovery config proxy_from_environment was provided")
	}

	if len(sdConfig.HTTPClientConfig.ProxyConnectHeader) > 0 {
		diags.Add(diag.SeverityLevelWarn, "unsupported azure service discovery config proxy_connect_header was provided")
	}

	return diags
}

func toManagedIdentity(sdConfig *promazure.SDConfig) *azure.ManagedIdentity {
	if sdConfig == nil {
		return nil
	}

	return &azure.ManagedIdentity{
		ClientID: sdConfig.ClientID,
	}
}

func toDiscoveryAzureOauth2(oAuth2 *promconfig.OAuth2, tenantId string) *azure.OAuth {
	if oAuth2 == nil {
		return nil
	}

	return &azure.OAuth{
		ClientID:     oAuth2.ClientID,
		TenantID:     tenantId,
		ClientSecret: rivertypes.Secret(oAuth2.ClientSecret),
	}
}
