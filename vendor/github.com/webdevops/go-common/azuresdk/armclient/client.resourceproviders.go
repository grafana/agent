package armclient

import (
	"context"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"

	"github.com/webdevops/go-common/utils/to"
)

const (
	CacheIdentifierResourceProviders = "resourceproviders:%s"
)

// GetResourceProvider return Azure Resource Providers by subscriptionID and providerNamespace
func (azureClient *ArmClient) GetResourceProvider(ctx context.Context, subscriptionID, providerNamespace string) (*armresources.Provider, error) {
	list, err := azureClient.ListCachedResourceProviders(ctx, subscriptionID)
	if err != nil {
		return nil, err
	}

	providerNamespace = strings.ToLower(providerNamespace)

	if provider, exists := list[providerNamespace]; exists {
		return provider, nil
	}

	return nil, nil
}

// IsResourceProviderRegistered returns if the Azure Resource Providers is registered in a subscription
func (azureClient *ArmClient) IsResourceProviderRegistered(ctx context.Context, subscriptionID, providerNamespace string) (bool, error) {
	list, err := azureClient.ListCachedResourceProviders(ctx, subscriptionID)
	if err != nil {
		return false, err
	}

	providerNamespace = strings.ToLower(providerNamespace)

	if provider, exists := list[providerNamespace]; exists {
		return provider.RegistrationState != nil && strings.EqualFold(*provider.RegistrationState, "Registered"), nil
	}

	return false, nil
}

// ListCachedResourceProviders return cached list of Azure Resource Providers as map (key is namespace)
func (azureClient *ArmClient) ListCachedResourceProviders(ctx context.Context, subscriptionID string) (map[string]*armresources.Provider, error) {
	result, err := azureClient.cacheData(fmt.Sprintf(CacheIdentifierResourceProviders, subscriptionID), func() (interface{}, error) {
		azureClient.logger.Debug("updating cached Azure ResourceProviders list")
		list, err := azureClient.ListResourceProviders(ctx, subscriptionID)
		if err != nil {
			return nil, err
		}
		azureClient.logger.WithField("subscriptionID", subscriptionID).Debugf("found %v Azure ResourceProviders", len(list))
		return list, nil
	})
	if err != nil {
		return nil, err
	}

	return result.(map[string]*armresources.Provider), nil
}

// ListResourceProviders return cached list of Azure Resource Providers as map (key is namespace)
func (azureClient *ArmClient) ListResourceProviders(ctx context.Context, subscriptionID string) (map[string]*armresources.Provider, error) {
	list := map[string]*armresources.Provider{}

	client, err := armresources.NewProvidersClient(subscriptionID, azureClient.GetCred(), azureClient.NewArmClientOptions())
	if err != nil {
		return nil, err
	}

	pager := client.NewListPager(nil)
	for pager.More() {
		result, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		if result.Value == nil {
			continue
		}

		for _, provider := range result.Value {
			list[to.StringLower(provider.Namespace)] = provider
		}
	}

	// update cache
	azureClient.cache.SetDefault(fmt.Sprintf(CacheIdentifierResourceProviders, subscriptionID), list)

	return list, nil
}
