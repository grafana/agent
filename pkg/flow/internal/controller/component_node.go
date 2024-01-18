package controller

import "github.com/grafana/river/ast"

// ComponentNode is a generic representation of a Flow component.
// This is an extension of the ComponentInfo interface because although
// a dag.Node might be running a component, it might not necessarily be its direct representation.
type ComponentNode interface {
	ComponentInfo

	// UpdateBlock updates the River block used to construct arguments for the managed component.
	UpdateBlock(b *ast.BlockStmt)
}
