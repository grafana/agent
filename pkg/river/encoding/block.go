package encoding

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/grafana/agent/pkg/river/internal/reflectutil"
	"github.com/grafana/agent/pkg/river/internal/rivertags"
	"github.com/grafana/agent/pkg/river/internal/value"
)

// blockField represents the json representation of a river block.
type blockField struct {
	field `json:",omitempty"`
	ID    string        `json:"id,omitempty"`
	Label string        `json:"label,omitempty"`
	Body  []interface{} `json:"body,omitempty"`
}

func newBlock(reflectValue reflect.Value, f rivertags.Field) (*blockField, error) {
	bf := &blockField{}
	return bf, bf.convertBlock(reflectValue, f)
}

func (bf *blockField) hasValue() bool {
	if bf == nil {
		return false
	}
	return len(bf.Body) > 0
}

func (bf *blockField) convertBlock(reflectValue reflect.Value, f rivertags.Field) error {
	for reflectValue.Kind() == reflect.Pointer {
		if reflectValue.IsNil() {
			return nil
		}
		reflectValue = reflectValue.Elem()
	}
	if reflectValue.Kind() != reflect.Struct {
		return fmt.Errorf("convertBlock can only be called on struct kinds, got %s", reflectValue.Kind())
	}

	bf.Name = strings.Join(f.Name, ".")
	bf.Type = "block"
	bf.Label = getBlockLabel(reflectValue)

	fields, err := getFieldsForBlock(reflectValue.Interface())
	if err != nil {
		return err
	}
	bf.Body = fields
	return nil
}

// getBlockLabel returns the label for a given block.
func getBlockLabel(rv reflect.Value) string {
	tags := rivertags.Get(rv.Type())
	for _, tag := range tags {
		if tag.Flags&rivertags.FlagLabel != 0 {
			return reflectutil.FieldWalk(rv, tag.Index, false).String()
		}
	}

	return ""
}

func getFieldsForBlock(input interface{}) ([]interface{}, error) {
	val := value.Encode(input)
	reflectVal := val.Reflect()
	rt := rivertags.Get(reflectVal.Type())
	var fields []interface{}
	for _, t := range rt {
		fieldRef := reflectutil.FieldWalk(reflectVal, t.Index, false)
		fieldVal := value.FromRaw(fieldRef)

		if t.IsBlock() && (fieldRef.Kind() == reflect.Array || fieldRef.Kind() == reflect.Slice) {
			for i := 0; i < fieldRef.Len(); i++ {
				arrEle := fieldRef.Index(i).Interface()
				bf, err := newBlock(reflect.ValueOf(arrEle), t)
				if err != nil {
					return nil, err
				}
				if bf.hasValue() {
					fields = append(fields, bf)
				}
			}
		} else if t.IsBlock() {
			bf, err := newBlock(fieldRef, t)

			if err != nil {
				return nil, err
			}
			if bf.hasValue() {
				fields = append(fields, bf)
			}
		} else {
			af, err := newAttribute(fieldVal, t)
			if err != nil {
				return nil, err
			}
			if af.hasValue() {
				fields = append(fields, af)
			}
		}
	}
	return fields, nil
}
