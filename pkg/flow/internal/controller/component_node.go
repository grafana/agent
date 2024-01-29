package controller

import (
	"github.com/grafana/agent/component"
	"github.com/grafana/river/ast"
)

// ComponentNode is a generic representation of a Flow component.
type ComponentNode interface {
	RunnableNode

	// CurrentHealth returns the current health of the component.
	CurrentHealth() component.Health

	// Arguments returns the current arguments of the managed component.
	Arguments() component.Arguments

	// Exports returns the current set of exports from the managed component.
	Exports() component.Exports

	// Label returns the component label.
	Label() string

	// ComponentName returns the name of the component.
	ComponentName() string

	// ID returns the component ID of the managed component from its River block.
	ID() ComponentID

	// UpdateBlock updates the River block used to construct arguments for the managed component.
	UpdateBlock(b *ast.BlockStmt)
}
