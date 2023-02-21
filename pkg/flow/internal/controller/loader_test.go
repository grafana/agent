package controller_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/go-kit/log"
	"github.com/grafana/agent/pkg/flow/internal/controller"
	"github.com/grafana/agent/pkg/flow/internal/dag"
	"github.com/grafana/agent/pkg/river/ast"
	"github.com/grafana/agent/pkg/river/diag"
	"github.com/grafana/agent/pkg/river/parser"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
)

func TestLoader(t *testing.T) {
	testFile := `
		testcomponents.tick "ticker" {
			frequency = "1s"
		}

		testcomponents.passthrough "static" {
			input = "hello, world!"
		}

		testcomponents.passthrough "ticker" {
			input = testcomponents.tick.ticker.tick_time
		}

		testcomponents.passthrough "forwarded" {
			input = testcomponents.passthrough.ticker.output
		}
	`

	// corresponds to testFile
	testGraphDefinition := graphDefinition{
		Nodes: []string{
			"configNode", // The config node is always present
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

	newGlobals := func() controller.ComponentGlobals {
		return controller.ComponentGlobals{
			Logger:          log.NewNopLogger(),
			TraceProvider:   trace.NewNoopTracerProvider(),
			DataPath:        t.TempDir(),
			OnExportsChange: func(cn *controller.ComponentNode) { /* no-op */ },
			Registerer:      prometheus.NewRegistry(),
		}
	}

	t.Run("New Graph", func(t *testing.T) {
		l := controller.NewLoader(newGlobals(), controller.NewControllerMetrics(prometheus.DefaultRegisterer))
		diags := applyFromContent(t, l, []byte(testFile))
		require.NoError(t, diags.ErrorOrNil())
		requireGraph(t, l.Graph(), testGraphDefinition)
	})

	t.Run("Copy existing components and delete stale ones", func(t *testing.T) {
		startFile := `
			// Component that should be copied over to the new graph
			testcomponents.tick "ticker" {
				frequency = "1s"
			}

			// Component that will not exist in the new graph
			testcomponents.tick "remove_me" {
				frequency = "1m"
			}
		`
		l := controller.NewLoader(newGlobals(), controller.NewControllerMetrics(prometheus.NewRegistry()))
		diags := applyFromContent(t, l, []byte(startFile))
		origGraph := l.Graph()
		require.NoError(t, diags.ErrorOrNil())

		diags = applyFromContent(t, l, []byte(testFile))
		require.NoError(t, diags.ErrorOrNil())
		newGraph := l.Graph()

		// Ensure that nodes were copied over and not recreated
		require.Equal(t, origGraph.GetByID("testcomponents.tick.ticker"), newGraph.GetByID("testcomponents.tick.ticker"))
		require.Nil(t, newGraph.GetByID("testcomponents.tick.remove_me")) // The new graph shouldn't have the old node
	})

	t.Run("Load with invalid components", func(t *testing.T) {
		invalidFile := `
			doesnotexist "bad_component" {
			}
		`
		l := controller.NewLoader(newGlobals(), controller.NewControllerMetrics(prometheus.NewRegistry()))
		diags := applyFromContent(t, l, []byte(invalidFile))
		require.ErrorContains(t, diags.ErrorOrNil(), `Unrecognized component name "doesnotexist`)
	})

	t.Run("Partial load with invalid reference", func(t *testing.T) {
		invalidFile := `
			testcomponents.tick "ticker" {
				frequency = "1s"
			}

			testcomponents.passthrough "valid" {
				input = testcomponents.tick.ticker.tick_time
			}

			testcomponents.passthrough "invalid" {
				input = testcomponents.tick.doesnotexist.tick_time
			}
		`
		l := controller.NewLoader(newGlobals(), controller.NewControllerMetrics(prometheus.NewRegistry()))
		diags := applyFromContent(t, l, []byte(invalidFile))
		require.Error(t, diags.ErrorOrNil())

		requireGraph(t, l.Graph(), graphDefinition{
			Nodes: []string{
				"configNode", // The config node is always present
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
			testcomponents.tick "ticker" {
				frequency = "1s"
			}

			testcomponents.passthrough "static" {
				input = testcomponents.passthrough.forwarded.output
			}

			testcomponents.passthrough "ticker" {
				input = testcomponents.passthrough.static.output
			}

			testcomponents.passthrough "forwarded" {
				input = testcomponents.passthrough.ticker.output
			}
		`
		l := controller.NewLoader(newGlobals(), controller.NewControllerMetrics(prometheus.NewRegistry()))
		diags := applyFromContent(t, l, []byte(invalidFile))
		require.Error(t, diags.ErrorOrNil())
	})

	t.Run("Handling of singleton component labels", func(t *testing.T) {
		invalidFile := `
			testcomponents.tick {
			}
			testcomponents.singleton "first" {
			}
		`
		l := controller.NewLoader(newGlobals(), controller.NewControllerMetrics(prometheus.NewRegistry()))
		diags := applyFromContent(t, l, []byte(invalidFile))
		require.ErrorContains(t, diags[0], `Component "testcomponents.tick" must have a label`)
		require.ErrorContains(t, diags[1], `Component "testcomponents.singleton" does not support labels`)
	})
}

// TestScopeWithFailingComponent is used to ensure that the scope is filled out, even if the component
// fails to properly start.
func TestScopeWithFailingComponent(t *testing.T) {
	testFile := `
		testcomponents.tick "ticker" {
			frequenc = "1s"
		}

		testcomponents.passthrough "static" {
			input = "hello, world!"
		}

		testcomponents.passthrough "ticker" {
			input = testcomponents.tick.ticker.tick_time
		}

		testcomponents.passthrough "forwarded" {
			input = testcomponents.passthrough.ticker.output
		}
	`
	newGlobals := func() controller.ComponentGlobals {
		return controller.ComponentGlobals{
			Logger:          log.NewNopLogger(),
			TraceProvider:   trace.NewNoopTracerProvider(),
			DataPath:        t.TempDir(),
			OnExportsChange: func(cn *controller.ComponentNode) { /* no-op */ },
			Registerer:      prometheus.NewRegistry(),
		}
	}

	l := controller.NewLoader(newGlobals(), controller.NewControllerMetrics(prometheus.NewRegistry()))
	diags := applyFromContent(t, l, []byte(testFile))
	require.Error(t, diags.ErrorOrNil())
	require.Len(t, diags, 1)
	require.True(t, strings.Contains(diags.Error(), "Failed to build component: decoding River: missing required attribute \"frequency\""))
}

func applyFromContent(t *testing.T, l *controller.Loader, bb []byte) diag.Diagnostics {
	t.Helper()

	var diags diag.Diagnostics

	file, err := parser.ParseFile(t.Name(), bb)

	var parseDiags diag.Diagnostics
	if errors.As(err, &parseDiags); parseDiags.HasErrors() {
		return parseDiags
	}

	var blocks []*ast.BlockStmt
	for _, stmt := range file.Body {
		switch stmt := stmt.(type) {
		case *ast.BlockStmt:
			blocks = append(blocks, stmt)
		default:
			diags = append(diags, diag.Diagnostic{
				Severity: diag.SeverityLevelError,
				Message:  "unexpected statement",
				StartPos: ast.StartPos(stmt).Position(),
				EndPos:   ast.EndPos(stmt).Position(),
			})
		}
	}
	if diags.HasErrors() {
		return diags
	}

	applyDiags, _ := l.Apply(nil, &delegateComponent{}, nil, blocks, nil)
	diags = append(diags, applyDiags...)

	return diags
}

type delegateComponent struct {
}

// ID satisfies the DelegateComponent interface
func (dc *delegateComponent) ID() string {
	return ""
}

// IDs satisfies the DelegateComponent interface
func (dc *delegateComponent) IDs() []string {
	return []string{}
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
