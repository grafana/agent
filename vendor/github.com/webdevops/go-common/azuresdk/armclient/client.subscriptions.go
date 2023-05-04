package armclient

import (
	"context"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions"
)

const (
	CacheIdentifierSubscriptions = "subscriptions"
)

// ListCachedSubscriptionsWithFilter return list of subscription with filter by subscription ids
func (azureClient *ArmClient) ListCachedSubscriptionsWithFilter(ctx context.Context, subscriptionFilter ...string) (map[string]*armsubscriptions.Subscription, error) {
	availableSubscriptions, err := azureClient.ListCachedSubscriptions(ctx)
	if err != nil {
		return nil, err
	}

	// filter subscriptions
	if len(subscriptionFilter) > 0 {
		tmp := map[string]*armsubscriptions.Subscription{}
		for _, subscription := range availableSubscriptions {
			for _, subscriptionID := range subscriptionFilter {
				if strings.EqualFold(subscriptionID, *subscription.SubscriptionID) {
					tmp[*subscription.SubscriptionID] = subscription
				}
			}
		}

		availableSubscriptions = tmp
	}

	return availableSubscriptions, nil
}

// ListCachedSubscriptions return cached list of Azure Subscriptions as map (key is subscription id)
func (azureClient *ArmClient) ListCachedSubscriptions(ctx context.Context) (map[string]*armsubscriptions.Subscription, error) {
	result, err := azureClient.cacheData(CacheIdentifierSubscriptions, func() (interface{}, error) {
		azureClient.logger.Debug("updating cached Azure Subscription list")
		list, err := azureClient.ListSubscriptions(ctx)
		if err != nil {
			return nil, err
		}
		azureClient.logger.Debugf("found %v Azure Subscriptions", len(list))
		return list, nil
	})
	if err != nil {
		return nil, err
	}

	return result.(map[string]*armsubscriptions.Subscription), nil
}

// ListSubscriptions return list of Azure Subscriptions as map (key is subscription id)
func (azureClient *ArmClient) ListSubscriptions(ctx context.Context) (map[string]*armsubscriptions.Subscription, error) {
	list := map[string]*armsubscriptions.Subscription{}

	client, err := armsubscriptions.NewClient(azureClient.GetCred(), azureClient.NewArmClientOptions())
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

		for _, subscription := range result.Value {
			if len(azureClient.subscriptionFilter) > 0 {
				// use subscription filter
				for _, subscriptionId := range azureClient.subscriptionFilter {
					if strings.EqualFold(*subscription.SubscriptionID, subscriptionId) {
						list[*subscription.SubscriptionID] = subscription
						break
					}
				}
			} else {
				list[*subscription.SubscriptionID] = subscription
			}
		}
	}

	// update cache
	azureClient.cache.SetDefault(CacheIdentifierSubscriptions, list)

	return list, nil
}
