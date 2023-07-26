package consul

import (
	"testing"

	"github.com/grafana/agent/pkg/river"
	"github.com/stretchr/testify/require"
)

func TestRiverConfig(t *testing.T) {
	var exampleRiverConfig = `
	server = "consul.example.com:8500"
	services = ["my-service"]
	token = "my-token"
	allow_stale = false
	node_meta = { foo = "bar" }
`

	var args Arguments
	err := river.Unmarshal([]byte(exampleRiverConfig), &args)
	require.NoError(t, err)
}

func TestBadRiverConfig(t *testing.T) {
	var exampleRiverConfig = `
	server = "consul.example.com:8500"
	services = ["my-service"]
	bearer_token = "token"
	bearer_token_file = "/path/to/file.token"
`

	// Make sure the squashed HTTPClientConfig Validate function is being utilized correctly
	var args Arguments
	err := river.Unmarshal([]byte(exampleRiverConfig), &args)
	require.ErrorContains(t, err, "at most one of bearer_token & bearer_token_file must be configured")
}
