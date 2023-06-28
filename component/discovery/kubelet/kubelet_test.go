package kubelet

import (
	"testing"

	"github.com/grafana/agent/pkg/river"
	"github.com/stretchr/testify/require"
)

func TestRiverConfig(t *testing.T) {
	var exampleRiverConfig = `
	bearer_token_file = "/path/to/file.token"
`

	var args Arguments
	err := river.Unmarshal([]byte(exampleRiverConfig), &args)
	require.NoError(t, err)
}

func TestBadRiverConfig(t *testing.T) {
	var exampleRiverConfig = `
	bearer_token = "token"
	bearer_token_file = "/path/to/file.token"
`

	// Make sure the squashed HTTPClientConfig Validate function is being utilized correctly
	var args Arguments
	err := river.Unmarshal([]byte(exampleRiverConfig), &args)
	require.ErrorContains(t, err, "at most one of bearer_token & bearer_token_file must be configured")

	// Make sure that kubelet URL defaults to https://localhost:10250
	var args2 Arguments
	err = river.Unmarshal([]byte{}, &args2)
	require.NoError(t, err)
	require.Equal(t, args2.KubeletURL.String(), "https://localhost:10250")
}
