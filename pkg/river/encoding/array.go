package encoding

import (
	"fmt"

	"github.com/grafana/agent/pkg/river/internal/value"
)

// ArrayField represents an array node.
type ArrayField struct {
	Type  string       `json:"type"`
	Value []RiverValue `json:"value,omitempty"`
}

func newArray(val value.Value) (*ArrayField, error) {
	af := &ArrayField{Type: "array"}
	return af, af.convertArray(val)
}

func (af *ArrayField) hasValue() bool {
	if af == nil {
		return false
	}
	return len(af.Value) > 0
}

func (af *ArrayField) convertArray(val value.Value) error {
	if !isArray(val) {
		return fmt.Errorf("convertArray requires a field that is an slice/array got %T", val.Interface())
	}
	af.Type = "array"
	for i := 0; i < val.Len(); i++ {
		arrEle := val.Index(i).Interface()
		arrVal := value.Encode(arrEle)

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
