package controller

import (
	"fmt"

	"github.com/grafana/agent/pkg/flow/internal/dag"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// Reference describes an HCL expression reference to a ComponentNode.
type Reference struct {
	Target *ComponentNode // Component being referenced

	// Traversal describes which field within Target is being accessed. It is
	// relative to Target and not an absolute Traversal.
	Traversal hcl.Traversal
}

// ComponentReferences returns the list of references a component is making to
// other components.
func ComponentReferences(cn *ComponentNode, g *dag.Graph) ([]Reference, hcl.Diagnostics) {
	var (
		traversals = componentTraversals(cn)

		diags hcl.Diagnostics
	)

	refs := make([]Reference, 0, len(traversals))
	for _, t := range traversals {
		ref, refDiags := resolveTraversal(t, g)
		diags = diags.Extend(refDiags)
		if refDiags.HasErrors() {
			continue
		}
		refs = append(refs, ref)
	}

	return refs, diags
}

// componentTraversals gets the set of hcl.Traverals for a given component.
func componentTraversals(cn *ComponentNode) []hcl.Traversal {
	cn.mut.RLock()
	defer cn.mut.RUnlock()
	return expressionsFromSyntaxBody(cn.block.Body.(*hclsyntax.Body))
}

// expressionsFromSyntaxBody recurses through body and finds all variable
// references.
func expressionsFromSyntaxBody(body *hclsyntax.Body) []hcl.Traversal {
	var exprs []hcl.Traversal

	for _, attrib := range body.Attributes {
		exprs = append(exprs, attrib.Expr.Variables()...)
	}
	for _, block := range body.Blocks {
		exprs = append(exprs, expressionsFromSyntaxBody(block.Body)...)
	}

	return exprs
}

func resolveTraversal(t hcl.Traversal, g *dag.Graph) (Reference, hcl.Diagnostics) {
	var (
		diags hcl.Diagnostics

		split   = t.SimpleSplit()
		partial = ComponentID{split.RootName()}
		rem     = split.Rel
	)

Lookup:
	for {
		if n := g.GetByID(partial.String()); n != nil {
			return Reference{
				Target:    n.(*ComponentNode),
				Traversal: rem,
			}, nil
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
	return Reference{}, diags
}
