package controller

import (
	"fmt"
	"sync"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/pkg/flow/internal/dag"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"

	_ "github.com/grafana/agent/pkg/flow/internal/testcomponents" // Include test components
)

// The Loader builds and evaluates ComponentNodes from HCL blocks.
type Loader struct {
	log     log.Logger
	globals ComponentGlobals

	mut        sync.RWMutex
	graph      *dag.Graph
	components []*ComponentNode
	cache      *valueCache
	blocks     hcl.Blocks // Most recently loaded blocks, used for writing
}

// NewLoader creates a new Loader. Components built by the Loader will be built
// with co for their options.
func NewLoader(globals ComponentGlobals) *Loader {
	return &Loader{
		log:     globals.Logger,
		globals: globals,

		graph: &dag.Graph{},
		cache: newValueCache(),
	}
}

// Apply loads a new set of components into the Loader. Apply will drop any
// previously loaded component which is not described in the set of HCL blocks.
//
// Apply will reuse existing components if there is an existing component which
// matches the component ID specified by any of the provided HCL blocks. Reused
// components will be updated to point at the new HCL block.
//
// Apply will perform an evaluation of all loaded components before returning.
// The provided parentContext can be used to provide global variables and
// functions to components. A child context will be constructed from the parent
// to expose values of other components.
func (l *Loader) Apply(parentContext *hcl.EvalContext, blocks hcl.Blocks) hcl.Diagnostics {
	l.mut.Lock()
	defer l.mut.Unlock()

	var (
		diags    hcl.Diagnostics
		newGraph dag.Graph
	)

	populateDiags := l.populateGraph(&newGraph, blocks)
	diags = diags.Extend(populateDiags)

	wireDiags := l.wireGraphEdges(&newGraph)
	diags = diags.Extend(wireDiags)

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
		l.evaluate(parentContext, n.(*ComponentNode), true, true)
		return nil
	})

	l.components = components
	l.graph = &newGraph
	l.cache.SyncIDs(componentIDs)
	l.blocks = blocks
	return diags
}

func (l *Loader) populateGraph(g *dag.Graph, blocks hcl.Blocks) hcl.Diagnostics {
	// Fill our graph with components.
	var (
		diags    hcl.Diagnostics
		blockMap = make(map[string]*hcl.Block, len(blocks))
	)
	for _, block := range blocks {
		var c *ComponentNode
		id := BlockComponentID(block).String()

		if orig, redefined := blockMap[id]; redefined {
			diags = diags.Append(&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("Component %s redeclared", id),
				Detail:   fmt.Sprintf("%s: %s originally declared here", orig.DefRange.String(), id),
				Subject:  block.DefRange.Ptr(),
			})
			continue
		}
		blockMap[id] = block

		if exist := l.graph.GetByID(id); exist != nil {
			// Re-use the existing component and update its block
			c = exist.(*ComponentNode)
			c.UpdateBlock(block)
		} else {
			// Create a new component
			c = NewComponentNode(l.globals, block)
		}

		g.Add(c)
	}

	return diags
}

func (l *Loader) wireGraphEdges(g *dag.Graph) hcl.Diagnostics {
	var diags hcl.Diagnostics

	for _, n := range g.Nodes() {
		refs, nodeDiags := ComponentReferences(n.(*ComponentNode), g)
		for _, ref := range refs {
			g.AddEdge(dag.Edge{From: n, To: ref.Target})
		}
		diags = diags.Extend(nodeDiags)
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

// WriteBlocks returns a set of evaluated hclwrite blocks for each loaded
// component. Components are returned in the order they were supplied to
// Apply (i.e., the original order from the config file) and not topological
// order.
//
// Blocks will include health and debug information if debugInfo is true.
func (l *Loader) WriteBlocks(debugInfo bool) []*hclwrite.Block {
	l.mut.RLock()
	defer l.mut.RUnlock()

	blocks := make([]*hclwrite.Block, 0, len(l.components))

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

// Reevaluate reevaluates the arguments for c and any component which directly
// or indirectly depends on c.
//
// The provided parentContext can be used to provide global variables and
// functions to components. A child context will be constructed from the parent
// to expose values of other components.
func (l *Loader) Reevaluate(parentContext *hcl.EvalContext, c *ComponentNode) {
	l.mut.RLock()
	defer l.mut.RUnlock()

	// Make sure we're in-sync with the current exports of c.
	l.cache.CacheExports(c.ID(), c.Exports())

	_ = dag.WalkReverse(l.graph, []dag.Node{c}, func(n dag.Node) error {
		l.evaluate(parentContext, n.(*ComponentNode), true, false)
		return nil
	})
}

// evaluate constructs the final context for c and evalutes it. mut must be
// held when calling evaluate.
func (l *Loader) evaluate(parent *hcl.EvalContext, c *ComponentNode, cacheArgs, cacheExports bool) {
	ectx := l.cache.BuildContext(parent)
	if err := c.Evaluate(ectx); err != nil {
		level.Error(l.log).Log("msg", "failed to evaluate component", "component", c.NodeID(), "err", err)
		return
	}
	if cacheArgs {
		l.cache.CacheArguments(c.ID(), c.Arguments())
	}
	if cacheExports {
		l.cache.CacheExports(c.ID(), c.Exports())
	}
}
