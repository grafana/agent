package controller

import (
	"time"

	"github.com/grafana/agent/component"
)

// TODO: Comments
type NodeWithDependants interface {
	BlockNode

	LastUpdateTime() time.Time

	Exports() component.Exports

	ID() ComponentID
}
