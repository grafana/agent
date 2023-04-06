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
`

	var args Arguments
	err := river.Unmarshal([]byte(exampleRiverConfig), &args)
	require.NoError(t, err)

	expected := Arguments{
		RefreshInterval:  5 * time.Minute,
		Port:             8181,
		HTTPClientConfig: DefaultArguments.HTTPClientConfig,
	}
	assert.Equal(t, expected, args)

	var fullerExampleRiverConfig = `
	refresh_interval = "3m"
	port = 9119
	proxy_url = "http://proxy:8080"
	follow_redirects = true
	enable_http2 = false
	basic_auth {
		username = "username"
		password = "password"
	}`
	err = river.Unmarshal([]byte(fullerExampleRiverConfig), &args)
	require.NoError(t, err)
	assert.Equal(t, 3*time.Minute, args.RefreshInterval)
	assert.Equal(t, 9119, args.Port)
	assert.Equal(t, "username", args.HTTPClientConfig.BasicAuth.Username)
	assert.Equal(t, "password", string(args.HTTPClientConfig.BasicAuth.Password))
	assert.Equal(t, "http://proxy:8080", args.HTTPClientConfig.ProxyURL.String())
	assert.Equal(t, true, args.HTTPClientConfig.FollowRedirects)
	assert.Equal(t, false, args.HTTPClientConfig.EnableHTTP2)
}

func TestBadRiverConfig(t *testing.T) {
	var exampleRiverConfig = `
	refresh_interval = "5m"
	port = 8181
	bearer_token = "token"
	bearer_token_file = "/path/to/file.token"
	`

	// Make sure the squashed HTTPClientConfig Validate function is being utilized correctly
	var args Arguments
	err := river.Unmarshal([]byte(exampleRiverConfig), &args)
	require.ErrorContains(t, err, "at most one of bearer_token & bearer_token_file must be configured")
}
