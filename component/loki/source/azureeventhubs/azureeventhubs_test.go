package azureeventhubs

import (
	"testing"

	"github.com/grafana/agent/pkg/river"
	"github.com/stretchr/testify/require"
)

func TestRiverConfigOAuth(t *testing.T) {
	var exampleRiverConfig = `

	fully_qualified_namespace = "my-ns.servicebus.windows.net:9093"
	event_hubs                = ["test"]
	forward_to                = []

	authentication {
		mechanism = "oauth"
	}
`

	var args Arguments
	err := river.Unmarshal([]byte(exampleRiverConfig), &args)
	require.NoError(t, err)
}

func TestRiverConfigConnectionString(t *testing.T) {
	var exampleRiverConfig = `

	fully_qualified_namespace = "my-ns.servicebus.windows.net:9093"
	event_hubs                = ["test"]
	forward_to                = []

	authentication {
		mechanism = "connection_string"
        connection_string = "my-conn-string"
	}
`

	var args Arguments
	err := river.Unmarshal([]byte(exampleRiverConfig), &args)
	require.NoError(t, err)
}
