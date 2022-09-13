package encoding

import (
	"fmt"

	"github.com/grafana/agent/pkg/river/internal/value"
)

// MapField represents a value in river.
type MapField struct {
	Type  string      `json:"type,omitempty"`
	Value []*KeyField `json:"value,omitempty"`
}

func newMap(val value.Value) (*MapField, error) {
	mf := &MapField{}
	return mf, mf.convertMap(val)
}

func (mf *MapField) hasValue() bool {
	if mf == nil {
		return false
	}
	return len(mf.Value) > 0
}

func (mf *MapField) convertMap(val value.Value) error {
	mf.Type = object
	fields := make([]*KeyField, 0)
	for _, key := range val.Keys() {
		kf := &KeyField{}

		kf.Key = key
		mapVal, found := val.Key(key)
		if !found {
			return fmt.Errorf("unable to find key %s for value type %d", key, val.Type())
		}
		rv, err := convertRiverValue(mapVal)
		if err != nil {
			return err
		}
		if rv.hasValue() {
			kf.Value = rv
		} else {
			return fmt.Errorf("unable to find value for %T in map", val.Interface())
		}
		fields = append(fields, kf)
	}
	mf.Value = fields
	return nil
}
