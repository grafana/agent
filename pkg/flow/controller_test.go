package flow

import (
	"testing"

	"github.com/grafana/agent/pkg/flow/component"
	"github.com/grafana/agent/pkg/flow/internal/dag"
	"github.com/grafana/agent/pkg/flow/internal/testcomponents"
	"github.com/grafana/agent/pkg/util"
	"github.com/stretchr/testify/require"
)

func TestController_LoadFile_Evaluation(t *testing.T) {
	ctrl, _ := newController(testOptions(t))

	// Use testFile from graph_builder_test.go.
	f, diags := ReadFile(t.Name(), []byte(testFile))
	require.False(t, diags.HasErrors())
	require.NotNil(t, f)

	err := ctrl.LoadFile(f)
	require.NoError(t, err)
	require.Len(t, ctrl.components, 4)

	// Check the inputs and outputs of things that should be immediately resolved
	// without having to run the components.
	in, out := getFields(t, ctrl.graph, "testcomponents.passthrough.static")
	require.Equal(t, "hello, world!", in.(testcomponents.PassthroughConfig).Input)
	require.Equal(t, "hello, world!", out.(testcomponents.PassthroughExports).Output)
}

func getFields(t *testing.T, g *dag.Graph, nodeID string) (component.Config, component.Exports) {
	t.Helper()

	n := g.GetByID(nodeID)
	require.NotNil(t, n, "couldn't find node %q in graph", nodeID)

	uc := n.(*userComponent)
	return uc.CurrentConfig(), uc.CurrentExports()
}

func testOptions(t *testing.T) Options {
	return Options{
		Logger:   util.TestLogger(t),
		DataPath: t.TempDir(),
	}
}
