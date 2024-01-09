package controller

import (
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/pkg/flow/internal/dag"
)

type ComponentNode interface {
	dag.Node

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
}
