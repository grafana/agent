package controller_test

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/pkg/flow/internal/controller"
	"github.com/grafana/agent/pkg/flow/internal/dag"
	"github.com/grafana/agent/pkg/flow/logging"
	"github.com/grafana/river/ast"
	"github.com/grafana/river/diag"
	"github.com/grafana/river/parser"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"

	_ "github.com/grafana/agent/pkg/flow/internal/testcomponents" // Include test components
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

	testConfig := `
		logging {
			level = "debug"
			format = "logfmt"
		}

		tracing {
			sampling_fraction = 1
		}
	`

	// corresponds to testFile
	testGraphDefinition := graphDefinition{
		Nodes: []string{
			"testcomponents.tick.ticker",
			"testcomponents.passthrough.static",
			"testcomponents.passthrough.ticker",
			"testcomponents.passthrough.forwarded",
			"logging",
			"tracing",
		},
		OutEdges: []edge{
			{From: "testcomponents.passthrough.ticker", To: "testcomponents.tick.ticker"},
			{From: "testcomponents.passthrough.forwarded", To: "testcomponents.passthrough.ticker"},
		},
	}

	t.Run("New Graph", func(t *testing.T) {
		l := setupLoader(t)
		diags := applyFromContent(t, l, []byte(testFile), []byte(testConfig), nil)
		require.NoError(t, diags.ErrorOrNil())
		requireGraph(t, l.Graph(), testGraphDefinition)
	})

	t.Run("New Graph No Config", func(t *testing.T) {
		l := setupLoader(t)
		diags := applyFromContent(t, l, []byte(testFile), nil, nil)
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
		l := setupLoader(t)
		diags := applyFromContent(t, l, []byte(startFile), []byte(testConfig), nil)
		origGraph := l.Graph()
		require.NoError(t, diags.ErrorOrNil())

		diags = applyFromContent(t, l, []byte(testFile), []byte(testConfig), nil)
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
		l := setupLoader(t)
		diags := applyFromContent(t, l, []byte(invalidFile), nil, nil)
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
		l := setupLoader(t)
		diags := applyFromContent(t, l, []byte(invalidFile), nil, nil)
		require.Error(t, diags.ErrorOrNil())

		requireGraph(t, l.Graph(), graphDefinition{
			Nodes:    nil,
			OutEdges: nil,
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
		l := setupLoader(t)
		diags := applyFromContent(t, l, []byte(invalidFile), nil, nil)
		require.Error(t, diags.ErrorOrNil())
	})
}

func TestDeclareBlockSort(t *testing.T) {
	testDeclare := `
	declare "c" {
		a "default" {}
	}
	declare "a" {
		declare "b" {
			d "default" {}
		}
	}
	declare "e" {
		a "default" {}
		d "default" {}
		c "default" {}
	}
	declare "d" {}
	`

	l := setupLoader(t)
	declareBlocks, diags := fileToBlock(t, []byte(testDeclare))
	require.NoError(t, diags.ErrorOrNil())

	sortedBlocks, diags := l.SortDeclareBlocks(declareBlocks)
	require.NoError(t, diags.ErrorOrNil())

	expectedLabels := []string{"d", "a", "c", "e"}
	declareBlocksLabels := make([]string, len(declareBlocks))
	for i, block := range sortedBlocks {
		declareBlocksLabels[i] = block.Label
	}
	require.Equal(t, expectedLabels, declareBlocksLabels)

	testDeclareCircularDependencies := `
		declare "c" {
			a "default" {}
		}
		declare "e" {
			a "default" {}
			d "default" {}
			c "default" {}
		}
		declare "a" {
			declare "b" {
				d "default" {}
			}
		}
		declare "d" {
			c "default" {}
		}
		declare "f" {}
	`

	declareBlocks, diags = fileToBlock(t, []byte(testDeclareCircularDependencies))
	require.NoError(t, diags.ErrorOrNil())

	_, diags = l.SortDeclareBlocks(declareBlocks)
	require.ErrorContains(t, diags.ErrorOrNil(), `Detected a cycle in declare dependencies; cannot sort: [a c d e]`)
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

	l := setupLoader(t)
	diags := applyFromContent(t, l, []byte(testFile), nil, nil)
	require.Error(t, diags.ErrorOrNil())
	require.Len(t, diags, 1)
	require.True(t, strings.Contains(diags.Error(), `unrecognized attribute name "frequenc"`))
}

type testCase struct {
	name                string
	testFile            string
	testDeclare         string
	expectedGraph       graphDefinition
	expectedExportValue int
	nodeID              string
}

func TestDeclareComponent(t *testing.T) {
	testCases := []testCase{
		{
			name: "BasicComponent",
			testFile: `
				add "example" {
					a = 5
					b = 6
				}
			`,
			testDeclare: `
				declare "add" {
					argument "a" { }
					argument "b" { }
					export "sum" {
						value = argument.a.value + argument.b.value
					}
				}
			`,
			expectedGraph: graphDefinition{
				Nodes: []string{
					"tracing",
					"logging",
					"add.example.argument.a",
					"add.example.argument.b",
					"add.example",
					"add.example.export.sum",
				},
				OutEdges: []edge{
					{From: "add.example.argument.a", To: "add.example"},
					{From: "add.example.argument.b", To: "add.example"},
					{From: "add.example.export.sum", To: "add.example.argument.a"},
					{From: "add.example.export.sum", To: "add.example.argument.b"},
				},
			},
			nodeID:              "add.example.export.sum",
			expectedExportValue: 11,
		},
		{
			name: "NestedComponent",
			testFile: `
				add "example" {
					a = 5
					b = 6
				}
			`,
			testDeclare: `
			declare "add" {
				argument "a" { }
				argument "b" { }

				declare "surpriseDuChef" {
					argument "a" { }
					argument "b" { }
					export "sum" {
						value = argument.a.value + argument.b.value
					}
				}

				surpriseDuChef "theOne" {
					a = argument.a.value
					b = 1
				}

				export "sum" {
					value = surpriseDuChef.theOne.export.sum + argument.b.value
				}
			}
		`,
			expectedGraph: graphDefinition{
				Nodes: []string{
					"tracing",
					"logging",
					"add.example.argument.a",
					"add.example.argument.b",
					"add.example",
					"add.example.surpriseDuChef.theOne",
					"add.example.surpriseDuChef.theOne.export.sum",
					"add.example.surpriseDuChef.theOne.argument.a",
					"add.example.surpriseDuChef.theOne.argument.b",
					"add.example.export.sum",
				},
				OutEdges: []edge{
					{From: "add.example.argument.a", To: "add.example"},
					{From: "add.example.argument.b", To: "add.example"},
					{From: "add.example.export.sum", To: "add.example.argument.b"},
					{From: "add.example.export.sum", To: "add.example.surpriseDuChef.theOne.export.sum"},
					{From: "add.example.surpriseDuChef.theOne", To: "add.example.argument.a"},
					{From: "add.example.surpriseDuChef.theOne.argument.a", To: "add.example.surpriseDuChef.theOne"},
					{From: "add.example.surpriseDuChef.theOne.argument.b", To: "add.example.surpriseDuChef.theOne"},
					{From: "add.example.surpriseDuChef.theOne.export.sum", To: "add.example.surpriseDuChef.theOne.argument.a"},
					{From: "add.example.surpriseDuChef.theOne.export.sum", To: "add.example.surpriseDuChef.theOne.argument.b"},
				},
			},
			nodeID:              "add.example.export.sum",
			expectedExportValue: 12,
		},
		{
			name: "MultiInterconnectedInstancesWithNestedDeclare",
			testFile: `
				add "example" {
					a = 5 
					b = 6 
				}
				add "example2" {
					a = 2
					b = add.example.export.sum
				}
			`,
			testDeclare: `
			declare "add" {
				argument "a" { }
				argument "b" { }
	
				declare "surpriseDuChef" {
					argument "a" { }
					argument "b" { }
					export "sum" {
						value = argument.a.value + argument.b.value
					}
				}
	
				surpriseDuChef "theOne" {
					a = argument.a.value
					b = 1
				}
			
				export "sum" {
					value = surpriseDuChef.theOne.export.sum + argument.b.value
				}
			}
		`,
			expectedGraph: graphDefinition{
				Nodes: []string{
					"tracing",
					"logging",
					"add.example.argument.a",
					"add.example.argument.b",
					"add.example",
					"add.example.surpriseDuChef.theOne",
					"add.example.surpriseDuChef.theOne.export.sum",
					"add.example.surpriseDuChef.theOne.argument.a",
					"add.example.surpriseDuChef.theOne.argument.b",
					"add.example.export.sum",
					"add.example2.argument.a",
					"add.example2.argument.b",
					"add.example2",
					"add.example2.surpriseDuChef.theOne",
					"add.example2.surpriseDuChef.theOne.export.sum",
					"add.example2.surpriseDuChef.theOne.argument.a",
					"add.example2.surpriseDuChef.theOne.argument.b",
					"add.example2.export.sum",
				},
				OutEdges: []edge{
					{From: "add.example.argument.a", To: "add.example"},
					{From: "add.example.argument.b", To: "add.example"},
					{From: "add.example.export.sum", To: "add.example.argument.b"},
					{From: "add.example.export.sum", To: "add.example.surpriseDuChef.theOne.export.sum"},
					{From: "add.example.surpriseDuChef.theOne", To: "add.example.argument.a"},
					{From: "add.example.surpriseDuChef.theOne.argument.a", To: "add.example.surpriseDuChef.theOne"},
					{From: "add.example.surpriseDuChef.theOne.argument.b", To: "add.example.surpriseDuChef.theOne"},
					{From: "add.example.surpriseDuChef.theOne.export.sum", To: "add.example.surpriseDuChef.theOne.argument.a"},
					{From: "add.example.surpriseDuChef.theOne.export.sum", To: "add.example.surpriseDuChef.theOne.argument.b"},
					{From: "add.example2", To: "add.example.export.sum"},
					{From: "add.example2.argument.a", To: "add.example2"},
					{From: "add.example2.argument.b", To: "add.example2"},
					{From: "add.example2.export.sum", To: "add.example2.argument.b"},
					{From: "add.example2.export.sum", To: "add.example2.surpriseDuChef.theOne.export.sum"},
					{From: "add.example2.surpriseDuChef.theOne", To: "add.example2.argument.a"},
					{From: "add.example2.surpriseDuChef.theOne.argument.a", To: "add.example2.surpriseDuChef.theOne"},
					{From: "add.example2.surpriseDuChef.theOne.argument.b", To: "add.example2.surpriseDuChef.theOne"},
					{From: "add.example2.surpriseDuChef.theOne.export.sum", To: "add.example2.surpriseDuChef.theOne.argument.a"},
					{From: "add.example2.surpriseDuChef.theOne.export.sum", To: "add.example2.surpriseDuChef.theOne.argument.b"},
				},
			},
			nodeID:              "add.example2.export.sum",
			expectedExportValue: 15,
		},
		{
			name: "BasicComponentWithOptional",
			testFile: `
				add "example" {
					a = 5
				}
			`,
			testDeclare: `
				declare "add" {
					argument "a" { }
					argument "b" {
						optional = true
						default = 4
					}
					export "sum" {
						value = argument.a.value + argument.b.value
					}
				}
			`,
			expectedGraph: graphDefinition{
				Nodes: []string{
					"tracing",
					"logging",
					"add.example.argument.a",
					"add.example.argument.b",
					"add.example",
					"add.example.export.sum",
				},
				OutEdges: []edge{
					{From: "add.example.argument.a", To: "add.example"},
					{From: "add.example.argument.b", To: "add.example"},
					{From: "add.example.export.sum", To: "add.example.argument.a"},
					{From: "add.example.export.sum", To: "add.example.argument.b"},
				},
			},
			nodeID:              "add.example.export.sum",
			expectedExportValue: 9,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			l := setupLoader(t)
			diags := applyFromContent(t, l, []byte(tc.testFile), nil, []byte(tc.testDeclare))
			require.NoError(t, diags.ErrorOrNil())
			requireGraph(t, l.Graph(), tc.expectedGraph)
			require.Equal(t, l.Graph().GetByID(tc.nodeID).(*controller.ExportConfigNode).Value(), int(tc.expectedExportValue))
		})
	}
}

