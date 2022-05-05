package flow

import (
	"os"
	"testing"

	"github.com/go-kit/log"
	"github.com/grafana/agent/pkg/flow/internal/dag"
	_ "github.com/grafana/agent/pkg/flow/internal/testcomponents" // Include test components
	"github.com/hashicorp/hcl/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testFile = `
	log_level  = "debug"
	log_format = "logfmt"

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

// graphDefinition corresponding to testFile
var testGraphDefinition = graphDefinition{
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

// Test_buildGraph_New builds a new DAG without a previous graph.
func Test_buildGraph_New(t *testing.T) {
	g, diags := buildGraphFromContent(t, nil, []byte(testFile))
	require.NotNil(t, g)
	require.False(t, diags.HasErrors())
	requireGraph(t, g, testGraphDefinition)
}

// Test_buildGraph_Update builds a new DAG from a previous graph.
func Test_buildGraph_Update(t *testing.T) {
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

	prevGraph, diags := buildGraphFromContent(t, nil, []byte(startFile))
	require.NotNil(t, prevGraph)
	require.False(t, diags.HasErrors())

	g, diags := buildGraphFromContent(t, prevGraph, []byte(testFile))
	require.NotNil(t, g)
	require.False(t, diags.HasErrors())
	requireGraph(t, g, testGraphDefinition)

	// Esnure that nodes were copied over and not recreated
	require.Equal(t, prevGraph.GetByID("testcomponents.tick.ticker"), g.GetByID("testcomponents.tick.ticker"))
	require.Nil(t, g.GetByID("testcomponents.tick.remove-me")) // The new graph shouldn't have the old node
}

// Test_buildGraph_InvalidReference ensures that the graph can partially load
// even if a component has an invalid reference.
func Test_buildGraph_InvalidReference(t *testing.T) {
	testFile := `
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

	g, diags := buildGraphFromContent(t, nil, []byte(testFile))
	require.NotNil(t, g)
	require.True(t, diags.HasErrors())

	requireGraph(t, g, graphDefinition{
		Nodes: []string{
			"testcomponents.tick.ticker",
			"testcomponents.passthrough.valid",
			"testcomponents.passthrough.invalid",
		},
		OutEdges: []edge{
			{From: "testcomponents.passthrough.valid", To: "testcomponents.tick.ticker"},
		},
	})
}

func buildGraphFromContent(t *testing.T, prev *dag.Graph, bb []byte) (*dag.Graph, hcl.Diagnostics) {
	t.Helper()

	f, diags := ReadFile(t.Name(), bb)
	require.NotNil(t, f)

	if !assert.False(t, diags.HasErrors()) {
		dw := hcl.NewDiagnosticTextWriter(os.Stderr, map[string]*hcl.File{
			f.Name: f.HCL,
		}, 80, false)

		_ = dw.WriteDiagnostics(diags)
		t.FailNow()
	}

	opts := userComponentOptions{
		Logger:        log.NewNopLogger(),
		OnStateChange: func(uc *userComponent) { /* no-op */ },
	}
	return buildGraph(opts, prev, f.Components)
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
