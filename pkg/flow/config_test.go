package flow_test

import (
	"os"
	"strings"
	"testing"

	"github.com/grafana/agent/pkg/flow"
	"github.com/hashicorp/hcl/v2"
	"github.com/stretchr/testify/require"

	_ "github.com/grafana/agent/pkg/flow/internal/testcomponents" // Include test components
)

func TestReadFile(t *testing.T) {
	content := `
		testcomponents "tick" "ticker-a" {
			frequency = "1s"
		}

		testcomponents "passthrough" "static" {
			input = "hello, world!"
		}
	`

	f, diags := flow.ReadFile(t.Name(), []byte(content))
	require.NotNil(t, f)
	requireNoDiagErrors(t, f, diags)

	require.Len(t, f.Components, 2)
	require.Equal(t, "testcomponents.tick.ticker-a", getBlockID(f.Components[0]))
	require.Equal(t, "testcomponents.passthrough.static", getBlockID(f.Components[1]))
}

func TestReadFile_Defaults(t *testing.T) {
	f, diags := flow.ReadFile(t.Name(), []byte(``))
	require.NotNil(t, f)
	requireNoDiagErrors(t, f, diags)

	require.Len(t, f.Components, 0)
}

func TestReadFile_InvalidComponent(t *testing.T) {
	content := `
		doesnotexist "hello-world" {
		}
	`

	f, diags := flow.ReadFile(t.Name(), []byte(content))
	require.Nil(t, f)
	require.True(t, diags.HasErrors())
	require.Equal(t, `Blocks of type "doesnotexist" are not expected here.`, diags[0].Detail)
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
