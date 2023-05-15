package apache

import (
	"testing"

	"github.com/grafana/agent/component/discovery"
	"github.com/stretchr/testify/require"
)

func TestCustomizeTargetValid(t *testing.T) {
	args := Arguments{
		ApacheAddr: "http://localhost/server-status?auto",
	}

	baseTarget := discovery.Target{}
	newTargets := customizeTarget(baseTarget, args)
	require.Equal(t, 1, len(newTargets))
	require.Equal(t, "localhost", newTargets[0]["instance"])
}
