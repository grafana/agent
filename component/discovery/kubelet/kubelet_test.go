package kubelet

import (
	"testing"

	"github.com/grafana/agent/pkg/river"
	"github.com/stretchr/testify/require"
)

func TestRiverConfig(t *testing.T) {
	var exampleRiverConfig = `
	kubelet_url = "https://10.0.0.1:10255"
`

	var args Arguments
	err := river.Unmarshal([]byte(exampleRiverConfig), &args)
	require.NoError(t, err)
}

func TestBadRiverConfig(t *testing.T) {
	var exampleRiverConfig = `
	kubelet_url = "https://10.0.0.1:10255"
	bearer_token = "token"
	bearer_token_file = "/path/to/file.token"
`

	// Make sure the squashed HTTPClientConfig Validate function is being utilized correctly
	var args Arguments
	err := river.Unmarshal([]byte(exampleRiverConfig), &args)
	require.ErrorContains(t, err, "at most one of bearer_token & bearer_token_file must be configured")

	var missingKubeletURL = ""
	// Make sure that kubelet URL is required
	var args2 Arguments
	err = river.Unmarshal([]byte(missingKubeletURL), &args2)
	require.ErrorContains(t, err, "missing required attribute \"kubelet_url\"")
}
