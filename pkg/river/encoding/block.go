package encoding

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/grafana/agent/pkg/river/internal/rivertags"
	"github.com/grafana/agent/pkg/river/internal/value"
)

// blockField represents the json representation of a river block.
type blockField struct {
	field           `json:",omitempty"`
	ID              string        `json:"id,omitempty"`
	Label           string        `json:"label,omitempty"`
	Body            []interface{} `json:"body,omitempty"`
	attributeFields []*attributeField
	blockFields     []*blockField
}

func newBlock(reflectValue reflect.Value, f rivertags.Field) (*blockField, error) {
	bf := &blockField{}
	return bf, bf.convertBlock(reflectValue, f)
}

// MarshalJSON implements json marshaller.
func (bf *blockField) MarshalJSON() ([]byte, error) {
	bf.Body = make([]interface{}, 0)
	for _, x := range bf.attributeFields {
		bf.Body = append(bf.Body, x)
	}
	for _, x := range bf.blockFields {
		bf.Body = append(bf.Body, x)
	}
	type temp blockField
	return json.Marshal((*temp)(bf))
}

func (bf *blockField) isValid() bool {
	if bf == nil {
		return false
	}
	return len(bf.blockFields)+len(bf.attributeFields) > 0
}

func (bf *blockField) convertBlock(reflectValue reflect.Value, f rivertags.Field) error {
	for reflectValue.Kind() == reflect.Pointer {
		if reflectValue.IsNil() {
			return nil
		}
		reflectValue = reflectValue.Elem()
	}
	if reflectValue.Kind() != reflect.Struct {
		return fmt.Errorf("convertBlock cannot work on interface or slices")
	}

	bf.Name = strings.Join(f.Name, ".")
	bf.Type = "block"

	riverFields := rivertags.Get(reflectValue.Type())
	for _, rf := range riverFields {
		fieldIn := reflectValue.FieldByIndex(rf.Index)
		fieldVal := fieldIn.Interface()

		// Blocks can only have sub blocks and attributes
		if rf.IsBlock() {
			newBF, err := newBlock(reflect.ValueOf(fieldVal), rf)
			if err != nil {
				return nil
			}
			if newBF.isValid() {
				bf.blockFields = append(bf.blockFields, newBF)
			}
		} else if rf.IsAttr() {
			newAttr, err := newAttribute(value.Encode(fieldVal), rf)
			if err != nil {
				return nil
			}
			if newAttr.isValid() {
				bf.attributeFields = append(bf.attributeFields, newAttr)
			}
		}
	}
	return nil
}
