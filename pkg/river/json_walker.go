package river

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/grafana/agent/pkg/river/token/builder"

	"github.com/grafana/agent/pkg/river/internal/rivertags"
	"github.com/grafana/agent/pkg/river/internal/value"
)

// ComponentField represents a component in river.
type ComponentField struct {
	Field        `json:",omitempty"`
	References   []string    `json:"referencesTo"`
	ReferencedBy []string    `json:"referencedBy"`
	Health       *Health     `json:"health"`
	Original     string      `json:"original"`
	Arguments    interface{} `json:"arguments,omitempty"`
	Exports      interface{} `json:"exports,omitempty"`
	DebugInfo    interface{} `json:"debugInfo,omitempty"`
}

// Field represents a value in river.
type Field struct {
	ID    string      `json:"id,omitempty"`
	Key   string      `json:"key,omitempty"`
	Label string      `json:"label,omitempty"`
	Name  string      `json:"name,omitempty"`
	Type  string      `json:"type,omitempty"`
	Value interface{} `json:"value,omitempty"`
	Body  []*Field    `json:"body,omitempty"`
}

// Health represents the health of a component.
type Health struct {
	State       string    `json:"state"`
	Message     string    `json:"message"`
	UpdatedTime time.Time `json:"updatedTime"`
}

const attr = "attr"

// ConvertComponentToJSON converts a set of component information into a generic Field json representation.
func ConvertComponentToJSON(
	id []string,
	args interface{},
	exports interface{},
	debug interface{},
	references, referencedby []string,
	health *Health,
	original string,
) (*ComponentField, error) {

	nf := &ComponentField{
		Field: Field{
			ID:    strings.Join(id, "."),
			Name:  strings.Join(id[0:2], "."),
			Type:  "block",
			Value: nil,
		},
		References:   references,
		ReferencedBy: referencedby,
		Health:       health,
		Original:     original,
	}
	if len(id) == 3 {
		nf.Label = id[2]
	}

	cArgs, err := convertComponentChild(args)
	if err != nil {
		return nil, err
	}
	if cArgs != nil {
		nf.Arguments = cArgs
	}
	cExports, err := convertComponentChild(exports)
	if err != nil {
		return nil, err
	}
	if cExports != nil {
		nf.Exports = cExports
	}
	cDebug, err := convertComponentChild(debug)
	if err != nil {
		return nil, err
	}
	if cDebug != nil {
		nf.DebugInfo = cDebug
	}
	return nf, nil
}

// convertComponentChild is used to convert arguments, exports, health and debuginfo.
func convertComponentChild(input interface{}) ([]*Field, error) {
	if input == nil {
		return nil, nil
	}
	_, vt, vIn := getActualValue(input)
	fields := make([]*Field, 0)
	rt := rivertags.Get(vt)
	for _, t := range rt {
		fieldValue := vIn.FieldByIndex(t.Index)
		fieldIn, fieldT, fieldVal := getActualValue(fieldValue.Interface())
		// Blocks can only happen at this level
		if t.IsBlock() && (fieldT.Kind() == reflect.Array || fieldT.Kind() == reflect.Slice) {
			for i := 0; i < fieldVal.Len(); i++ {
				arrEle := fieldVal.Index(i).Interface()
				found, err := convertBlock(arrEle, &t)
				if err != nil {
					return nil, err
				}
				if found != nil {
					fields = append(fields, found)
				}
			}
		} else if t.IsBlock() {
			found, err := convertBlock(fieldIn, &t)
			if err != nil {
				return nil, err
			}
			if found != nil {
				fields = append(fields, found)
			}
		} else {
			found, err := convertAttribute(fieldIn, vt, &t)
			if err != nil {
				return nil, err
			}
			if found != nil {
				fields = append(fields, found)
			}
		}
	}
	return fields, nil
}

