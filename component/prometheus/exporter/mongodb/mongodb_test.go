package mongodb

import (
	"testing"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/pkg/integrations/mongodb_exporter"
	"github.com/grafana/agent/pkg/river"
	"github.com/stretchr/testify/require"
)

func TestRiverUnmarshal(t *testing.T) {
	riverConfig := `
	mongodb_uri = "mongodb://127.0.0.1:27017"
	`

	var args Arguments
	err := river.Unmarshal([]byte(riverConfig), &args)
	require.NoError(t, err)

	expected := Arguments{
		URI: "mongodb://127.0.0.1:27017",
	}

	require.Equal(t, expected, args)
}

func TestConvert(t *testing.T) {
	riverConfig := `
	mongodb_uri = "mongodb://127.0.0.1:27017"
	`
	var args Arguments
	err := river.Unmarshal([]byte(riverConfig), &args)
	require.NoError(t, err)

	res := args.Convert()

	expected := mongodb_exporter.Config{
		URI: "mongodb://127.0.0.1:27017",
	}
	require.Equal(t, expected, *res)
}

func TestCustomizeTarget(t *testing.T) {
	args := Arguments{
		URI: "mongodb://127.0.0.1:27017",
	}

	baseTarget := discovery.Target{}
	newTargets := customizeTarget(baseTarget, args)
	require.Equal(t, 1, len(newTargets))
	require.Equal(t, "127.0.0.1:27017", newTargets[0]["instance"])
}
