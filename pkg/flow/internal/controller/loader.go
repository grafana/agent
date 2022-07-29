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
	"github.com/grafana/agent/pkg/river/token/builder"
	"github.com/grafana/agent/pkg/river/vm"
	"github.com/hashicorp/go-multierror"
	"github.com/prometheus/client_golang/prometheus"

	_ "github.com/grafana/agent/pkg/flow/internal/testcomponents" // Include test components
)

// The Loader builds and evaluates ComponentNodes from River blocks.
type Loader struct {
	log     log.Logger
	globals ComponentGlobals

	mut        sync.RWMutex
	graph      *dag.Graph
	components []*ComponentNode
	cache      *valueCache
	blocks     []*ast.BlockStmt // Most recently loaded blocks, used for writing
	cm         *componentMetrics
}

// NewLoader creates a new Loader. Components built by the Loader will be built
// with co for their options.
func NewLoader(globals ComponentGlobals, reg prometheus.Registerer) *Loader {
	return &Loader{
		log:     globals.Logger,
		globals: globals,

		graph: &dag.Graph{},
		cache: newValueCache(),
		cm:    newControllerMetrics(reg),
	}
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

	wireDiags := l.wireGraphEdges(&newGraph)
	diags = append(diags, wireDiags...)

	// Validate graph to detect cycles
	err := dag.Validate(&newGraph)
	if err != nil {
		diags = append(diags, multierrToDiags(err)...)
		return diags
	}

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

		// We cache both arguments and exports during an initial load in case the
		// component is new; we want to make sure that all fields are available
		// before the component updates its exports for the first time.
		if err := l.evaluate(parentScope, n.(*ComponentNode), true, true); err != nil {
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
			if _, exists := component.Get(componentName); !exists {
				diags.Add(diag.Diagnostic{
					Severity: diag.SeverityLevelError,
					Message:  fmt.Sprintf("Unrecognized component name %q", componentName),
					StartPos: block.NamePos.Position(),
					EndPos:   block.NamePos.Add(len(componentName) - 1).Position(),
				})
				continue
			}

			// Create a new component
			c = NewComponentNode(l.globals, block, l.cm)
		}

		g.Add(c)
	}

	return diags
}

func (l *Loader) wireGraphEdges(g *dag.Graph) diag.Diagnostics {
	var diags diag.Diagnostics

	for _, n := range g.Nodes() {
		refs, nodeDiags := ComponentReferences(n.(*ComponentNode), g)
		for _, ref := range refs {
			g.AddEdge(dag.Edge{From: n, To: ref.Target})
		}
		diags = append(diags, nodeDiags...)
	}

	return diags
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

// WriteBlocks returns a set of evaluated token/builder blocks for each loaded
// component. Components are returned in the order they were supplied to Apply
// (i.e., the original order from the config file) and not topological order.
//
// Blocks will include health and debug information if debugInfo is true.
func (l *Loader) WriteBlocks(debugInfo bool) []*builder.Block {
	l.mut.RLock()
	defer l.mut.RUnlock()

	blocks := make([]*builder.Block, 0, len(l.components))

	for _, b := range l.blocks {
		id := BlockComponentID(b).String()
		node, _ := l.graph.GetByID(id).(*ComponentNode)
		if node == nil {
			continue
		}

		blocks = append(blocks, WriteComponent(node, debugInfo))
	}

	return blocks
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

	// Make sure we're in-sync with the current exports of c.
	l.cache.CacheExports(c.ID(), c.Exports())

	_ = dag.WalkReverse(l.graph, []dag.Node{c}, func(n dag.Node) error {
		if n == c {
			// Skip over the starting component; the starting component passed to
			// EvaluateDependencies had its exports changed and none of its input
			// arguments will need re-evaluation.
			return nil
		}
		_ = l.evaluate(parentScope, n.(*ComponentNode), true, false)
		return nil
	})
}

// evaluate constructs the final context for c and evaluates it. mut must be
// held when calling evaluate.
func (l *Loader) evaluate(parent *vm.Scope, c *ComponentNode, cacheArgs, cacheExports bool) error {
	ectx := l.cache.BuildContext(parent)
	if err := c.Evaluate(ectx); err != nil {
		level.Error(l.log).Log("msg", "failed to evaluate component", "component", c.NodeID(), "err", err)
		return err
	}
	if cacheArgs {
		l.cache.CacheArguments(c.ID(), c.Arguments())
	}
	if cacheExports {
		l.cache.CacheExports(c.ID(), c.Exports())
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
