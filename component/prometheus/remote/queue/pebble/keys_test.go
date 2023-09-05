package pebble

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKey(t *testing.T) {
	// Simple adding one element.
	ks := newMetadata()
	ks.add(1, 15, 22)
	require.Len(t, ks.keys(), 1)

	ks.removeKeys([]uint64{1})
	require.Len(t, ks.keys(), 0)

	ks.add(1, 15, 22)
	require.Len(t, ks.keys(), 1)

	ks.clear()
	require.Len(t, ks.keys(), 0)

	// Insert the keys in backwards order, this means when we check ordering
	// we ensure it actually did work.
	for i := uint64(100); i > 0; i-- {
		ks.add(i, int64(i), 1)
	}
	keys := ks.keys()
	require.Len(t, keys, 100)

	// Ensure keys are ordered ascending.
	previous := uint64(0)
	for _, x := range keys {
		if previous == 0 {
			previous = x
			continue
		}
		require.True(t, previous < x)
		previous = x
	}
	// Half the keys - 1 should be expired.
	expiredKeys := ks.keysWithExpiredTTL(50)
	require.Len(t, expiredKeys, 49)
}