// convertThing is the switchboard for conversions
func convertThing(in interface{}, f *rivertags.Field) (*Field, error) {
	if in == nil {
		return nil, nil
	}

	in, inType, inValue := getActualValue(in)
	if inType.Kind() == reflect.Pointer {
		return nil, nil
	}
	if inType.Kind() == reflect.Struct && inValue.IsZero() {
		return nil, nil
	}
	if f != nil && f.IsBlock() {
		return convertBlock(in, f)
	}

	// Capsules would be seen as a struct below but we want to treat them special.
	if value.RiverType(inType) == value.TypeCapsule {
		return convertValue(in)
	}

	if inType.Kind() == reflect.Array || inType.Kind() == reflect.Slice {
		return convertArray(inType, inValue, f)
	} else if inType.Kind() == reflect.Struct {
		// Normal structures have river tags so handle them like structs. But some things regex, time, etc act like
		// structs but need to be considered a value. They distill into string, so create an attr to hold them and move on.
		if f != nil && f.IsAttr() {
			v, err := convertValue(in)
			if err != nil {
				return nil, err
			}
			return &Field{
				Type:  attr,
				Value: v,
				Name:  strings.Join(f.Name, "."),
			}, nil
		} else if !rivertags.HasRiverTags(inType) {
			// This is used when the caller is converting a value directly.
			v, err := convertValue(in)
			if err != nil {
				return nil, err
			}
			return v, nil
		} else {
			return convertStruct(in, inValue, f)
		}
	} else if inType.Kind() == reflect.Map {
		return convertMap(inValue)
	}
	if f != nil {
		return convertAttribute(in, inType, f)
	}
	return convertValue(in)
}

func convertAttribute(in interface{}, t reflect.Type, f *rivertags.Field) (*Field, error) {
	if !f.IsAttr() {
		return nil, fmt.Errorf("convertAttribute requires a field that is an attribute got %T", in)
	}
	nf := &Field{
		Name: strings.Join(f.Name, "."),
		Type: attr,
	}
	if isValue(t) {
		v, err := convertValue(in)
		if err != nil {
			return nil, err
		}
		nf.Value = v
	} else {
		v, err := convertThing(in, nil)
		if err != nil {
			return nil, err
		}
		nf.Value = v
	}
	if nf.Value == nil || reflect.ValueOf(nf.Value).IsZero() {
		return nil, nil
	}
	return nf, nil
}

func convertArray(inType reflect.Type, inValue reflect.Value, f *rivertags.Field) (*Field, error) {
	if inType.Kind() != reflect.Array && inType.Kind() != reflect.Slice {
		return nil, fmt.Errorf("convertArray requires a field that is an slice/array got %T", inValue.Interface())
	}
	nf := &Field{}
	nf.Type = "array"

	blocks := make([]interface{}, 0)
	for i := 0; i < inValue.Len(); i++ {
		arrEle := inValue.Index(i).Interface()
		found, err := convertThing(arrEle, f)
		if err != nil {
			return nil, err
		}
		if found != nil {
			blocks = append(blocks, found)
		}
	}

	nf.Value = blocks
	return nf, nil
}

func isValue(t reflect.Type) bool {
	rt := value.RiverType(t)
	switch rt {
	case value.TypeNull, value.TypeNumber, value.TypeString, value.TypeBool, value.TypeFunction, value.TypeCapsule:
		return true
	}
	return false
}

