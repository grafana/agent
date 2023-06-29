package prometheusconvert

import (
	"time"

	"github.com/grafana/agent/component/common/config"
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/discovery/azure"
	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/pkg/river/rivertypes"
	prom_config "github.com/prometheus/common/config"
	prom_azure "github.com/prometheus/prometheus/discovery/azure"
)

func appendDiscoveryAzure(pb *prometheusBlocks, label string, sdConfig *prom_azure.SDConfig) (discovery.Exports, diag.Diagnostics) {
	discoveryAzureArgs, diags := toDiscoveryAzure(sdConfig)
	block := common.NewBlockWithOverride([]string{"discovery", "azure"}, label, discoveryAzureArgs)
	pb.discoveryBlocks = append(pb.discoveryBlocks, block)
	return newDiscoverExports("discovery.azure." + label + ".targets"), diags
}

func toDiscoveryAzure(sdConfig *prom_azure.SDConfig) (*azure.Arguments, diag.Diagnostics) {
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

func validateDiscoveryAzure(sdConfig *prom_azure.SDConfig) diag.Diagnostics {
	return validateHttpClientConfig(&sdConfig.HTTPClientConfig)
}

func toManagedIdentity(sdConfig *prom_azure.SDConfig) *azure.ManagedIdentity {
	if sdConfig == nil {
		return nil
	}

	return &azure.ManagedIdentity{
		ClientID: sdConfig.ClientID,
	}
}

func toDiscoveryAzureOauth2(oAuth2 *prom_config.OAuth2, tenantId string) *azure.OAuth {
	if oAuth2 == nil {
		return nil
	}

	return &azure.OAuth{
		ClientID:     oAuth2.ClientID,
		TenantID:     tenantId,
		ClientSecret: rivertypes.Secret(oAuth2.ClientSecret),
	}
}
