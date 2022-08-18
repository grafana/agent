package river

import (
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
	References   []string `json:"referencesTo"`
	ReferencedBy []string `json:"referencedBy"`
	Health       *Health  `json:"health"`
	Original     string   `json:"original"`
	Arguments    []*Field `json:"arguments,omitempty"`
	Exports      []*Field `json:"exports,omitempty"`
}

// Field represents a value in river.
type Field struct {
	ID    string      `json:"id,omitempty"`
	Key   string      `json:"key,omitempty"`
	Label string      `json:"label,omitempty"`
	Name  string      `json:"name,omitempty"`
	Type  string      `json:"type,omitempty"`
	Value interface{} `json:"value,omitempty"`
	Body  interface{} `json:"body,omitempty"`
}

// Health represents the health of a component.
type Health struct {
	State       string    `json:"state"`
	Message     string    `json:"message"`
	UpdatedTime time.Time `json:"updatedTime"`
}

// ConvertComponentToJSON converts a set of component information into a generic Field json representation.
func ConvertComponentToJSON(
	id []string,
	args interface{},
	exports interface{},
	references, referencedby []string,
	health *Health,
	original string,
) *ComponentField {

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

	cArgs := convertArguments(args)
	if cArgs != nil {
		nf.Arguments = cArgs
	}
	cExports := convertExports(exports)
	if cExports != nil {
		nf.Exports = cExports
	}
	return nf
}

func convertArguments(args interface{}) []*Field {
	if args == nil {
		return nil
	}
	f := convertStruct(args, nil)
	return f.Value.([]*Field)
}

func convertExports(exports interface{}) []*Field {
	if exports == nil {
		return nil
	}
	f := convertStruct(exports, nil)
	return f.Value.([]*Field)
}

// ConvertToFieldWithName allows conversion of an top level object for testing.
func ConvertToFieldWithName(in interface{}, name string) *Field {
	return ConvertToField(in, &rivertags.Field{
		Name: []string{name},
	})
}

// ConvertToField converts a river object to a JSON field representation.
func ConvertToField(in interface{}, f *rivertags.Field) *Field {
	nf := &Field{}
	if f != nil && f.IsAttr() {
		nf.Type = "attr"
	}
	if f != nil && len(f.Name) > 0 {
		nf.Key = f.Name[len(f.Name)-1]
		nf.ID = f.Name[len(f.Name)-1]
		nf.Name = f.Name[len(f.Name)-1]
	}
	var vIn reflect.Value
	// Find the actual object.
	if in != nil {
		in, _, vIn = getActualStruct(in)
	} else {
		nf.Value = &Field{
			Type: "null",
		}
		return nf
	}

	// Dont write zero value records
	if reflect.ValueOf(in).IsZero() {
		return nil
	}

	// Handle items that explicitly use tokenizer, these are always considered capsule values.
	if tkn, ok := in.(builder.Tokenizer); ok {
		tokens := tkn.RiverTokenize()
		nf.Value = &Field{
			Type:  "capsule",
			Value: tokens[0].Lit,
		}
		return nf
	}

	rt := value.RiverType(reflect.TypeOf(in))
	rv := value.NewValue(reflect.ValueOf(in), rt)
	switch rt {
	case value.TypeNull:
		nf.Value = &Field{
			Type: "null",
		}
		return nf
	case value.TypeNumber:
		numField := &Field{
			Type: "number",
		}
		switch value.MakeNumberKind(vIn.Kind()) {
		case value.NumberKindInt:
			numField.Value = rv.Int()
		case value.NumberKindUint:
			numField.Value = rv.Uint()
		case value.NumberKindFloat:
			numField.Value = rv.Float()
		}
		nf.Value = numField
		return nf
	case value.TypeString:
		nf.Value = &Field{
			Type:  "string",
			Value: rv.Text(),
		}
		return nf
	case value.TypeBool:
		nf.Value = &Field{
			Type:  "bool",
			Value: rv.Bool(),
		}
		return nf
	case value.TypeArray:
		// If this is an array and a block we need to treat those differently. More like an array of blocks
		if f.IsBlock() {
			nf.Type = "block"
			fields := make([]*Field, 0)
			for i := 0; i < vIn.Len(); i++ {
				arrEle := vIn.Index(i).Interface()
				found := convertStruct(arrEle, f)
				if found != nil {
					fields = append(fields, found)
				}
			}
			nf.Body = fields
			return nf
		}
		fields := make([]*Field, 0)
		for i := 0; i < vIn.Len(); i++ {
			arrEle := vIn.Index(i).Interface()
			found := ConvertToField(arrEle, nil)
			if found != nil {
				fields = append(fields, found)
			}
		}

		arrField := &Field{
			Type:  "array",
			Value: fields,
		}
		nf.Value = arrField
		return nf
	case value.TypeObject:
		return convertStruct(in, f)
	case value.TypeFunction:
		panic("func not handled")
	case value.TypeCapsule:
		return &Field{
			Type:  "capsule",
			Value: rv.Describe(),
		}
	}
	return nil
}

func convertStruct(in interface{}, f *rivertags.Field) *Field {
	in, _, vIn := getActualStruct(in)
	nf := &Field{
		Type: "attr",
	}
	if f != nil && len(f.Name) > 0 {
		nf.Key = f.Name[len(f.Name)-1]
		nf.ID = f.Name[len(f.Name)-1]
		nf.Name = f.Name[len(f.Name)-1]
	}
	if vIn.Kind() == reflect.Struct {
		if f != nil && f.IsBlock() {
			nf.Type = "block"
			nf.ID = strings.Join(f.Name, ".")
			// remote_write "t1"
			if len(f.Name) == 2 {
				nf.Name = f.Name[0]
				if f.Name[1] != "" {
					nf.Label = f.Name[1]
				}
			}
		} else {
			nf.Type = "attr"
		}

		fields := make([]*Field, 0)
		riverFields := rivertags.Get(reflect.TypeOf(in))
		for _, rf := range riverFields {
			fieldValue := vIn.FieldByIndex(rf.Index)
			found := ConvertToField(fieldValue.Interface(), &rf)
			if found != nil {
				fields = append(fields, found)
			}
		}
		if nf.Type == "block" {
			nf.Body = fields
		} else {
			nf.Value = fields
		}

		return nf
	} else if vIn.Kind() == reflect.Map {
		if f != nil && f.IsAttr() {
			nf.Type = "attr"
		} else {
			nf.Type = "object"
		}

		fields := make([]*Field, 0)
		iter := vIn.MapRange()
		for iter.Next() {
			mf := &Field{}
			mf.Key = iter.Key().String()
			mf.Name = iter.Key().String()
			mf.Value = ConvertToField(iter.Value().Interface(), nil)
			if mf.Value != nil {
				fields = append(fields, mf)
			}
		}
		nf.Value = fields
		return nf
	} else {
		if f.IsBlock() && f.IsOptional() {
			return nil
		}
	}
	return nil
}

func getActualStruct(in interface{}) (interface{}, reflect.Type, reflect.Value) {
	nt := reflect.TypeOf(in)
	vIn := reflect.ValueOf(in)
	for nt.Kind() == reflect.Pointer && !vIn.IsZero() {
		vIn = vIn.Elem()
		nt = vIn.Type()
	}
	return vIn.Interface(), nt, vIn
}
