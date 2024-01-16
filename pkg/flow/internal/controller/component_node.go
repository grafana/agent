package controller

import "github.com/grafana/river/ast"

// ComponentNode is a dag.Node that manages a component.
type ComponentNode interface {
	NodeWithComponent

	// UpdateBlock updates the River block used to construct arguments for the managed component.
	UpdateBlock(b *ast.BlockStmt)
}
