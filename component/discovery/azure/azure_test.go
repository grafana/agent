package azure

import (
	"net/url"
	"testing"
	"time"

	"github.com/grafana/agent/component/common/config"
	"github.com/grafana/river"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/require"
	"gotest.tools/assert"
)

func TestRiverUnmarshal(t *testing.T) {
	riverCfg := `
		environment = "AzureTestCloud"
		port = 8080
		subscription_id = "subid"
		refresh_interval = "10m"
		resource_group = "test"
		oauth {
			client_id = "clientid"
			tenant_id = "tenantid"
			client_secret = "clientsecret"
		}
		enable_http2 = true
		follow_redirects = false
		proxy_url = "http://example:8080"`

	var args Arguments
	err := river.Unmarshal([]byte(riverCfg), &args)
	require.NoError(t, err)

	assert.Equal(t, "AzureTestCloud", args.Environment)
	assert.Equal(t, 8080, args.Port)
	assert.Equal(t, "subid", args.SubscriptionID)
	assert.Equal(t, 10*time.Minute, args.RefreshInterval)
	assert.Equal(t, "test", args.ResourceGroup)
	assert.Equal(t, "clientid", args.OAuth.ClientID)
	assert.Equal(t, "tenantid", args.OAuth.TenantID)
	assert.Equal(t, "clientsecret", string(args.OAuth.ClientSecret))
	assert.Equal(t, true, args.EnableHTTP2)
	assert.Equal(t, false, args.FollowRedirects)
	assert.Equal(t, "http://example:8080", args.ProxyURL.String())
}

func TestRiverUnmarshal_OAuthRequiredFields(t *testing.T) {
	riverCfg := `
		environment = "AzureTestCloud"
		port = 8080
		subscription_id = "subid"
		refresh_interval = "10m"
		resource_group = "test"
		oauth {
			client_id = "clientid"
		}`
	var args Arguments
	err := river.Unmarshal([]byte(riverCfg), &args)
	require.Error(t, err)
}

func TestValidate(t *testing.T) {
	noAuth := `
		environment = "AzureTestCloud"
		port = 8080
		subscription_id = "subid"
		refresh_interval = "10m"
		resource_group = "test"`

	var args Arguments
	err := river.Unmarshal([]byte(noAuth), &args)
	require.ErrorContains(t, err, "exactly one of oauth or managed_identity must be specified")

	bothAuth := `
		environment = "AzureTestCloud"
		port = 8080
		subscription_id = "subid"
		refresh_interval = "10m"
		resource_group = "test"
		oauth {
			client_id = "clientid"
			tenant_id = "tenantid"
			client_secret = "clientsecret"
		}
		managed_identity {
			client_id = "clientid"
		}`
	var args2 Arguments
	err = river.Unmarshal([]byte(bothAuth), &args2)
	require.ErrorContains(t, err, "exactly one of oauth or managed_identity must be specified")

	invalidTLS := `
		environment = "AzureTestCloud"
		port = 8080
		subscription_id = "subid"
		refresh_interval = "10m"
		resource_group = "test"
		managed_identity {
			client_id = "clientid"
		}
		tls_config {
			cert_file = "certfile"
			cert_pem = "certpem"
		}`
	var args3 Arguments
	err = river.Unmarshal([]byte(invalidTLS), &args3)
	require.ErrorContains(t, err, "at most one of cert_pem and cert_file must be configured")
}

func TestConvert(t *testing.T) {
	proxyUrl, _ := url.Parse("http://example:8080")
	riverArgsOAuth := Arguments{
		Environment:     "AzureTestCloud",
		Port:            8080,
		SubscriptionID:  "subid",
		RefreshInterval: 10 * time.Minute,
		ResourceGroup:   "test",
		OAuth: &OAuth{
			ClientID:     "clientid",
			TenantID:     "tenantid",
			ClientSecret: "clientsecret",
		},
		FollowRedirects: false,
		EnableHTTP2:     false,
		ProxyURL: config.URL{
			URL: proxyUrl,
		},
	}

	promArgs := riverArgsOAuth.Convert()
	assert.Equal(t, "AzureTestCloud", promArgs.Environment)
	assert.Equal(t, 8080, promArgs.Port)
	assert.Equal(t, "subid", promArgs.SubscriptionID)
	assert.Equal(t, model.Duration(10*time.Minute), promArgs.RefreshInterval)
	assert.Equal(t, "test", promArgs.ResourceGroup)
	assert.Equal(t, "clientid", promArgs.ClientID)
	assert.Equal(t, "tenantid", promArgs.TenantID)
	assert.Equal(t, "clientsecret", string(promArgs.ClientSecret))
	assert.Equal(t, false, promArgs.HTTPClientConfig.FollowRedirects)
	assert.Equal(t, false, promArgs.HTTPClientConfig.EnableHTTP2)
	assert.Equal(t, "http://example:8080", promArgs.HTTPClientConfig.ProxyURL.String())

	riverArgsManagedIdentity := Arguments{
		Environment:     "AzureTestCloud",
		Port:            8080,
		SubscriptionID:  "subid",
		RefreshInterval: 10 * time.Minute,
		ResourceGroup:   "test",
		ManagedIdentity: &ManagedIdentity{
			ClientID: "clientid",
		},
		FollowRedirects: true,
		EnableHTTP2:     true,
		ProxyURL: config.URL{
			URL: proxyUrl,
		},
	}

	promArgs = riverArgsManagedIdentity.Convert()
	assert.Equal(t, "AzureTestCloud", promArgs.Environment)
	assert.Equal(t, 8080, promArgs.Port)
	assert.Equal(t, "subid", promArgs.SubscriptionID)
	assert.Equal(t, model.Duration(10*time.Minute), promArgs.RefreshInterval)
	assert.Equal(t, "test", promArgs.ResourceGroup)
	assert.Equal(t, "clientid", promArgs.ClientID)
	assert.Equal(t, "", promArgs.TenantID)
	assert.Equal(t, "", string(promArgs.ClientSecret))
	assert.Equal(t, true, promArgs.HTTPClientConfig.FollowRedirects)
	assert.Equal(t, true, promArgs.HTTPClientConfig.EnableHTTP2)
	assert.Equal(t, "http://example:8080", promArgs.HTTPClientConfig.ProxyURL.String())
}
