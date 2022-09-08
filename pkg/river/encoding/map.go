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

func (mf *MapField) isValid() bool {
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
		vf, arrF, mapF, err := convertRiverValue(mapVal)
		if err != nil {
			return err
		}
		if vf.isValid() {
			kf.Value = vf
		} else if arrF.isValid() {
			kf.Value = arrF
		} else if mapF.isValid() {
			kf.Value = mapF
		} else {
			return fmt.Errorf("unable to find value for %T in map", val.Interface())
		}
		fields = append(fields, kf)
	}
	mf.Value = fields
	return nil
}
