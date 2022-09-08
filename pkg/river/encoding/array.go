package encoding

import (
	"encoding/json"
	"fmt"

	"github.com/grafana/agent/pkg/river/internal/value"
)

// ArrayField represents an array node.
type ArrayField struct {
	Type        string        `json:"type"`
	Value       []interface{} `json:"value,omitempty"`
	valueFields []*ValueField
	arrayFields []*ArrayField
	mapFields   []*MapField
}

func newArray(val value.Value) (*ArrayField, error) {
	af := &ArrayField{Type: "array"}
	return af, af.convertArray(val)
}

// MarshalJSON implements json marshaller.
func (af *ArrayField) MarshalJSON() ([]byte, error) {
	af.Value = make([]interface{}, 0)
	for _, x := range af.valueFields {
		af.Value = append(af.Value, x)
	}
	for _, x := range af.arrayFields {
		af.Value = append(af.Value, x)
	}
	for _, x := range af.mapFields {
		af.Value = append(af.Value, x)
	}
	type temp ArrayField
	return json.Marshal((*temp)(af))
}

func (af *ArrayField) isValid() bool {
	if af == nil {
		return false
	}
	return len(af.valueFields)+len(af.mapFields)+len(af.arrayFields) > 0
}

func (af *ArrayField) convertArray(val value.Value) error {
	if !isArray(val) {
		return fmt.Errorf("convertArray requires a field that is an slice/array got %T", val.Interface())
	}
	af.Type = "array"
	values := make([]*ValueField, 0)
	arrays := make([]*ArrayField, 0)
	maps := make([]*MapField, 0)

	for i := 0; i < val.Len(); i++ {
		arrEle := val.Index(i).Interface()
		arrVal := value.Encode(arrEle)

		vf, arrF, mf, err := convertRiverValue(arrVal)
		if err != nil {
			return err
		}

		if vf.isValid() {
			values = append(values, vf)
		} else if arrF.isValid() {
			arrays = append(arrays, arrF)
		} else if mf.isValid() {
			maps = append(maps, mf)
		} else {
			return fmt.Errorf("unable to find value for %T in convertArray", val.Interface())
		}
	}
	af.valueFields = values
	af.arrayFields = arrays
	af.mapFields = maps
	return nil
}
