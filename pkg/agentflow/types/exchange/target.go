package exchange

import "github.com/iancoleman/orderedmap"

type Target struct {
	address  string
	labels   *orderedmap.OrderedMap
	metadata *orderedmap.OrderedMap
}

func NewTarget(address string, labels *orderedmap.OrderedMap, metadata *orderedmap.OrderedMap) Target {
	return Target{
		address:  address,
		labels:   labels,
		metadata: metadata,
	}
}

func CopyTarget(in Target) Target {
	return Target{
		address:  in.Address(),
		labels:   in.Labels(),
		metadata: in.Metadata(),
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
