package exchange

import (
	"github.com/iancoleman/orderedmap"
)

type TargetState int

const (
	New TargetState = iota + 1
	Deleted
	Updated
)

type Target struct {
	source   string
	address  string
	labels   *orderedmap.OrderedMap
	metadata *orderedmap.OrderedMap
	state    TargetState
}

func NewTarget(address string, source string, labels *orderedmap.OrderedMap, metadata *orderedmap.OrderedMap, state TargetState) Target {
	return Target{
		address:  address,
		source:   source,
		labels:   labels,
		metadata: metadata,
		state:    state,
	}
}

func (t *Target) Address() string {
	return t.address
}

func (t *Target) Labels() *orderedmap.OrderedMap {
	return copyMap(t.labels)
}

func (t *Target) Metadata() *orderedmap.OrderedMap {
	return copyMap(t.metadata)
}

func (t *Target) Source() string {
	return t.source
}

func (t *Target) State() TargetState {
	return t.state
}
