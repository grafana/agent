package controller

import (
	"time"

	"github.com/grafana/agent/component"
)

// NodeWithDependants must be implemented by BlockNode that can trigger other nodes to be evaluated.
type NodeWithDependants interface {
	BlockNode

	// LastUpdateTime returns the time corresponding to the last time where the node changed its exports.
	LastUpdateTime() time.Time

	// Exports returns the current set of exports from the managed component.
	Exports() component.Exports

	// ID returns the component ID of the managed component from its River block.
	ID() ComponentID
}
