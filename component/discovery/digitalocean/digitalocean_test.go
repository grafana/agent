package digitalocean

import (
	"testing"
	"time"

	"github.com/grafana/agent/pkg/river"
	"github.com/stretchr/testify/require"
	"gotest.tools/assert"
)

func TestRiverUnmarshal(t *testing.T) {
	var exampleRiverConfig = `
	refresh_interval = "5m"
	port = 8181
	bearer_token = "token"
`

	var args Arguments
	err := river.Unmarshal([]byte(exampleRiverConfig), &args)
	require.NoError(t, err)

	assert.Equal(t, 5*time.Minute, args.RefreshInterval)
	assert.Equal(t, 8181, args.Port)
	assert.Equal(t, "token", string(args.HTTPClientConfig.BearerToken))

	var fullerExampleRiverConfig = `
	refresh_interval = "3m"
	port = 9119
	proxy_url = "http://proxy:8080"
	follow_redirects = true
	enable_http2 = false
	bearer_token = "token"
	`
	err = river.Unmarshal([]byte(fullerExampleRiverConfig), &args)
	require.NoError(t, err)
	assert.Equal(t, 3*time.Minute, args.RefreshInterval)
	assert.Equal(t, 9119, args.Port)
	assert.Equal(t, "http://proxy:8080", args.HTTPClientConfig.ProxyURL.String())
	assert.Equal(t, true, args.HTTPClientConfig.FollowRedirects)
	assert.Equal(t, false, args.HTTPClientConfig.EnableHTTP2)
}

func TestBadRiverConfig(t *testing.T) {
	var badConfigTooManyBearerTokens = `
	refresh_interval = "5m"
	port = 8181
	bearer_token = "token"
	bearer_token_file = "/path/to/file.token"
	`

	var args Arguments
	err := river.Unmarshal([]byte(badConfigTooManyBearerTokens), &args)
	require.ErrorContains(t, err, "at most one of bearer_token & bearer_token_file must be configured")

	var badConfigIncorrectAuth = `
	refresh_interval = "5m"
	port = 8181
	basic_auth {
		username = "username"
		password = "password"
	}
	`
	var args2 Arguments
	err = river.Unmarshal([]byte(badConfigIncorrectAuth), &args2)
	require.ErrorContains(t, err, "digitalocean uses bearer tokens to authenticate with the API, bearer token or bearer token file must be specified")
}
