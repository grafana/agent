package encoding

import (
	"sort"

	"github.com/grafana/agent/pkg/river/internal/value"
)

// mapField represents a value in river.
type mapField struct {
	Type  string      `json:"type,omitempty"`
	Value []*keyField `json:"value,omitempty"`
}

func newRiverMap(val value.Value) (*mapField, error) {
	mf := &mapField{}
	return mf, mf.convertMap(val)
}

func (mf *mapField) hasValue() bool {
	if mf == nil {
		return false
	}
	return len(mf.Value) > 0
}

func (mf *mapField) convertMap(val value.Value) error {
	mf.Type = objectType
	fields := make([]*keyField, 0)

	keys := val.Keys()

	// If v isn't an ordered object (i.e., a Go map), sort the keys so they have
	// a deterministic print order.
	if !val.OrderedKeys() {
		sort.Strings(keys)
	}

	for _, key := range keys {
		kf := &keyField{}

		kf.Key = key
		mapVal, found := val.Key(key)
		if !found {
			continue
		}
		rv, err := convertRiverValue(mapVal)
		if err != nil {
			return err
		}
		if rv.hasValue() {
			kf.Value = rv
		} else {
			kf.Value = &valueField{Type: "null"}
		}
		fields = append(fields, kf)
	}
	mf.Value = fields
	return nil
}
