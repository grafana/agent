package node_exporter

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestFlags makes sure that boolean flags and some known non-boolean flags
// work as expected
func TestFlags(t *testing.T) {
	var f flags
	f.add("--path.rootfs", "/")
	require.Equal(t, []string{"--path.rootfs", "/"}, f.accepted)

	// Set up booleans to use as pointers
	var (
		truth = true

		// You know, the opposite of truth?
		falth = false
	)

	f = flags{}
	f.addBools(map[*bool]string{&truth: "collector.textfile"})
	require.Equal(t, []string{"--collector.textfile"}, f.accepted)

	f = flags{}
	f.addBools(map[*bool]string{&falth: "collector.textfile"})
	require.Equal(t, []string{"--no-collector.textfile"}, f.accepted)
}
