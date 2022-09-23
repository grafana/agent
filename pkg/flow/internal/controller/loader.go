package controller

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/pkg/flow/internal/dag"
	"github.com/grafana/agent/pkg/river/ast"
	"github.com/grafana/agent/pkg/river/diag"
	"github.com/grafana/agent/pkg/river/vm"
	"github.com/hashicorp/go-multierror"

	_ "github.com/grafana/agent/pkg/flow/internal/testcomponents" // Include test components
)

// The Loader builds and evaluates ComponentNodes from River blocks.
type Loader struct {
	log     log.Logger
	globals ComponentGlobals

	mut           sync.RWMutex
	graph         *dag.Graph
	originalGraph *dag.Graph
	components    []*ComponentNode
	cache         *valueCache
	blocks        []*ast.BlockStmt // Most recently loaded blocks, used for writing
	cm            *controllerMetrics
}

// NewLoader creates a new Loader. Components built by the Loader will be built
// with co for their options.
func NewLoader(globals ComponentGlobals) *Loader {
	l := &Loader{
		log:     globals.Logger,
		globals: globals,

		graph: &dag.Graph{},
		cache: newValueCache(),
		cm:    newControllerMetrics(globals.Registerer),
	}
	cc := newControllerCollector(l)
	if globals.Registerer != nil {
		globals.Registerer.MustRegister(cc)
	}
	return l
}

// Apply loads a new set of components into the Loader. Apply will drop any
// previously loaded component which is not described in the set of River
// blocks.
//
// Apply will reuse existing components if there is an existing component which
// matches the component ID specified by any of the provided River blocks.
// Reused components will be updated to point at the new River block.
//
// Apply will perform an evaluation of all loaded components before returning.
// The provided parentContext can be used to provide global variables and
// functions to components. A child context will be constructed from the parent
// to expose values of other components.
func (l *Loader) Apply(parentScope *vm.Scope, blocks []*ast.BlockStmt) diag.Diagnostics {
	start := time.Now()
	l.mut.Lock()
	defer l.mut.Unlock()
	l.cm.controllerEvaluation.Set(1)
	defer l.cm.controllerEvaluation.Set(0)

	var (
		diags    diag.Diagnostics
		newGraph dag.Graph
	)

	populateDiags := l.populateGraph(&newGraph, blocks)
	diags = append(diags, populateDiags...)

	l.wireGraphEdges(&newGraph)

	// Validate graph to detect cycles
	err := dag.Validate(&newGraph)
	if err != nil {
		diags = append(diags, multierrToDiags(err)...)
		return diags
	}
	// Copy the original graph, this is so we can have access to the original graph for things like displaying a UI or
	// debug information.
	l.originalGraph = newGraph.Clone()
	// Perform a transitive reduction of the graph to clean it up.
	dag.Reduce(&newGraph)

	var (
		components   = make([]*ComponentNode, 0, len(blocks))
		componentIDs = make([]ComponentID, 0, len(blocks))
	)

	// Evaluate all of the components.
	_ = dag.WalkTopological(&newGraph, newGraph.Leaves(), func(n dag.Node) error {
		c := n.(*ComponentNode)

		components = append(components, c)
		componentIDs = append(componentIDs, c.ID())

		if err := l.evaluate(parentScope, n.(*ComponentNode)); err != nil {
			var evalDiags diag.Diagnostics
			if errors.As(err, &evalDiags) {
				diags = append(diags, evalDiags...)
			} else {
				diags.Add(diag.Diagnostic{
					Severity: diag.SeverityLevelError,
					Message:  fmt.Sprintf("Failed to build component: %s", err),
					StartPos: ast.StartPos(n.(*ComponentNode).block).Position(),
					EndPos:   ast.EndPos(n.(*ComponentNode).block).Position(),
				})
			}
		}
		return nil
	})

	l.components = components
	l.graph = &newGraph
	l.cache.SyncIDs(componentIDs)
	l.blocks = blocks
	l.cm.componentEvaluationTime.Observe(time.Since(start).Seconds())
	return diags
}

