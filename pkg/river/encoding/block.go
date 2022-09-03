package encoding

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/grafana/agent/pkg/river/internal/rivertags"
	"github.com/grafana/agent/pkg/river/internal/value"
)

// BlockField represents the json representation of a river block.
type BlockField struct {
	Field           `json:",omitempty"`
	ID              string        `json:"id,omitempty"`
	Label           string        `json:"label,omitempty"`
	Body            []interface{} `json:"body,omitempty"`
	attributeFields []*AttributeField
	blockFields     []*BlockField
}

func newBlock(val value.Value, f rivertags.Field) (*BlockField, error) {
	bf := &BlockField{}
	return bf, bf.convertBlock(val, f)
}

// MarshalJSON implements json marshaller.
func (bf *BlockField) MarshalJSON() ([]byte, error) {
	bf.Body = make([]interface{}, 0)
	for _, x := range bf.attributeFields {
		bf.Body = append(bf.Body, x)
	}
	for _, x := range bf.blockFields {
		bf.Body = append(bf.Body, x)
	}
	type temp BlockField
	return json.Marshal((*temp)(bf))
}

func (bf *BlockField) isValid() bool {
	if bf == nil {
		return false
	}
	return len(bf.blockFields)+len(bf.attributeFields) > 0
}

func (bf *BlockField) convertBlock(val value.Value, f rivertags.Field) error {
	if val.Reflect().Kind() != reflect.Struct {
		return fmt.Errorf("convertBlock cannot work on interface or slices")
	}
	if val.Interface() == nil {
		return nil
	}
	if val.Reflect().IsZero() {
		return nil
	}
	bf.Name = strings.Join(f.Name, ".")
	bf.Type = "block"

	riverFields := rivertags.Get(val.Reflect().Type())
	for _, rf := range riverFields {
		fieldIn := val.Reflect().FieldByIndex(rf.Index)
		fieldVal := value.Encode(fieldIn.Interface())

		// Blocks can only have sub blocks and attributes
		if rf.IsBlock() {
			newBF, err := newBlock(fieldVal, rf)
			if err != nil {
				return nil
			}
			if newBF.isValid() {
				bf.blockFields = append(bf.blockFields, newBF)
			}
		} else if rf.IsAttr() {
			newAttr, err := newAttribute(fieldVal, rf)
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
