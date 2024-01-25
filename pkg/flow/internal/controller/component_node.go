package controller

import (
	"context"

	"github.com/grafana/agent/component"
)

// ComponentNode is a generic representation of a Flow component.
type ComponentNode interface {
	RunnableNode

	// CurrentHealth returns the current health of the node.
	CurrentHealth() component.Health

	// Arguments returns the current arguments of the managed component.
	Arguments() component.Arguments

	// Exports returns the current set of exports from the managed component.
	Exports() component.Exports

	// Label returns the label for the block or "" if none was specified.
	Label() string

	// ComponentName returns the name of the block.
	ComponentName() string

	// Run runs the managed component in the calling goroutine until ctx is canceled.
	Run(ctx context.Context) error

	// ID returns the component ID of the managed component from its River block.
	ID() ComponentID
}
