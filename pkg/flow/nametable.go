package flow

import (
	"fmt"
	"strings"

	"github.com/grafana/agent/pkg/flow/dag"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

// nametable stores a set of components and allows for them to be looked up by
// path.
//
// nametable is not safe for concurrent modification.
type nametable struct {
	graph dag.Graph
}

// Add inserts the componentNode into the nametable.
func (nt *nametable) Add(cn *componentNode) {
	ref := cn.Reference()

	var lastNode dag.Node

	// Add entries into the nametable for the reference root (i.e., all parts of
	// the reference path minus the very last)
	for i := 0; i < len(ref)-1; i++ {
		ent := ntPartialReference{
			ParentPath: strings.Join(ref[:i], "."),
			Value:      ref[i],
		}

		nt.graph.Add(ent)
		if lastNode != nil {
			nt.graph.AddEdge(dag.Edge{From: lastNode, To: ent})
		}

		lastNode = ent
	}

	// Add the component itself.
	ent := ntEntry{
		ref: ntPartialReference{
			ParentPath: strings.Join(ref[:len(ref)-1], "."),
			Value:      ref[len(ref)-1],
		},
		cn: cn,
	}
	nt.graph.Add(ent)

	if lastNode != nil {
		nt.graph.AddEdge(dag.Edge{From: lastNode, To: ent})
	}
}

// LookupTraversal returns a componentNode from the graph based on the
// traversal. The entire traversal is not used; just the subset of it used to
// identify a node. Field names within that node are not known or validated.
func (nt *nametable) LookupTraversal(t hcl.Traversal) (*componentNode, hcl.Diagnostics) {
	// Try to convert t into as much as a path as possible.
	var (
		ref   reference
		diags hcl.Diagnostics

		split = t.SimpleSplit()
	)

	// The reference is as much of the hcl.Traversal that is a name lookup. The
	// resulting reference may include more than just a component (i.e., may
	// include fields).
	ref = append(ref, split.RootName())
	for _, tt := range split.Rel {
		switch tt := tt.(type) {
		case hcl.TraverseAttr:
			ref = append(ref, tt.Name)
		default:
			break
		}
	}

	// search iterates over the set of nn for name. Returns either nil (if
	// nothing is found), ntPartialReference (when a label is found), or
	// componentNode (when the component is found).
	search := func(nn []dag.Node, name string) dag.Node {
		for _, n := range nn {
			switch n := n.(type) {
			case ntPartialReference:
				if n.Value == name {
					return n
				}
			case ntEntry:
				if n.ref.Value == name {
					return n.cn
				}
			}
		}
		return nil
	}

	var lastNode dag.Node

	// Now, iterate over our reference and try to find the component by
	// traversing the tree.
	for i, frag := range ref {
		var searchSet []dag.Node
		if lastNode == nil {
			// No previous node; our search set is the set of roots.
			searchSet = nt.graph.Roots()
		} else {
			// Search set is outgoing edges of last node.
			searchSet = nt.graph.Dependencies(lastNode)
		}

		result := search(searchSet, frag)
		if comp, ok := result.(*componentNode); ok {
			// Found the component; we can stop looking.
			return comp, nil
		} else if result != nil {
			lastNode = result
		} else {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid reference",
				Detail:   fmt.Sprintf("Could not resolve partial reference %q", ref[:i+1].String()),
				Subject:  t.SourceRange().Ptr(),
			})
			return nil, diags
		}
	}

	diags = diags.Append(&hcl.Diagnostic{
		Severity: hcl.DiagError,
		Summary:  "Invalid reference",
		Detail:   fmt.Sprintf("no components start with %q", split.RootName()),
		Subject:  split.Abs.SourceRange().Ptr(),
	})
	return nil, diags
}

// BuildEvalContext builds an hcl.EvalContext from the values of the input
// nodes. Values of other nodes will not be incldued.
func (nt *nametable) BuildEvalContext(from []dag.Node) (*hcl.EvalContext, error) {
	ectx := &hcl.EvalContext{
		Variables: make(map[string]cty.Value),
	}

	ns := make(nodeSet)
	for _, n := range from {
		ns[n] = struct{}{}
	}
	if len(ns) == 0 {
		// Early return: nothing to do
		return ectx, nil
	}

	for _, root := range nt.graph.Roots() {
		varName := root.(ntPartialReference).Value

		val, err := nt.buildValue(root, ns)
		if err != nil {
			return nil, err
		}
		if val.IsKnown() {
			ectx.Variables[varName] = val
		}
	}

	if len(ectx.Variables) == 0 {
		ectx.Variables = nil
	}
	return ectx, nil
}

type nodeSet map[dag.Node]struct{}

func (nt *nametable) buildValue(ntNode dag.Node, from nodeSet) (cty.Value, error) {
	switch n := ntNode.(type) {
	case ntPartialReference:
		attrs := make(map[string]cty.Value)

		for _, n := range nt.graph.Dependencies(n) {
			val, err := nt.buildValue(n, from)
			if err != nil {
				return cty.DynamicVal, err
			}
			if !val.IsKnown() {
				continue
			}

			switch n := n.(type) {
			case ntPartialReference:
				attrs[n.Value] = val
			case ntEntry:
				attrs[n.ref.Value] = val
			}
		}

		if len(attrs) > 0 {
			return cty.ObjectVal(attrs), nil
		}
		return cty.DynamicVal, nil

	case ntEntry:
		if _, ok := from[n.cn]; !ok {
			return cty.DynamicVal, nil
		}
		return n.cn.CurrentState(), nil

	default:
		panic(fmt.Sprintf("unexpected nametable type %T", n))
	}
}

type ntPartialReference struct {
	ParentPath string
	Value      string
}

func (pr ntPartialReference) Name() string {
	if pr.ParentPath == "" {
		return pr.Value
	}
	return fmt.Sprintf("%s.%s", pr.ParentPath, pr.Value)
}

type ntEntry struct {
	ref ntPartialReference
	cn  *componentNode
}

func (ent ntEntry) Name() string {
	return fmt.Sprintf("<%s>", ent.cn.Name())
}
