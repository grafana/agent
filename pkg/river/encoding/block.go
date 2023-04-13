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

func newBlock(namePrefix []string, reflectValue reflect.Value, f rivertags.Field) (*blockField, error) {
	bf := &blockField{}
	return bf, bf.convertBlock(namePrefix, reflectValue, f)
}

func (bf *blockField) hasValue() bool {
	if bf == nil {
		return false
	}
	return len(bf.Body) > 0
}

func (bf *blockField) convertBlock(namePrefix []string, reflectValue reflect.Value, f rivertags.Field) error {
	for reflectValue.Kind() == reflect.Pointer {
		if reflectValue.IsNil() {
			return nil
		}
		reflectValue = reflectValue.Elem()
	}

	switch reflectValue.Kind() {
	case reflect.Struct:
		bf.Name = strings.Join(mergeStringSlices(namePrefix, f.Name), ".")
		bf.Type = "block"
		bf.Label = getBlockLabel(reflectValue)

		fields, err := getFieldsForBlockStruct(namePrefix, reflectValue.Interface())
		if err != nil {
			return err
		}
		bf.Body = fields
		return nil

	case reflect.Map:
		if reflectValue.Type().Key().Kind() != reflect.String {
			return fmt.Errorf("convertBlock given unsupported map type; expected map[string]T, got %s", reflectValue.Type())
		}

		bf.Name = strings.Join(mergeStringSlices(namePrefix, f.Name), ".")
		bf.Type = "block"

		fields, err := getFieldsForBlockMap(reflectValue)
		if err != nil {
			return err
		}
		bf.Body = fields
		return nil

	default:
		return fmt.Errorf("convertBlock can only be called on struct or map kinds, got %s", reflectValue.Kind())
	}
}

// getBlockLabel returns the label for a given block.
func getBlockLabel(rv reflect.Value) string {
	tags := rivertags.Get(rv.Type())
	for _, tag := range tags {
		if tag.Flags&rivertags.FlagLabel != 0 {
			return reflectutil.Get(rv, tag).String()
		}
	}

	return ""
}

func getFieldsForBlockStruct(namePrefix []string, input interface{}) ([]interface{}, error) {
	val := value.Encode(input)
	reflectVal := val.Reflect()
	rt := rivertags.Get(reflectVal.Type())
	var fields []interface{}
	for _, t := range rt {
		fieldRef := reflectutil.Get(reflectVal, t)
		fieldVal := value.FromRaw(fieldRef)

		if t.IsBlock() && (fieldRef.Kind() == reflect.Array || fieldRef.Kind() == reflect.Slice) {
			for i := 0; i < fieldRef.Len(); i++ {
				arrEle := fieldRef.Index(i).Interface()
				bf, err := newBlock(namePrefix, reflect.ValueOf(arrEle), t)
				if err != nil {
					return nil, err
				}
				if bf.hasValue() {
					fields = append(fields, bf)
				}
			}
		} else if t.IsBlock() {
			bf, err := newBlock(namePrefix, fieldRef, t)

			if err != nil {
				return nil, err
			}
			if bf.hasValue() {
				fields = append(fields, bf)
			}
		} else if t.IsEnum() {
			newPrefix := mergeStringSlices(namePrefix, t.Name)

			for i := 0; i < fieldRef.Len(); i++ {
				innerFields, err := getFieldsForEnum(newPrefix, fieldRef.Index(i))
				if err != nil {
					return nil, err
				}
				fields = append(fields, innerFields...)
			}
		} else if t.IsAttr() {
			af, err := newAttribute(fieldVal, t)
			if err != nil {
				return nil, err
			}
			if af.hasValue() {
				fields = append(fields, af)
			}
		} else if t.IsLabel() {
			// Label is inherent in the block already so this can be a noop
			continue
		} else {
			panic(fmt.Sprintf("river/encoding: unrecognized field %#v", t))
		}
	}
	return fields, nil
}

func getFieldsForBlockMap(val reflect.Value) ([]interface{}, error) {
	var fields []interface{}

	it := val.MapRange()
	for it.Next() {
		// Make a fake field so newAttribute works properly.
		field := rivertags.Field{
			Name:  []string{it.Key().String()},
			Flags: rivertags.FlagAttr,
		}
		attr, err := newAttribute(value.FromRaw(it.Value()), field)
		if err != nil {
			return nil, err
		}

		fields = append(fields, attr)
	}

	return fields, nil
}

func mergeStringSlices(a, b []string) []string {
	if len(a) == 0 {
		return b
	} else if len(b) == 0 {
		return a
	}

	res := make([]string, 0, len(a)+len(b))
	res = append(res, a...)
	res = append(res, b...)
	return res
}

func getFieldsForEnum(name []string, enumElement reflect.Value) ([]interface{}, error) {
	var result []interface{}

	for enumElement.Kind() == reflect.Pointer {
		if enumElement.IsNil() {
			return nil, nil
		}
		enumElement = enumElement.Elem()
	}

	fields := rivertags.Get(enumElement.Type())

	// Find the first non-zero field and encode it as a block.
	for _, field := range fields {
		fieldVal := reflectutil.Get(enumElement, field)
		if !fieldVal.IsValid() || fieldVal.IsZero() {
			continue
		}

		bf, err := newBlock(name, fieldVal, field)
		if err != nil {
			return nil, err
		}
		if bf.hasValue() {
			result = append(result, bf)
		}
		break
	}

	return result, nil
}
