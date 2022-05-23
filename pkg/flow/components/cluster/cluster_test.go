package cluster

import (
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/pkg/flow"
	"github.com/grafana/agent/pkg/flow/internal/controller"
	"github.com/grafana/agent/pkg/flow/internal/dag"
	"github.com/grafana/agent/pkg/util"
	"github.com/hashicorp/hcl/v2"
	"github.com/stretchr/testify/require"
	"os"
	"strings"
	"testing"
)

func TestReadFile(t *testing.T) {
	content := `
		cluster {
			self = "localhost"
			peers = ["p1"]
		}
		
		targetout "t1" {
			name = "t1"
			input = cluster.output
		}

		targetout "t2" {
			name = "t2"
			input = cluster.output
		}
	`

	f, diags := flow.ReadFile(t.Name(), []byte(content))
	require.NotNil(t, f)
	requireNoDiagErrors(t, f, diags)

	require.Len(t, f.Components, 3)
	require.Equal(t, "cluster", getBlockID(f.Components[0]))
	require.Equal(t, "targetout.t1", getBlockID(f.Components[1]))
	require.Equal(t, "targetout.t2", getBlockID(f.Components[2]))
}

func TestController_LoadFile_Evaluation(t *testing.T) {
	content := `
		cluster {
			self = "localhost"
			peers = ["p1"]
		}
		
		targetout "t1" {
			name = "t1"
			input = cluster.output
		}

		targetout "t2" {
			name = "t2"
			input = cluster.output
		}
	`
	f, diags := flow.ReadFile(t.Name(), []byte(content))
	require.NotNil(t, f)

	require.False(t, diags.HasErrors())
	require.NotNil(t, f)
	ctrl, _ := flow.NewFlow(testOptions(t))
	err := ctrl.LoadFile(f)
	require.NoError(t, err)
}

func getFields(t *testing.T, g *dag.Graph, nodeID string) (component.Arguments, component.Exports) {
	t.Helper()

	n := g.GetByID(nodeID)
	require.NotNil(t, n, "couldn't find node %q in graph", nodeID)

	uc := n.(*controller.ComponentNode)
	return uc.Arguments(), uc.Exports()
}

func requireNoDiagErrors(t *testing.T, f *flow.File, diags hcl.Diagnostics) {
	t.Helper()

	dw := hcl.NewDiagnosticTextWriter(os.Stderr, map[string]*hcl.File{
		f.Name: f.HCL,
	}, 80, false)

	_ = dw.WriteDiagnostics(diags)

	require.False(t, diags.HasErrors())
}

func getBlockID(b *hcl.Block) string {
	var parts []string
	parts = append(parts, b.Type)
	parts = append(parts, b.Labels...)
	return strings.Join(parts, ".")
}

func testOptions(t *testing.T) flow.Options {
	return flow.Options{
		Logger:   util.TestLogger(t),
		DataPath: t.TempDir(),
	}
}
