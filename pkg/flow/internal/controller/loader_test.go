package controller_test

import (
	"testing"

	"github.com/go-kit/log"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/pkg/flow/internal/controller"
	"github.com/grafana/agent/pkg/flow/internal/dag"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/stretchr/testify/require"
)

func TestLoader(t *testing.T) {
	testFile := `
		testcomponents "tick" "ticker" {
			frequency = "1s"
		}

		testcomponents "passthrough" "static" {
			input = "hello, world!"
		}

		testcomponents "passthrough" "ticker" {
			input = testcomponents.tick.ticker.tick_time
		}

		testcomponents "passthrough" "forwarded" {
			input = testcomponents.passthrough.ticker.output
		}
	`

	// corresponds to testFile
	testGraphDefinition := graphDefinition{
		Nodes: []string{
			"testcomponents.tick.ticker",
			"testcomponents.passthrough.static",
			"testcomponents.passthrough.ticker",
			"testcomponents.passthrough.forwarded",
		},
		OutEdges: []edge{
			{From: "testcomponents.passthrough.ticker", To: "testcomponents.tick.ticker"},
			{From: "testcomponents.passthrough.forwarded", To: "testcomponents.passthrough.ticker"},
		},
	}

	globals := controller.ComponentGlobals{
		Logger:          log.NewNopLogger(),
		DataPath:        t.TempDir(),
		OnExportsChange: func(cn *controller.ComponentNode) { /* no-op */ },
	}

	t.Run("New Graph", func(t *testing.T) {
		l := controller.NewLoader(globals)
		diags := applyFromContent(t, l, []byte(testFile))
		require.False(t, diags.HasErrors())
		requireGraph(t, l.Graph(), testGraphDefinition)
	})

	t.Run("Copy existing components and delete stale ones", func(t *testing.T) {
		startFile := `
			// Component that should be copied over to the new graph
			testcomponents "tick" "ticker" {
				frequency = "1s"
			}

			// Component that will not exist in the new graph
			testcomponents "tick" "remove-me" {
				frequency = "1m"
			}
		`
		l := controller.NewLoader(globals)
		diags := applyFromContent(t, l, []byte(startFile))
		origGraph := l.Graph()
		require.False(t, diags.HasErrors())

		diags = applyFromContent(t, l, []byte(testFile))
		require.False(t, diags.HasErrors())
		newGraph := l.Graph()

		// Ensure that nodes were copied over and not recreated
		require.Equal(t, origGraph.GetByID("testcomponents.tick.ticker"), newGraph.GetByID("testcomponents.tick.ticker"))
		require.Nil(t, newGraph.GetByID("testcomponents.tick.remove-me")) // The new graph shouldn't have the old node
	})

	t.Run("Partial load with invalid reference", func(t *testing.T) {
		invalidFile := `
			testcomponents "tick" "ticker" {
				frequency = "1s"
			}

			testcomponents "passthrough" "valid" {
				input = testcomponents.tick.ticker.tick_time
			}

			testcomponents "passthrough" "invalid" {
				input = testcomponents.tick.doesnotexist.tick_time
			}
		`
		l := controller.NewLoader(globals)
		diags := applyFromContent(t, l, []byte(invalidFile))
		require.True(t, diags.HasErrors())

		requireGraph(t, l.Graph(), graphDefinition{
			Nodes: []string{
				"testcomponents.tick.ticker",
				"testcomponents.passthrough.valid",
				"testcomponents.passthrough.invalid",
			},
			OutEdges: []edge{
				{From: "testcomponents.passthrough.valid", To: "testcomponents.tick.ticker"},
			},
		})
	})

	t.Run("File has cycles", func(t *testing.T) {
		invalidFile := `
			testcomponents "tick" "ticker" {
				frequency = "1s"
			}

			testcomponents "passthrough" "static" {
				input = testcomponents.passthrough.forwarded.output
			}

			testcomponents "passthrough" "ticker" {
				input = testcomponents.passthrough.static.output
			}

			testcomponents "passthrough" "forwarded" {
				input = testcomponents.passthrough.ticker.output
			}
		`
		l := controller.NewLoader(globals)
		diags := applyFromContent(t, l, []byte(invalidFile))
		require.True(t, diags.HasErrors())
	})
}

func applyFromContent(t *testing.T, l *controller.Loader, bb []byte) hcl.Diagnostics {
	t.Helper()

	file, diags := hclsyntax.ParseConfig(bb, t.Name(), hcl.InitialPos)
	if diags.HasErrors() {
		return diags
	}

	blockSchema := component.RegistrySchema()
	content, remainDiags := file.Body.Content(blockSchema)
	diags = diags.Extend(remainDiags)
	if diags.HasErrors() {
		return diags
	}

	applyDiags := l.Apply(nil, content.Blocks)
	diags = diags.Extend(applyDiags)

	return diags
}

type graphDefinition struct {
	Nodes    []string
	OutEdges []edge
}

type edge struct{ From, To string }

func requireGraph(t *testing.T, g *dag.Graph, expect graphDefinition) {
	t.Helper()

	var (
		actualNodes []string
		actualEdges []edge
	)

	for _, n := range g.Nodes() {
		actualNodes = append(actualNodes, n.NodeID())
	}
	require.ElementsMatch(t, expect.Nodes, actualNodes, "List of nodes do not match")

	for _, e := range g.Edges() {
		actualEdges = append(actualEdges, edge{
			From: e.From.NodeID(),
			To:   e.To.NodeID(),
		})
	}
	require.ElementsMatch(t, expect.OutEdges, actualEdges, "List of edges do not match")
}
