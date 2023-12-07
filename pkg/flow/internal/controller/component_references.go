package controller

import (
	"fmt"

	"github.com/grafana/agent/pkg/flow/internal/dag"
	"github.com/grafana/river/ast"
	"github.com/grafana/river/diag"
	"github.com/grafana/river/vm"
)

// Traversal describes accessing a sequence of fields relative to a component.
// Traversal only include uninterrupted sequences of field accessors; for an
// expression "component.field_a.field_b.field_c[0].inner_field", the Traversal
// will be (field_a, field_b, field_c).
type Traversal []*ast.Ident

// Reference describes an River expression reference to a BlockNode.
type Reference struct {
	Target BlockNode // BlockNode being referenced

	// Traversal describes which nested field relative to Target is being
	// accessed.
	Traversal Traversal
}

// ComponentReferences returns the list of references a component is making to
// other components.
func ComponentReferences(cn dag.Node, g *dag.Graph) ([]Reference, diag.Diagnostics) {
	var (
		traversals []Traversal

		diags diag.Diagnostics
	)

	switch cn := cn.(type) {
	case BlockNode:
		if cn.Block() != nil {
			traversals = expressionsFromBody(cn.Block().Body)
		}
	}

	refs := make([]Reference, 0, len(traversals))
	for _, t := range traversals {
		// We use an empty scope to determine if a reference refers to something in
		// the stdlib, since vm.Scope.Lookup will search the scope tree + the
		// stdlib.
		//
		// Any call to an stdlib function is ignored.
		var emptyScope vm.Scope
		if _, ok := emptyScope.Lookup(t[0].Name); ok {
			continue
		}

		ref, resolveDiags := resolveTraversal(t, g)
		diags = append(diags, resolveDiags...)
		if resolveDiags.HasErrors() {
			continue
		}
		refs = append(refs, ref)
	}

	return refs, diags
}

// expressionsFromSyntaxBody recurses through body and finds all variable
// references.
func expressionsFromBody(body ast.Body) []Traversal {
	var w traversalWalker
	ast.Walk(&w, body)

	// Flush after the walk in case there was an in-progress traversal.
	w.flush()
	return w.traversals
}

type traversalWalker struct {
	traversals []Traversal

	buildTraversal   bool      // Whether
	currentTraversal Traversal // currentTraversal being built.
}

func (tw *traversalWalker) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.IdentifierExpr:
		// Identifiers always start new traversals. Pop the last one.
		tw.flush()
		tw.buildTraversal = true
		tw.currentTraversal = append(tw.currentTraversal, n.Ident)

	case *ast.AccessExpr:
		ast.Walk(tw, n.Value)

		// Fields being accessed should get only added to the traversal if one is
		// being built. This will be false for accesses like a().foo.
		if tw.buildTraversal {
			tw.currentTraversal = append(tw.currentTraversal, n.Name)
		}
		return nil

	case *ast.IndexExpr:
		// Indexing interrupts traversals so we flush after walking the value.
		ast.Walk(tw, n.Value)
		tw.flush()
		ast.Walk(tw, n.Index)
		return nil

	case *ast.CallExpr:
		// Calls interrupt traversals so we flush after walking the value.
		ast.Walk(tw, n.Value)
		tw.flush()
		for _, arg := range n.Args {
			ast.Walk(tw, arg)
		}
		return nil
	}

	return tw
}

// flush will flush the in-progress traversal to the traversals list and unset
// the buildTraversal state.
func (tw *traversalWalker) flush() {
	if tw.buildTraversal && len(tw.currentTraversal) > 0 {
		tw.traversals = append(tw.traversals, tw.currentTraversal)
	}
	tw.buildTraversal = false
	tw.currentTraversal = nil
}

func resolveTraversal(t Traversal, g *dag.Graph) (Reference, diag.Diagnostics) {
	var (
		diags diag.Diagnostics

		partial = ComponentID{t[0].Name}
		rem     = t[1:]
	)

	for {
		if n := g.GetByID(partial.String()); n != nil {
			return Reference{
				Target:    n.(BlockNode),
				Traversal: rem,
			}, nil
		}

		if len(rem) == 0 {
			// Stop: there's no more elements to look at in the traversal.
			break
		}

		// Append the next name in the traversal to our partial reference.
		partial = append(partial, rem[0].Name)
		rem = rem[1:]
	}

	diags = append(diags, diag.Diagnostic{
		Severity: diag.SeverityLevelError,
		Message:  fmt.Sprintf("component %q does not exist", partial),
		StartPos: ast.StartPos(t[0]).Position(),
		EndPos:   ast.StartPos(t[len(t)-1]).Position(),
	})
	return Reference{}, diags
}
