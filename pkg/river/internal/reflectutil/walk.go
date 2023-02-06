package reflectutil

import (
	"reflect"
)

// FieldWalk returns the nested field of value corresponding to index.
// FieldWalk panics if not given a struct.
//
// It is similar to [reflect/Value.FieldByIndex] but can handle traversing
// through nil pointers. If allocate is true, FieldWalk allocates any
// intermediate nil pointers while traversing the struct. If allocate is false,
// FieldWalk returns a non-settable zero value for the final field.
func FieldWalk(value reflect.Value, index []int, allocate bool) reflect.Value {
	if len(index) == 1 {
		return value.Field(index[0])
	}

	if value.Kind() != reflect.Struct {
		panic("FieldWalk must be given a Struct, but found " + value.Kind().String())
	}

	for i, next := range index {
		for value.Kind() == reflect.Pointer {
			if value.IsNil() {
				if !allocate {
					return fieldWalkZero(value, index[i:])
				}
				value.Set(reflect.New(value.Type().Elem()))
			}

			value = value.Elem()
		}

		value = value.Field(next)
	}

	return value
}

// fieldWalkZero returns a non-settable zero value while walking value.
func fieldWalkZero(value reflect.Value, index []int) reflect.Value {
	typ := value.Type()

	for _, next := range index {
		for typ.Kind() == reflect.Pointer {
			typ = typ.Elem()
		}
		typ = typ.Field(next).Type
	}

	return reflect.Zero(typ)
}
