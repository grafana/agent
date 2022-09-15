package encoding

import (
	"fmt"

	"github.com/grafana/agent/pkg/river/internal/value"
)

// arrayField represents an array node.
type arrayField struct {
	Type  string       `json:"type"`
	Value []riverField `json:"value,omitempty"`
}

func newRiverArray(val value.Value) (*arrayField, error) {
	af := &arrayField{Type: "array"}
	return af, af.convertArray(val)
}

func (af *arrayField) hasValue() bool {
	if af == nil {
		return false
	}
	return len(af.Value) > 0
}

func (af *arrayField) convertArray(val value.Value) error {
	if !isRiverArray(val) {
		return fmt.Errorf("convertArray requires a field that is an slice/array got %T", val.Interface())
	}
	af.Type = "array"
	for i := 0; i < val.Len(); i++ {
		arrVal := val.Index(i)

		rv, err := convertRiverValue(arrVal)
		if err != nil {
			return err
		}
		if !rv.hasValue() {
			continue
		}
		af.Value = append(af.Value, rv)
	}
	return nil
}
