package remotewrite

import (
	"testing"

	"github.com/grafana/agent/pkg/river"
	"github.com/stretchr/testify/require"
)

func TestRiverConfig(t *testing.T) {
	var exampleRiverConfig = `
		external_labels = {
			cluster = "local",
		}

		endpoint {
			name           = "test-url"
			url            = "http://0.0.0.0:11111/api/v1/write"
			remote_timeout = "100ms"

			queue_config {
				batch_send_deadline = "100ms"
			}
		}
`

	var args Arguments
	err := river.Unmarshal([]byte(exampleRiverConfig), &args)
	require.NoError(t, err)
}

func TestBadRiverConfig(t *testing.T) {
	var exampleRiverConfig = `
		external_labels = {
			cluster = "local",
		}

		endpoint {
			name           = "test-url"
			url            = "http://0.0.0.0:11111/api/v1/write"
			remote_timeout = "100ms"
			bearer_token = "token"
			bearer_token_file = "/path/to/file.token"

			queue_config {
				batch_send_deadline = "100ms"
			}
		}
`

	// Make sure the squashed HTTPClientConfig Validate function is being utilized correctly
	var args Arguments
	err := river.Unmarshal([]byte(exampleRiverConfig), &args)
	require.ErrorContains(t, err, "at most one of bearer_token & bearer_token_file must be configured")
}
