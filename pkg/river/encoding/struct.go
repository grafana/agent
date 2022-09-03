package encoding

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/grafana/agent/pkg/river/internal/rivertags"
	"github.com/grafana/agent/pkg/river/internal/value"
)

// StructField represents a struct river object for json conversion.
type StructField struct {
	Type  string      `json:"type"`
	Value []*KeyField `json:"value"`
}

func newStruct(val value.Value) (*StructField, error) {
	sf := &StructField{Type: object}
	return sf, sf.convertStruct(val)
}

func (sf *StructField) isValid() bool {
	if sf == nil {
		return false
	}
	return len(sf.Value) > 0
}

func (sf *StructField) convertStruct(val value.Value) error {
	if val.Reflect().Kind() != reflect.Struct {
		return fmt.Errorf("convertStruct cannot work on non-structs")
	}
	fields := make([]*KeyField, 0)
	riverFields := rivertags.Get(val.Reflect().Type())
	for _, rf := range riverFields {
		fieldValue := val.Reflect().FieldByIndex(rf.Index)
		structVal := value.Encode(fieldValue.Interface())
		kf := &KeyField{}
		kf.Key = strings.Join(rf.Name, ".")
		vf, arrF, mf, structF, err := convertRiverValue(structVal)
		if err != nil {
			return err
		}
		if vf != nil {
			kf.Value = vf
		} else if arrF != nil {
			kf.Value = arrF
		} else if mf != nil {
			kf.Value = mf
		} else if structF != nil {
			kf.Value = structF
		} else {
			return fmt.Errorf("unable to find value for %T in struct", val.Interface())
		}
		fields = append(fields, kf)
	}
	sf.Value = fields
	return nil
}
