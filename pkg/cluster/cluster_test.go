package cluster

import (
	"testing"

	"github.com/rfratto/ckit"
	"github.com/stretchr/testify/require"
)

func TestLocalNode_Lookup(t *testing.T) {
	t.Run("replicationFactor 0 returns nothing", func(t *testing.T) {
		ln := NewLocalNode("localhost:8888")
		res, err := ln.Lookup(0, 0)
		require.NoError(t, err)
		require.Len(t, res, 0)
	})

	t.Run("replicationFactor 1 returns self", func(t *testing.T) {
		ln := NewLocalNode("localhost:8888")
		res, err := ln.Lookup(0, 1)

		require.NoError(t, err)

		expect := []ckit.Peer{{
			Name:  "local",
			Addr:  "localhost:8888",
			Self:  true,
			State: ckit.StateParticipant,
		}}
		require.Equal(t, expect, res)
	})

	t.Run("replicationFactor >1 returns error", func(t *testing.T) {
		ln := NewLocalNode("localhost:8888")
		res, err := ln.Lookup(0, 2)
		require.EqualError(t, err, "need 2 nodes; only 1 available")
		require.Nil(t, res)
	})
}

func TestLocalNode_Peers(t *testing.T) {
	t.Run("always returns self", func(t *testing.T) {
		ln := NewLocalNode("localhost:8888")

		expect := []ckit.Peer{{
			Name:  "local",
			Addr:  "localhost:8888",
			Self:  true,
			State: ckit.StateParticipant,
		}}
		require.Equal(t, expect, ln.Peers())
	})
}
