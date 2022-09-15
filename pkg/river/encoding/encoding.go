package encoding

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/grafana/agent/pkg/river/token/builder"

	"github.com/grafana/agent/pkg/river/internal/rivertags"
	"github.com/grafana/agent/pkg/river/internal/value"
)

const attr = "attr"
const object = "object"

// riverField is an interface that wraps the various concrete options for a river value.
type riverField interface {
	hasValue() bool
}

// ConvertRiverBodyToJSON is used to convertBase arguments, exports, health and debuginfo.
func ConvertRiverBodyToJSON(input interface{}) (*json.RawMessage, error) {
	if input == nil {
		return nil, nil
	}
	val := value.Encode(input)
	rt := rivertags.Get(val.Reflect().Type())
	var fields []interface{}
	for _, t := range rt {
		fieldValue := val.Reflect().FieldByIndex(t.Index)
		enVal := value.Encode(fieldValue.Interface())
		// Array blocks can only happen at this level
		if t.IsBlock() && (enVal.Reflect().Kind() == reflect.Array || enVal.Reflect().Kind() == reflect.Slice) {
			for i := 0; i < enVal.Reflect().Len(); i++ {
				arrEle := enVal.Reflect().Index(i).Interface()
				bf, err := newBlock(reflect.ValueOf(arrEle), t)
				if err != nil {
					return nil, err
				}
				if bf.isValid() {
					fields = append(fields, bf)
				}
			}
		} else if t.IsBlock() {
			bf, err := newBlock(reflect.ValueOf(enVal), t)

			if err != nil {
				return nil, err
			}
			if bf.isValid() {
				fields = append(fields, bf)
			}
		} else {
			af, err := newAttribute(enVal, t)
			if err != nil {
				return nil, err
			}
			if af.isValid() {
				fields = append(fields, af)
			}
		}
	}
	bb, err := json.Marshal(fields)
	if err != nil {
		return nil, err
	}
	raw := json.RawMessage(bb)
	return &raw, nil
}

func isFieldValue(val value.Value) bool {
	switch val.Type() {
	case value.TypeNull, value.TypeNumber, value.TypeString, value.TypeBool, value.TypeFunction, value.TypeCapsule:
		return true
	}
	return false
}

// convertValue is used to transform the underlying value of a river tag to a field
func convertValue(val value.Value) (*valueField, error) {
	// Handle items that explicitly use tokenizer, these are always considered capsule values.
	if tkn, ok := val.Interface().(builder.Tokenizer); ok {
		tokens := tkn.RiverTokenize()
		return &valueField{
			Type:  "capsule",
			Value: tokens[0].Lit,
		}, nil
	}
	switch val.Type() {
	case value.TypeNull:
		return &valueField{
			Type: "null",
		}, nil
	case value.TypeNumber:
		return &valueField{
			Type:  "number",
			Value: val.Interface(),
		}, nil
	case value.TypeString:
		return &valueField{
			Type:  "string",
			Value: val.Text(),
		}, nil
	case value.TypeBool:
		return &valueField{
			Type:  "bool",
			Value: val.Bool(),
		}, nil
	case value.TypeArray:
		return nil, fmt.Errorf("convertValue does not allow array types")
	case value.TypeObject:
		return nil, fmt.Errorf("convertValue does not allow object types")
	case value.TypeFunction:
		return &valueField{
			Type:  "function",
			Value: val.Describe(),
		}, nil
	case value.TypeCapsule:
		return &valueField{
			Type:  "capsule",
			Value: val.Describe(),
		}, nil
	default:
		return nil, fmt.Errorf("unable to convert %T", val.Interface())
	}
}

func convertRiverValue(val value.Value) (riverField, error) {
	switch {
	case isFieldValue(val):
		return convertValue(val)
	case isRiverArray(val):
		return newRiverArray(val)
	case isRiverMapOrStruct(val):
		return newRiverMap(val)
	default:
		return nil, fmt.Errorf("unknown value %T", val.Interface())
	}
}

func isRiverArray(val value.Value) bool {
	return val.Type() == value.TypeArray
}

func isRiverMapOrStruct(val value.Value) bool {
	return val.Type() == value.TypeObject
}
