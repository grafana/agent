package controller

import (
	"time"

	"github.com/grafana/agent/component"
)

// NodeWithDependants must be implemented by dag.Node that can trigger other nodes to be evaluated.
type NodeWithDependants interface {
	BlockNode

	LastUpdateTime() time.Time

	Exports() component.Exports

	ID() ComponentID
}