func setupLoader(t *testing.T) *controller.Loader {
	newLoaderOptions := func() controller.LoaderOptions {
		l, _ := logging.New(os.Stderr, logging.DefaultOptions)
		return controller.LoaderOptions{
			ComponentGlobals: controller.ComponentGlobals{
				Logger:            l,
				TraceProvider:     trace.NewNoopTracerProvider(),
				DataPath:          t.TempDir(),
				OnComponentUpdate: func(cn *controller.ComponentNode) { /* no-op */ },
				Registerer:        prometheus.NewRegistry(),
				NewModuleController: func(id string, availableServices []string) controller.ModuleController {
					return fakeModuleController{}
				},
			},
		}
	}
	return controller.NewLoader(newLoaderOptions())
}

func applyFromContent(t *testing.T, l *controller.Loader, componentBytes []byte, configBytes []byte, declareBytes []byte) diag.Diagnostics {
	t.Helper()

	var (
		diags           diag.Diagnostics
		componentBlocks []*ast.BlockStmt
		configBlocks    []*ast.BlockStmt = nil
		declareBlocks   []*ast.BlockStmt
	)

	componentBlocks, diags = fileToBlock(t, componentBytes)
	if diags.HasErrors() {
		return diags
	}

	if string(configBytes) != "" {
		configBlocks, diags = fileToBlock(t, configBytes)
		if diags.HasErrors() {
			return diags
		}
	}

	if string(declareBytes) != "" {
		declareBlocks, diags = fileToBlock(t, declareBytes)
		if diags.HasErrors() {
			return diags
		}
	}
	applyDiags := l.Apply(nil, componentBlocks, configBlocks, declareBlocks)
	diags = append(diags, applyDiags...)

	return diags
}

func fileToBlock(t *testing.T, bytes []byte) ([]*ast.BlockStmt, diag.Diagnostics) {
	var diags diag.Diagnostics
	file, err := parser.ParseFile(t.Name(), bytes)

	var parseDiags diag.Diagnostics
	if errors.As(err, &parseDiags); parseDiags.HasErrors() {
		return nil, parseDiags
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

	return blocks, diags
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

type fakeModuleController struct{}

func (f fakeModuleController) NewModule(id string, export component.ExportFunc) (component.Module, error) {
	return nil, nil
}

func (f fakeModuleController) ModuleIDs() []string {
	return nil
}

func (f fakeModuleController) ClearModuleIDs() {
}
