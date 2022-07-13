// Package value holds the internal representation for River values. River
// values act as a lightweight wrapper around reflect.Value.
package value

import (
	"encoding"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// TODO(rfratto): This package is missing three main features:
//
// 1. Ability to invoke function values
// 2. Proper encoding/decoding to Go structs with rvr block tags (currently,
//    labels are ignored)
// 3. Decoding to Go structs with missing required attributes should fail

// Go types used throughout the package.
var (
	goAny             = reflect.TypeOf((*interface{})(nil)).Elem()
	goString          = reflect.TypeOf(string(""))
	goByteSlice       = reflect.TypeOf([]byte(nil))
	goError           = reflect.TypeOf((*error)(nil)).Elem()
	goTextUnmarshaler = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
)

// NOTE(rfratto): This package is extremely sensitive to performance, so
// changes should be made with caution; run benchmarks when changing things.
//
// Value is optimized to be as small as possible and exist fully on the stack.
// This allows many values to avoid allocations, with the exception of creating
// arrays and objects.

// Value represents a River value.
type Value struct {
	rv reflect.Value
	ty Type
}

// Null is the null value.
var Null = Value{}

// Uint returns a Value from a uint64.
func Uint(u uint64) Value { return Value{rv: reflect.ValueOf(u), ty: TypeNumber} }

// Int returns a Value from an int64.
func Int(i int64) Value { return Value{rv: reflect.ValueOf(i), ty: TypeNumber} }

// Float returns a Value from a float64.
func Float(f float64) Value { return Value{rv: reflect.ValueOf(f), ty: TypeNumber} }

// String returns a Value from a string.
func String(s string) Value { return Value{rv: reflect.ValueOf(s), ty: TypeString} }

// Bool returns a Value from a bool.
func Bool(b bool) Value { return Value{rv: reflect.ValueOf(b), ty: TypeBool} }

// Object returns a new value from m. A copy of m is made for producing the
// Value.
func Object(m map[string]Value) Value {
	raw := reflect.MakeMapWithSize(reflect.TypeOf(map[string]interface{}(nil)), len(m))

	for k, v := range m {
		raw.SetMapIndex(reflect.ValueOf(k), v.rv)
	}

	return Value{rv: raw, ty: TypeObject}
}

// Array creates an array from the given values. A copy of the vv slice is made
// for producing the Value.
func Array(vv ...Value) Value {
	ty := reflect.ArrayOf(len(vv), goAny)
	raw := reflect.New(ty).Elem()

	for i, v := range vv {
		if v.ty == TypeNull {
			continue
		}
		raw.Index(i).Set(v.rv)
	}

	return Value{rv: raw, ty: TypeArray}
}

// Func makes a new function Value from f. Func panics if f does not map to a
// River function.
func Func(f interface{}) Value {
	rv := reflect.ValueOf(f)
	if RiverType(rv.Type()) != TypeFunction {
		panic("river/value: Func called with non-function type")
	}
	return Value{rv: rv, ty: TypeFunction}
}

// Capsule creates a new Capsule value from v. Capsule panics if v does not map
// to a River capsule.
func Capsule(v interface{}) Value {
	rv := reflect.ValueOf(v)
	if RiverType(rv.Type()) != TypeCapsule {
		panic("river/value: Capsule called with non-capsule type")
	}
	return Value{rv: rv, ty: TypeCapsule}
}

// Encode creates a new Value from v. If v is a pointer, v must be considered
// immutable and not change while the Value is used.
func Encode(v interface{}) Value {
	if v == nil {
		return Null
	}
	return makeValue(reflect.ValueOf(v))
}

// Type returns the River type for the value.
func (v Value) Type() Type { return v.ty }

// Describe returns a descriptive type name for the value. For capsule values,
// this prints the underlying Go type name. For other values, it prints the
// normal River type.
func (v Value) Describe() string {
	if v.ty != TypeCapsule {
		return v.ty.String()
	}
	return fmt.Sprintf("capsule(%s)", v.rv.Type())
}

// Bool returns the boolean value for v. It panics if v is not a bool.
func (v Value) Bool() bool {
	if v.ty != TypeBool {
		panic("river/value: Bool called on non-bool type")
	}
	return v.rv.Bool()
}

// Int returns an int value for v. It panics if v is not a number.
func (v Value) Int() int64 {
	if v.ty != TypeNumber {
		panic("river/value: Int called on non-number type")
	}
	switch makeNumberKind(v.rv.Kind()) {
	case numberKindInt:
		return v.rv.Int()
	case numberKindUint:
		return int64(v.rv.Uint())
	case numberKindFloat:
		return int64(v.rv.Float())
	}
	panic("river/value: unreachable")
}

// Uint returns an uint value for v. It panics if v is not a number.
func (v Value) Uint() uint64 {
	if v.ty != TypeNumber {
		panic("river/value: Uint called on non-number type")
	}
	switch makeNumberKind(v.rv.Kind()) {
	case numberKindInt:
		return uint64(v.rv.Int())
	case numberKindUint:
		return v.rv.Uint()
	case numberKindFloat:
		return uint64(v.rv.Float())
	}
	panic("river/value: unreachable")
}

// Float returns a float value for v. It panics if v is not a number.
func (v Value) Float() float64 {
	if v.ty != TypeNumber {
		panic("river/value: Float called on non-number type")
	}
	switch makeNumberKind(v.rv.Kind()) {
	case numberKindInt:
		return float64(v.rv.Int())
	case numberKindUint:
		return float64(v.rv.Uint())
	case numberKindFloat:
		return v.rv.Float()
	}
	panic("river/value: unreachable")
}

// Len returns the length of v. Panics if v is not an array or object.
func (v Value) Len() int {
	switch v.ty {
	case TypeArray:
		return v.rv.Len()
	case TypeObject:
		switch v.rv.Kind() {
		case reflect.Struct:
			return getCachedTags(v.rv.Type()).Len()
		case reflect.Map:
			return v.rv.Len()
		}
	}
	panic("river/value: Len called on non-array and non-object value")
}

// Index returns index i of the Value. Panics if the value is not an array or
// if it is out of bounds of the array's size.
func (v Value) Index(i int) Value {
	if v.ty != TypeArray {
		panic("river/value: Index called on non-array value")
	}
	return makeValue(v.rv.Index(i))
}

// makeValue converts a reflect value into a Value, deferencing any pointers or
// interface{} values.
func makeValue(v reflect.Value) Value {
	if !v.IsValid() {
		return Null
	}
	for v.Kind() == reflect.Pointer || v.Type() == goAny {
		v = v.Elem()
		if !v.IsValid() {
			return Null
		}
	}
	return Value{rv: v, ty: RiverType(v.Type())}
}

// Keys returns the keys in v in unspecified order. It panics if v is not an
// object.
func (v Value) Keys() []string {
	if v.ty != TypeObject {
		panic("river/value: Keys called on non-object value")
	}

	switch v.rv.Kind() {
	case reflect.Struct:
		ff := getCachedTags(v.rv.Type())
		return ff.Keys()

	case reflect.Map:
		reflectKeys := v.rv.MapKeys()
		res := make([]string, len(reflectKeys))
		for i, rk := range reflectKeys {
			res[i] = rk.String()
		}
		return res
	}

	panic("river/value: unreachable")
}

// Key returns the value for a key in v. It panics if v is not an object. ok
// will be false if the key did not exist in the object.
func (v Value) Key(key string) (index Value, ok bool) {
	if v.ty != TypeObject {
		panic("river/value: Key called on non-object value")
	}

	switch v.rv.Kind() {
	case reflect.Struct:
		// TODO(rfratto): optimize
		ff := getCachedTags(v.rv.Type())
		f, foundField := ff.Get(key)
		if !foundField {
			return
		}

		val, err := v.rv.FieldByIndexErr(f.Index)
		if err != nil {
			return Null, true
		}
		return makeValue(val), true

	case reflect.Map:
		val := v.rv.MapIndex(reflect.ValueOf(key))
		if !val.IsValid() || val.IsZero() {
			return
		}
		return makeValue(val), true
	}

	panic("river/value: unreachable")
}

func convertValue(val Value, toType Type) (Value, error) {
	// TODO(rfratto): Use vm benchmarks to see if making this a method on Value
	// changes anything.

	fromType := val.Type()

	if fromType == toType {
		// no-op: val is already the right kind.
		return val, nil
	}

	switch fromType {
	case TypeNumber:
		switch toType {
		case TypeString: // number -> string
			strVal := newNumberValue(val.rv).ToString()
			return makeValue(reflect.ValueOf(strVal)), nil
		}

	case TypeString:
		sourceStr := val.rv.String()

		switch toType {
		case TypeNumber: // string -> number
			switch {
			case sourceStr == "":
				return Null, TypeError{Value: val, Expected: toType}

			case sourceStr[0] == '-':
				// String starts with a -; parse as a signed int.
				parsed, err := strconv.ParseInt(sourceStr, 10, 64)
				if err != nil {
					return Null, TypeError{Value: val, Expected: toType}
				}
				return Int(parsed), nil
			case strings.ContainsAny(sourceStr, ".eE"):
				// String contains something that a floating-point number would use;
				// convert.
				parsed, err := strconv.ParseFloat(sourceStr, 64)
				if err != nil {
					return Null, TypeError{Value: val, Expected: toType}
				}
				return Float(parsed), nil
			default:
				// Otherwise, treat the number as an unsigned int.
				parsed, err := strconv.ParseUint(sourceStr, 10, 64)
				if err != nil {
					return Null, TypeError{Value: val, Expected: toType}
				}
				return Uint(parsed), nil
			}
		}
	}

	return Null, TypeError{Value: val, Expected: toType}
}

func convertGoNumber(v reflect.Value, target reflect.Type) reflect.Value {
	nval := newNumberValue(v)

	switch target.Kind() {
	case reflect.Int:
		return reflect.ValueOf(int(nval.Int()))
	case reflect.Int8:
		return reflect.ValueOf(int8(nval.Int()))
	case reflect.Int16:
		return reflect.ValueOf(int16(nval.Int()))
	case reflect.Int32:
		return reflect.ValueOf(int32(nval.Int()))
	case reflect.Int64:
		return reflect.ValueOf(nval.Int())
	case reflect.Uint:
		return reflect.ValueOf(uint(nval.Uint()))
	case reflect.Uint8:
		return reflect.ValueOf(uint8(nval.Uint()))
	case reflect.Uint16:
		return reflect.ValueOf(uint16(nval.Uint()))
	case reflect.Uint32:
		return reflect.ValueOf(uint32(nval.Uint()))
	case reflect.Uint64:
		return reflect.ValueOf(nval.Uint())
	case reflect.Float32:
		return reflect.ValueOf(float32(nval.Float()))
	case reflect.Float64:
		return reflect.ValueOf(nval.Float())
	}

	panic("unsupported number conversion")
}
