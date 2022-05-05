package flow

import (
	"fmt"

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
func buildGraph(opts userComponentOptions, prev *dag.Graph, blocks hcl.Blocks) (*dag.Graph, hcl.Diagnostics) {
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
		var uc *userComponent

		id := blockToComponentName(block).String()

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
			uc = n.(*userComponent)
		} else if uc == nil {
			uc = newUserComponent(opts, block)
		}

		uc.SetBlock(block)
		newGraph.Add(uc)
	}

	// Create edges between nodes in our graph. This can only be done after all
	// the nodes exist.
	for _, node := range newGraph.Nodes() {
		var (
			uc         = node.(*userComponent)
			traversals = uc.Traversals()

			nodeDiags hcl.Diagnostics // Diagnostics specific to this node
		)
		for _, t := range traversals {
			target, lookupDiags := resolveTraversal(t, &newGraph)
			nodeDiags = nodeDiags.Extend(lookupDiags)
			if target == nil {
				continue
			}
			newGraph.AddEdge(dag.Edge{From: uc, To: target})
		}
		if nodeDiags.HasErrors() {
			setComponentError(uc, fmt.Errorf("failed resolving node references: %s", nodeDiags.Error()))
		}

		diags = diags.Extend(nodeDiags)
	}

	return &newGraph, diags
}

// resolveTraversal tries to map a traversal to a user component in the graph.
// The traversal will be incrementally searched until a node is found.
func resolveTraversal(t hcl.Traversal, g *dag.Graph) (*userComponent, hcl.Diagnostics) {
	var (
		diags hcl.Diagnostics

		split   = t.SimpleSplit()
		partial = userComponentName{split.RootName()}
		rem     = split.Rel
	)

Lookup:
	for {
		if n := g.GetByID(partial.String()); n != nil {
			return n.(*userComponent), nil
		}

		if len(rem) == 0 {
			// Stop: there's no more elements to look at in the traversal.
			break
		}

		// Find the next name in the traversal and append it to our reference.
		switch v := rem[0].(type) {
		case hcl.TraverseAttr:
			partial = append(partial, v.Name)
			// Shift rem forward one
			rem = rem[1:]
		default:
			break Lookup
		}
	}

	diags = diags.Append(&hcl.Diagnostic{
		Severity: hcl.DiagError,
		Summary:  "Invalid reference",
		Detail:   fmt.Sprintf("component %s does not exist", partial),
		Subject:  split.Abs.SourceRange().Ptr(),
	})
	return nil, diags
}
