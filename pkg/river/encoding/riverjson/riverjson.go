// Package riverjson encodes River as JSON.
package riverjson

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/grafana/agent/pkg/river/internal/reflectutil"
	"github.com/grafana/agent/pkg/river/internal/rivertags"
	"github.com/grafana/agent/pkg/river/internal/value"
	"github.com/grafana/agent/pkg/river/token/builder"
)

var goRiverDefaulter = reflect.TypeOf((*value.Defaulter)(nil)).Elem()

// MarshalBody marshals the provided Go value to a JSON representation of
// River. MarshalBody panics if not given a struct with River tags.
func MarshalBody(val interface{}) ([]byte, error) {
	rv := reflect.ValueOf(val)
	return json.Marshal(encodeStructAsBody(rv))
}

func encodeStructAsBody(rv reflect.Value) jsonBody {
	for rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return []jsonStatement{}
		}
		rv = rv.Elem()
	}

	if rv.Kind() == reflect.Invalid {
		return []jsonStatement{}
	} else if rv.Kind() != reflect.Struct {
		panic(fmt.Sprintf("river/encoding/riverjson: can only encode struct values to bodies, got %s", rv.Kind()))
	}

	fields := rivertags.Get(rv.Type())
	defaults := reflect.New(rv.Type()).Elem()
	if defaults.CanAddr() && defaults.Addr().Type().Implements(goRiverDefaulter) {
		defaults.Addr().Interface().(value.Defaulter).SetToDefault()
	}

	body := []jsonStatement{}

	for _, field := range fields {
		fieldVal := reflectutil.Get(rv, field)
		fieldValDefault := reflectutil.Get(defaults, field)

		var isEqual = fieldVal.Comparable() && fieldVal.Equal(fieldValDefault)
		var isZero = fieldValDefault.IsZero() && fieldVal.IsZero()

		if field.IsOptional() && (isEqual || isZero) {
			continue
		}

		body = append(body, encodeFieldAsStatements(nil, field, fieldVal)...)
	}

	return body
}

// encodeFieldAsStatements encodes an individual field from a struct as a set
// of statements. One field may map to multiple statements in the case of a
// slice of blocks.
func encodeFieldAsStatements(prefix []string, field rivertags.Field, fieldValue reflect.Value) []jsonStatement {
	fieldName := strings.Join(field.Name, ".")

	for fieldValue.Kind() == reflect.Pointer {
		if fieldValue.IsNil() {
			break
		}
		fieldValue = fieldValue.Elem()
	}

	switch {
	case field.IsAttr():
		return []jsonStatement{jsonAttr{
			Name:  fieldName,
			Type:  "attr",
			Value: buildJSONValue(value.FromRaw(fieldValue)),
		}}

	case field.IsBlock():
		fullName := mergeStringSlice(prefix, field.Name)

		switch {
		case fieldValue.IsZero():
			// It shouldn't be possible to have a required block which is unset, but
			// we'll encode something anyway.
			return []jsonStatement{jsonBlock{
				Name: strings.Join(fullName, "."),
				Type: "block",
				Body: nil,
			}}

		case fieldValue.Kind() == reflect.Slice, fieldValue.Kind() == reflect.Array:
			statements := []jsonStatement{}

			for i := 0; i < fieldValue.Len(); i++ {
				elem := fieldValue.Index(i)

				// Recursively call encodeField for each element in the slice/array.
				// The recursive call will hit the case below and add a new block for
				// each field encountered.
				statements = append(statements, encodeFieldAsStatements(prefix, field, elem)...)
			}

			return statements

		case fieldValue.Kind() == reflect.Struct:
			return []jsonStatement{jsonBlock{
				Name:  strings.Join(fullName, "."),
				Type:  "block",
				Label: getBlockLabel(fieldValue),
				Body:  encodeStructAsBody(fieldValue),
			}}
		}

	case field.IsEnum():
		// Blocks within an enum have a prefix set.
		newPrefix := mergeStringSlice(prefix, field.Name)

		switch {
		case fieldValue.Kind() == reflect.Slice, fieldValue.Kind() == reflect.Array:
			statements := []jsonStatement{}
			for i := 0; i < fieldValue.Len(); i++ {
				statements = append(statements, encodeEnumElementToStatements(newPrefix, fieldValue.Index(i))...)
			}
			return statements

		default:
			panic(fmt.Sprintf("river/encoding/riverjson: unrecognized enum kind %s", fieldValue.Kind()))
		}
	}

	return nil
}

func mergeStringSlice(a, b []string) []string {
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

func encodeEnumElementToStatements(prefix []string, enumElement reflect.Value) []jsonStatement {
	for enumElement.Kind() == reflect.Pointer {
		if enumElement.IsNil() {
			return nil
		}
		enumElement = enumElement.Elem()
	}

	fields := rivertags.Get(enumElement.Type())

	statements := []jsonStatement{}

	// Find the first non-zero field and encode it.
	for _, field := range fields {
		fieldVal := reflectutil.Get(enumElement, field)
		if !fieldVal.IsValid() || fieldVal.IsZero() {
			continue
		}

		statements = append(statements, encodeFieldAsStatements(prefix, field, fieldVal)...)
		break
	}

	return statements
}

// MarshalValue marshals the provided Go value to a JSON representation of
// River.
func MarshalValue(val interface{}) ([]byte, error) {
	riverValue := value.Encode(val)
	return json.Marshal(buildJSONValue(riverValue))
}

func buildJSONValue(v value.Value) jsonValue {
	if tk, ok := v.Interface().(builder.Tokenizer); ok {
		return jsonValue{
			Type:  "capsule",
			Value: tk.RiverTokenize()[0].Lit,
		}
	}

	switch v.Type() {
	case value.TypeNull:
		return jsonValue{Type: "null"}

	case value.TypeNumber:
		return jsonValue{Type: "number", Value: v.Number().Float()}

	case value.TypeString:
		return jsonValue{Type: "string", Value: v.Text()}

	case value.TypeBool:
		return jsonValue{Type: "bool", Value: v.Bool()}

	case value.TypeArray:
		elements := []interface{}{}

		for i := 0; i < v.Len(); i++ {
			element := v.Index(i)

			elements = append(elements, buildJSONValue(element))
		}

		return jsonValue{Type: "array", Value: elements}

	case value.TypeObject:
		keys := v.Keys()

		// If v isn't an ordered object (i.e., a go map), sort the keys so they
		// have a deterministic print order.
		if !v.OrderedKeys() {
			sort.Strings(keys)
		}

		fields := []jsonObjectField{}

		for i := 0; i < len(keys); i++ {
			field, _ := v.Key(keys[i])

			fields = append(fields, jsonObjectField{
				Key:   keys[i],
				Value: buildJSONValue(field),
			})
		}

		return jsonValue{Type: "object", Value: fields}

	case value.TypeFunction:
		return jsonValue{Type: "function", Value: v.Describe()}

	case value.TypeCapsule:
		return jsonValue{Type: "capsule", Value: v.Describe()}

	default:
		panic(fmt.Sprintf("river/encoding/riverjson: unrecognized value type %q", v.Type()))
	}
}
