package controller

import (
	"context"

	"github.com/grafana/agent/component"
	"github.com/prometheus/client_golang/prometheus"
)

// ComponentInfo is an interface that encapsulates methods for managing and retrieving detailed information about a component within a BlockNode.
type ComponentInfo interface {
	BlockNode

	// CurrentHealth returns the current health of the node.
	CurrentHealth() component.Health

	// DebugInfo returns debugging information from the managed component (if any).
	DebugInfo() interface{}

	// Arguments returns the current arguments of the managed component.
	Arguments() component.Arguments

	// Exports returns the current set of exports from the managed component.
	Exports() component.Exports

	// Component returns the instance of the managed component.
	Component() component.Component

	// ModuleIDs returns the current list of modules that this component is managing.
	ModuleIDs() []string

	// Label returns the label for the block or "" if none was specified.
	Label() string

	// BlockName returns the name of the block.
	BlockName() string

	// Run runs the managed component in the calling goroutine until ctx is canceled.
	Run(ctx context.Context) error

	// ID returns the component ID of the managed component from its River block.
	ID() ComponentID

	// Registry returns the prometheus registry of the component.
	Registry() *prometheus.Registry
}
