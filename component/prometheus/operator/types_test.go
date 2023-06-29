package operator

import (
	"testing"

	"github.com/grafana/agent/pkg/river"
	"github.com/stretchr/testify/require"
)

func TestRiverUnmarshal(t *testing.T) {
	var exampleRiverConfig = `
    forward_to = []
    namespaces = ["my-app"]
    selector {
        match_expression {
            key = "team"
            operator = "In"
            values = ["ops"]
        }
        match_labels = {
            team = "ops",
        }
    }
`

	var args Arguments
	err := river.Unmarshal([]byte(exampleRiverConfig), &args)
	require.NoError(t, err)
}

func TestEqual(t *testing.T) {
	a := Arguments{
		Namespaces: []string{"my-app"},
		Clustering: Clustering{Enabled: true},
	}
	b := Arguments{
		Namespaces: []string{"my-app"},
		Clustering: Clustering{Enabled: true},
	}
	c := Arguments{
		Namespaces: []string{"my-app", "other-app"},
		Clustering: Clustering{Enabled: false},
	}
	require.True(t, a.Equals(&b))
	require.False(t, a.Equals(&c))
}
