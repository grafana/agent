package consul

import (
	"testing"

	"github.com/grafana/agent/component/discovery"
	"github.com/stretchr/testify/require"
)

func TestCustomizeTargetValid(t *testing.T) {
	args := Arguments{
		Server: "http://localhost:8500",
	}

	baseTarget := discovery.Target{}
	newTargets := customizeTarget(baseTarget, args)
	require.Equal(t, 1, len(newTargets))
	require.Equal(t, "localhost:8500", newTargets[0]["instance"])
}
