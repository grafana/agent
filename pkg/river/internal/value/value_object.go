package value

import "reflect"

// Most Go values can be represented and decoded directly. Objects are
// different, since each field may indicate a deeply nested value.
//
// TODO(rfratto): document more

type structWrapper struct {
	structVal reflect.Value
	fields    *objectFields
	label     string // Non-empty string if this struct is wrapped in a label.
}

func wrapStruct(val reflect.Value) structWrapper {
	if val.Kind() != reflect.Struct {
		panic("river/value: wrapStruct called on non-struct value")
	}

	fields := getCachedTags(val.Type())

	var label string
	if f, ok := fields.LabelField(); ok {
		label = val.FieldByIndex(f.Index).String()
	}

	return structWrapper{
		structVal: val,
		fields:    fields,
		label:     label,
	}
}

func (sw structWrapper) Len() int {
	if len(sw.label) > 0 {
		return 1
	}
	return sw.fields.Len()
}

func (sw structWrapper) Keys() []string {
	if len(sw.label) > 0 {
		return []string{sw.label}
	}
	return sw.fields.Keys()
}

func (sw structWrapper) Key(key string) (index Value, ok bool) {
	if len(sw.label) > 0 {
		if key != sw.label {
			return
		}
		next := reflect.ValueOf(structWrapper{
			structVal: sw.structVal,
			fields:    sw.fields,
			// Unset the label now that we've traversed it
		})
		return Value{rv: next, ty: TypeObject}, true
	}

	keyType := sw.fields.Has(key)

	switch keyType {
	case objectKeyTypeInvalid:
		return // No such key

	case objectKeyTypeNestedField:
		// Continue traversing.
		nextNode, _ := sw.fields.NestedField(key)
		return Value{
			rv: reflect.ValueOf(structWrapper{
				structVal: sw.structVal,
				fields:    nextNode,
			}),
			ty: TypeObject,
		}, true

	case objectKeyTypeField:
		f, _ := sw.fields.Field(key)
		val, err := sw.structVal.FieldByIndexErr(f.Index)
		if err != nil {
			return Null, true
		}
		return makeValue(val), true
	}

	panic("river/value: unreachable")
}
