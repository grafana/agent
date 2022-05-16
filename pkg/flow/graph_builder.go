package flow

import (
	"fmt"

	"github.com/grafana/agent/pkg/flow/internal/controller"
	"github.com/grafana/agent/pkg/flow/internal/dag"
	"github.com/hashicorp/hcl/v2"
)

// buildGraph constructs a new DAG of user components from a set of HCL blocks
// with edges between components for valid HCL references.
//
// buildGraph will only build new user components if a component with that ID
// does not exist in prev. Otherwise, the existing component is copied to the
// new graph.
//
// The resulting graph is not evaluated, transitively reduced, or checked for
// cycles.
func buildGraph(opts controller.ComponentOptions, prev *dag.Graph, blocks hcl.Blocks) (*dag.Graph, hcl.Diagnostics) {
	if prev == nil {
		prev = &dag.Graph{}
	}

	var (
		diags    hcl.Diagnostics
		newGraph dag.Graph
	)

	// Construct our list of components and start populating our cache.
	var blockMap = make(map[string]*hcl.Block)
	for _, block := range blocks {
		var uc *controller.ComponentNode

		id := controller.BlockComponentID(block).String()

		// Reject blocks if there was another block with the same fully-qualified
		// ID.
		if exist := blockMap[id]; exist != nil {
			diags = diags.Append(&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("Component %s redeclared", id),
				Detail:   fmt.Sprintf("%s: %s originally declared here", exist.DefRange.String(), id),
				Subject:  block.DefRange.Ptr(),
			})
			continue
		}
		blockMap[id] = block

		// Either use the existing component or copy over a new one.
		if n := prev.GetByID(id); n != nil {
			uc = n.(*controller.ComponentNode)
		} else if uc == nil {
			uc = controller.NewComponentNode(opts, block)
		}

		uc.UpdateBlock(block)
		newGraph.Add(uc)
	}

	// Create edges between nodes in our graph. This can only be done after all
	// the nodes exist.
	for _, node := range newGraph.Nodes() {
		var (
			uc              = node.(*controller.ComponentNode)
			refs, nodeDiags = controller.ComponentReferences(uc, &newGraph)
		)
		for _, ref := range refs {
			newGraph.AddEdge(dag.Edge{From: uc, To: ref.Target})
		}
		diags = diags.Extend(nodeDiags)
	}

	return &newGraph, diags
}
