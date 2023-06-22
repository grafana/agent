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

func appendDiscoveryAzure(f *builder.File, label string, sdConfig *promazure.SDConfig) (discovery.Exports, diag.Diagnostics) {
	discoveryAzureArgs, diags := toDiscoveryAzure(sdConfig)
	common.AppendBlockWithOverride(f, []string{"discovery", "azure"}, label, discoveryAzureArgs)
	return newDiscoverExports("discovery.azure." + label + ".targets"), diags
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
	return validateHttpClientConfig(&sdConfig.HTTPClientConfig)
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