func (l *Loader) populateGraph(g *dag.Graph, blocks []*ast.BlockStmt) diag.Diagnostics {
	// Fill our graph with components.
	var (
		diags    diag.Diagnostics
		blockMap = make(map[string]*ast.BlockStmt, len(blocks))
	)
	for _, block := range blocks {
		var c *ComponentNode
		id := BlockComponentID(block).String()

		if orig, redefined := blockMap[id]; redefined {
			diags.Add(diag.Diagnostic{
				Severity: diag.SeverityLevelError,
				Message:  fmt.Sprintf("Component %s already declared at %s", id, ast.StartPos(orig).Position()),
				StartPos: block.NamePos.Position(),
				EndPos:   block.NamePos.Add(len(id) - 1).Position(),
			})
			continue
		}
		blockMap[id] = block

		if exist := l.graph.GetByID(id); exist != nil {
			// Re-use the existing component and update its block
			c = exist.(*ComponentNode)
			c.UpdateBlock(block)
		} else {
			componentName := strings.Join(block.Name, ".")
			registration, exists := component.Get(componentName)
			if !exists {
				diags.Add(diag.Diagnostic{
					Severity: diag.SeverityLevelError,
					Message:  fmt.Sprintf("Unrecognized component name %q", componentName),
					StartPos: block.NamePos.Position(),
					EndPos:   block.NamePos.Add(len(componentName) - 1).Position(),
				})
				continue
			}

			if registration.Singleton && block.Label != "" {
				diags.Add(diag.Diagnostic{
					Severity: diag.SeverityLevelError,
					Message:  fmt.Sprintf("Component %q does not support labels", componentName),
					StartPos: block.LabelPos.Position(),
					EndPos:   block.LabelPos.Add(len(block.Label) + 1).Position(),
				})
				continue
			}

			if !registration.Singleton && block.Label == "" {
				diags.Add(diag.Diagnostic{
					Severity: diag.SeverityLevelError,
					Message:  fmt.Sprintf("Component %q must have a label", componentName),
					StartPos: block.NamePos.Position(),
					EndPos:   block.NamePos.Add(len(componentName) - 1).Position(),
				})
				continue
			}

			// Create a new component
			c = NewComponentNode(l.globals, block)
		}

		g.Add(c)
	}

	return diags
}

func (l *Loader) wireGraphEdges(g *dag.Graph) {
	for _, n := range g.Nodes() {
		refs := ComponentReferences(n.(*ComponentNode), g)
		for _, ref := range refs {
			g.AddEdge(dag.Edge{From: n, To: ref.Target})
		}
	}
}

// Variables returns the Variables the Loader exposes for other Flow components
// to reference.
func (l *Loader) Variables() map[string]interface{} {
	return l.cache.BuildContext(nil).Variables
}

// Components returns the current set of loaded components.
func (l *Loader) Components() []*ComponentNode {
	l.mut.RLock()
	defer l.mut.RUnlock()
	return l.components
}

// Graph returns a copy of the DAG managed by the Loader.
func (l *Loader) Graph() *dag.Graph {
	l.mut.RLock()
	defer l.mut.RUnlock()
	return l.graph.Clone()
}

// OriginalGraph returns a copy of the graph before Reduce was called. This can be used if you want to show a UI of the
// original graph before the reduce function was called.
func (l *Loader) OriginalGraph() *dag.Graph {
	l.mut.RLock()
	defer l.mut.RUnlock()
	return l.originalGraph.Clone()
}

// EvaluateDependencies re-evaluates components which depend directly or
// indirectly on c. EvaluateDependencies should be called whenever a component
// updates its exports.
//
// The provided parentContext can be used to provide global variables and
// functions to components. A child context will be constructed from the parent
// to expose values of other components.
func (l *Loader) EvaluateDependencies(parentScope *vm.Scope, c *ComponentNode) {
	l.mut.RLock()
	defer l.mut.RUnlock()

	l.cm.controllerEvaluation.Set(1)
	defer l.cm.controllerEvaluation.Set(0)
	start := time.Now()

	// Make sure we're in-sync with the current exports of c.
	l.cache.CacheExports(c.ID(), c.Exports())

	_ = dag.WalkReverse(l.graph, []dag.Node{c}, func(n dag.Node) error {
		if n == c {
			// Skip over the starting component; the starting component passed to
			// EvaluateDependencies had its exports changed and none of its input
			// arguments will need re-evaluation.
			return nil
		}
		_ = l.evaluate(parentScope, n.(*ComponentNode))
		return nil
	})

	l.cm.componentEvaluationTime.Observe(time.Since(start).Seconds())
}

// evaluate constructs the final context for c and evaluates it. mut must be
// held when calling evaluate.
func (l *Loader) evaluate(parent *vm.Scope, c *ComponentNode) error {
	ectx := l.cache.BuildContext(parent)
	err := c.Evaluate(ectx)
	// Always update the cache both the arguments and exports, since both might
	// change when a component gets re-evaluated. We also want to cache the arguments and exports in case of an error
	l.cache.CacheArguments(c.ID(), c.Arguments())
	l.cache.CacheExports(c.ID(), c.Exports())
	if err != nil {
		level.Error(l.log).Log("msg", "failed to evaluate component", "component", c.NodeID(), "err", err)
		return err
	}
	return nil
}

func multierrToDiags(errors error) diag.Diagnostics {
	var diags diag.Diagnostics
	for _, err := range errors.(*multierror.Error).Errors {
		// TODO(rfratto): should this include position information?
		diags.Add(diag.Diagnostic{
			Severity: diag.SeverityLevelError,
			Message:  err.Error(),
		})
	}
	return diags
}