// convertValue is used to transform the underlying value of a river tag to a field
func convertValue(in interface{}) (*Field, error) {
	in, _, vIn := getActualValue(in)

	if reflect.ValueOf(in).IsZero() {
		return nil, nil
	}
	// Handle items that explicitly use tokenizer, these are always considered capsule values.
	if tkn, ok := in.(builder.Tokenizer); ok {
		tokens := tkn.RiverTokenize()
		return &Field{
			Type:  "capsule",
			Value: tokens[0].Lit,
		}, nil
	}
	newV := value.MakeValue(vIn)
	rt := value.RiverType(reflect.TypeOf(in))
	switch rt {
	case value.TypeNull:
		return &Field{
			Type: "null",
		}, nil
	case value.TypeNumber:
		return &Field{
			Type:  "number",
			Value: newV.Interface(),
		}, nil
	case value.TypeString:
		return &Field{
			Type:  "string",
			Value: newV.Text(),
		}, nil
	case value.TypeBool:
		return &Field{
			Type:  "bool",
			Value: newV.Bool(),
		}, nil
	case value.TypeArray:
		return nil, fmt.Errorf("convertValue does not allow array types")
	case value.TypeObject:
		return nil, fmt.Errorf("convertValue does not allow object types")
	case value.TypeFunction:
		return &Field{
			Type:  "function",
			Value: newV.Describe(),
		}, nil
	case value.TypeCapsule:
		return &Field{
			Type:  "capsule",
			Value: newV.Describe(),
		}, nil
	}
	return nil, fmt.Errorf("error while converting value to json %T", in)
}

func convertBlock(in interface{}, f *rivertags.Field) (*Field, error) {
	if in == nil {
		return nil, nil
	}
	in, _, vIn := getActualValue(in)
	if vIn.Kind() == reflect.Pointer && (vIn.IsZero() || vIn.IsNil()) {
		return nil, nil
	}

	if vIn.Kind() != reflect.Struct {
		return nil, fmt.Errorf("convertBlock cannot work on interface or slices")
	}

	nf := &Field{
		Name: strings.Join(f.Name, "."),
		Type: "block",
	}
	body := make([]*Field, 0)

	riverFields := rivertags.Get(reflect.TypeOf(in))
	for _, rf := range riverFields {
		fieldValue := vIn.FieldByIndex(rf.Index)
		found, err := convertThing(fieldValue.Interface(), &rf)
		if err != nil {
			return nil, err
		}
		if found != nil {
			body = append(body, found)
		}
	}
	// If we have no content then return nil
	if len(body) == 0 {
		return nil, nil
	}
	nf.Body = body
	return nf, nil
}

func convertMap(vIn reflect.Value) (*Field, error) {
	fields := make([]*Field, 0)
	iter := vIn.MapRange()
	for iter.Next() {
		mf := &Field{}
		mf.Key = iter.Key().String()
		val, vt, _ := getActualValue(iter.Value().Interface())
		hasTags := rivertags.HasRiverTags(vt)
		if hasTags {
			v, err := convertThing(iter.Value().Interface(), nil)
			if err != nil {
				return nil, err
			}
			mf.Value = v
		} else {
			v, err := convertValue(val)
			if err != nil {
				return nil, err
			}
			mf.Value = v
		}

		if mf.Value != nil {
			fields = append(fields, mf)
		}
	}
	if len(fields) == 0 {
		return nil, nil
	}
	return &Field{
		Type:  "object",
		Value: fields,
	}, nil
}

func convertStruct(in interface{}, vIn reflect.Value, f *rivertags.Field) (*Field, error) {
	nf := &Field{}
	if f != nil && len(f.Name) > 0 {
		nf.Name = f.Name[len(f.Name)-1]
	}
	if vIn.Kind() != reflect.Struct {
		return nil, fmt.Errorf("convertStruct cannot work on non-structs")
	}
	nf.Type = attr
	fields := make([]interface{}, 0)
	riverFields := rivertags.Get(reflect.TypeOf(in))
	for _, rf := range riverFields {
		fieldValue := vIn.FieldByIndex(rf.Index)
		found, err := convertThing(fieldValue.Interface(), &rf)
		if err != nil {
			return nil, err
		}
		if found != nil {
			fields = append(fields, found)
		}
	}
	nf.Value = fields
	return nf, nil
}

// getActualValue is used to find the concrete value
func getActualValue(in interface{}) (interface{}, reflect.Type, reflect.Value) {
	nt := reflect.TypeOf(in)
	vIn := reflect.ValueOf(in)
	for nt.Kind() == reflect.Pointer && !vIn.IsZero() {
		vIn = vIn.Elem()
		nt = vIn.Type()
	}
	return vIn.Interface(), nt, vIn
}
