package graphviz_test

import (
	"os/exec"
	"testing"

	"github.com/grafana/agent/pkg/flow/internal/graphviz"
	"github.com/stretchr/testify/require"
)

func TestGraphviz(t *testing.T) {
	_, err := exec.LookPath("dot")
	if err != nil {
		t.Skip("Skipping because graphviz is not installed")
	}

	testDot := `
		digraph G {
			a -> b 
			a -> c
		}
	`

	resp, err := graphviz.Dot([]byte(testDot), "dot")
	require.NoError(t, err)

	// We don't test the entire output of dot since it will do a lot of mutations
	// on even a simple graph (setting positions, sizes, etc.). However, we at least
	// make sure that it contains something we've expected, like the name of the
	// graph itself.
	require.Contains(t, string(resp), "digraph G")
}
