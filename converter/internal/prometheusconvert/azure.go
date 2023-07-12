package prometheusconvert

import (
	"time"

	"github.com/grafana/agent/component/common/config"
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/discovery/azure"
	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/pkg/river/rivertypes"
	prom_azure "github.com/prometheus/prometheus/discovery/azure"
)

func appendDiscoveryAzure(pb *prometheusBlocks, label string, sdConfig *prom_azure.SDConfig) discovery.Exports {
	discoveryAzureArgs := toDiscoveryAzure(sdConfig)
	name := []string{"discovery", "azure"}
	block := common.NewBlockWithOverride(name, label, discoveryAzureArgs)
	pb.discoveryBlocks = append(pb.discoveryBlocks, newPrometheusBlock(block, name, label))
	return newDiscoverExports("discovery.azure." + label + ".targets")
}

func toDiscoveryAzure(sdConfig *prom_azure.SDConfig) *azure.Arguments {
	if sdConfig == nil {
		return nil
	}

	return &azure.Arguments{
		Environment:     sdConfig.Environment,
		Port:            sdConfig.Port,
		SubscriptionID:  sdConfig.SubscriptionID,
		OAuth:           toDiscoveryAzureOauth2(sdConfig.ClientID, sdConfig.TenantID, string(sdConfig.ClientSecret)),
		ManagedIdentity: toManagedIdentity(sdConfig),
		RefreshInterval: time.Duration(sdConfig.RefreshInterval),
		ResourceGroup:   sdConfig.ResourceGroup,
		ProxyURL:        config.URL(sdConfig.HTTPClientConfig.ProxyURL),
		FollowRedirects: sdConfig.HTTPClientConfig.FollowRedirects,
		EnableHTTP2:     sdConfig.HTTPClientConfig.EnableHTTP2,
		TLSConfig:       *toTLSConfig(&sdConfig.HTTPClientConfig.TLSConfig),
	}
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

func toDiscoveryAzureOauth2(clientId string, tenantId string, clientSecret string) *azure.OAuth {
	return &azure.OAuth{
		ClientID:     clientId,
		TenantID:     tenantId,
		ClientSecret: rivertypes.Secret(clientSecret),
	}
}
