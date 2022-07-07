package value

import (
	"encoding"
	"fmt"
	"reflect"
)

// Encode creates a new Value from v. If v is a pointer, v must be considered
// immutable and not change while the Value is used.
func Encode(v interface{}) Value {
	if v == nil {
		return Null
	}
	return makeValue(reflect.ValueOf(v))
}

// Decode assigns a Value val to a Go pointer target. Decode will attempt to
// convert val to the type expected by target for assignment. If val cannot be
// converted, an error is returned. Pointers will be allocated as necessary
// when decoding.
//
// New arrays and slices will be allocated when decoding into target.
//
// River null values will decode into a nil Go pointer or the zero value for
// the non-pointer type.
//
// Decode will panic if target is not a pointer.
func Decode(val Value, target interface{}) error {
	rt := reflect.ValueOf(target)
	if rt.Kind() != reflect.Pointer {
		panic("river/value: Decode called with non-pointer value")
	}
	return decode(val, rt)
}

func decode(val Value, into reflect.Value) error {
	// Before decoding, we need to temporarily take the address of rt so we can
	// handle the case of it implementing supported interfaces.
	if into.CanAddr() {
		into = into.Addr()
	}

	switch {
	case into.Type().Implements(goTextUnmarshaler):
		var s string
		err := decode(val, reflect.ValueOf(&s))
		if err != nil {
			return err
		}
		err = into.Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(s))
		if err != nil {
			return DecodeError{Value: val, Inner: err}
		}
		return nil
	}

	// Fully deference rt and allocate pointers as necessary.
	for into.Kind() == reflect.Pointer {
		// If our value is Null, we want to set the first settable pointer to nil.
		if val.Type() == TypeNull && into.CanSet() {
			into.Set(reflect.Zero(into.Type()))
			return nil
		}

		if into.IsNil() {
			into.Set(reflect.New(into.Type().Elem()))
		}
		into = into.Elem()
	}

	// Fastest cases: we can directly assign values.
	switch {
	case val.rv.Type() == into.Type():
		into.Set(cloneGoValue(val.rv))
		return nil
	case into.Type() == goAny:
		into.Set(cloneGoValue(val.rv))
		return nil
	}

	targetType := RiverType(into.Type())

	// Track a value to use for decoding. This value will be updated if
	// conversion is necessary.
	//
	// NOTE(rfratto): we don't reassign to val here, since Go 1.18 thinks that
	// means it escapes the heap. We need to create a local variable to avoid
	// exctra allocations.
	convVal := val

	// Convert the value.
	switch {
	case convVal.Type() == TypeNull:
		into.Set(reflect.Zero(into.Type()))
	case val.rv.Type() == goByteSlice && into.Type() == goString: // []byte -> string
		into.Set(val.rv.Convert(goString))
		return nil
	case val.rv.Type() == goString && into.Type() == goByteSlice: // string -> []byte
		into.Set(val.rv.Convert(goByteSlice))
		return nil
	case convVal.Type() != targetType:
		var err error
		convVal, err = convertValue(convVal, targetType)
		if err != nil {
			return err
		}
	}

	// Slowest case: recursive decoding. Once we've reached this point, we know
	// that convVal.rv and into are compatible Go types.
	switch convVal.Type() {
	case TypeNumber:
		into.Set(convertGoNumber(convVal.rv, into.Type()))
		return nil
	case TypeString:
		into.Set(convVal.rv)
		return nil
	case TypeBool:
		into.Set(convVal.rv)
		return nil
	case TypeArray:
		return decodeArray(convVal, into)
	case TypeObject:
		return decodeObject(convVal, into)
	case TypeFunction:
		// If the function types had the exact same signature, they would've been
		// handled in the best case statement above. If we've hit this point,
		// they're not the same.
		//
		// For now, we return an error.
		//
		// TODO(rfratto): we may want to consider being more lax here, potentially
		// creating an adapter between the two functions.
		return DecodeError{
			Value: val,
			Inner: fmt.Errorf("expected %s, got %s", into.Type(), convVal.rv.Type()),
		}
	case TypeCapsule:
		// Capsule types require being identical go types, which would've been
		// handled in the best case statement above. If we hit this point, they're
		// not the same.
		//
		// TODO(rfratto): return a TypeError for this instead. TypeError isn't
		// appropriate at the moment because it would just print "capsule", which
		// doesn't contain all the information the user would want to know (e.g., a
		// capsule of what inner type?).
		return DecodeError{
			Value: val,
			Inner: fmt.Errorf("expected capsule(%s), got %s", into.Type(), convVal.Describe()),
		}
	default:
		panic("river/value: unexpected kind " + convVal.Type().String())
	}
}

func decodeArray(val Value, rt reflect.Value) error {
	switch rt.Kind() {
	case reflect.Slice:
		res := reflect.MakeSlice(rt.Type(), val.rv.Len(), val.rv.Len())
		for i := 0; i < val.rv.Len(); i++ {
			// Decode the original elements into the new elements.
			if err := decode(val.Index(i), res.Index(i)); err != nil {
				return ElementError{Value: val, Index: i, Inner: err}
			}
		}
		rt.Set(res)

	case reflect.Array:
		res := reflect.New(rt.Type()).Elem()
		for i := 0; i < val.rv.Len(); i++ {
			// Stop processing elements if the target array is too short.
			// TODO(rfratto): should we force array length to be identical?
			if i >= res.Len() {
				break
			}
			if err := decode(val.Index(i), res.Index(i)); err != nil {
				return ElementError{Value: val, Index: i, Inner: err}
			}
		}
		rt.Set(res)

	default:
		panic(fmt.Sprintf("river/value: unexpected array type %s", val.rv.Kind()))
	}

	return nil
}

func decodeObject(val Value, rt reflect.Value) error {
	switch val.rv.Kind() {
	case reflect.Struct:
		return decodeStructObject(val, rt)
	case reflect.Map:
		return decodeMapObject(val, rt)
	default:
		panic(fmt.Sprintf("river/value: unexpected object type %s", val.rv.Kind()))
	}
}

func decodeStructObject(val Value, rt reflect.Value) error {
	switch rt.Kind() {
	case reflect.Struct:
		// TODO(rfratto): can we find a way to encode optional keys that aren't
		// set?
		sourceTags := getCachedTags(val.rv.Type())
		targetTags := getCachedTags(rt.Type())

		for i := 0; i < sourceTags.Len(); i++ {
			key := sourceTags.Index(i)
			keyValue, _ := val.Key(key.Name)

			// Find the equivalent key in the Go struct.
			target, ok := targetTags.Get(key.Name)
			if !ok {
				return TypeError{Value: val, Expected: RiverType(rt.Type())}
			}
			if err := decodeToField(keyValue, rt, target.Index); err != nil {
				return FieldError{Value: val, Field: key.Name, Inner: err}
			}
		}

	case reflect.Map:
		if rt.Type().Key() != goString {
			// Maps with non-string types are treated as capsules and can't be
			// decoded from objects.
			return TypeError{Value: val, Expected: RiverType(rt.Type())}
		}

		res := reflect.MakeMapWithSize(rt.Type(), val.Len())

		sourceTags := getCachedTags(val.rv.Type())

		for i := 0; i < sourceTags.Len(); i++ {
			keyName := sourceTags.Index(i).Name
			keyValue, _ := val.Key(keyName)

			// Create a new value to hold the entry and decode into it.
			entry := reflect.New(rt.Type().Elem()).Elem()
			if err := decode(keyValue, entry); err != nil {
				return FieldError{Value: val, Field: keyName, Inner: err}
			}

			// Then set the map index.
			res.SetMapIndex(reflect.ValueOf(keyName), entry)
		}
		rt.Set(res)

	default:
		panic(fmt.Sprintf("river/value: unexpected Go object target type %s", rt.Kind()))
	}

	return nil
}

// decodeToField will decode val into a field within intoStruct indexed by the
// index slice. decodeToField will allocate pointers as necessary while
// traversing the struct fields.
func decodeToField(val Value, intoStruct reflect.Value, index []int) error {
	curr := intoStruct
	for _, next := range index {
		for curr.Kind() == reflect.Pointer {
			if curr.IsNil() {
				curr.Set(reflect.New(curr.Type().Elem()))
			}
			curr = curr.Elem()
		}

		curr = curr.Field(next)
	}

	return decode(val, curr)
}

func decodeMapObject(val Value, rt reflect.Value) error {
	switch rt.Kind() {
	case reflect.Struct:
		// TODO(rfratto): can we find a way to encode optional keys that aren't
		// set?
		targetTags := getCachedTags(rt.Type())

		for _, key := range val.Keys() {
			// We ignore the ok value below because we know it exists in the map.
			value, _ := val.Key(key)

			// Find the equivalent key in the Go struct.
			target, ok := targetTags.Get(key)
			if !ok {
				return MissingKeyError{Value: value, Missing: key}
			}

			if err := decodeToField(value, rt, target.Index); err != nil {
				return FieldError{Value: val, Field: key, Inner: err}
			}
		}

	case reflect.Map:
		if rt.Type().Key() != goString {
			// Maps with non-string types are treated as capsules and can't be
			// decoded from maps.
			return TypeError{Value: val, Expected: RiverType(rt.Type())}
		}

		res := reflect.MakeMapWithSize(rt.Type(), val.Len())

		for _, key := range val.Keys() {
			// We ignore the ok value below because we know it exists in the map.
			value, _ := val.Key(key)

			// Create a new value to hold the entry and decode into it.
			entry := reflect.New(rt.Type().Elem()).Elem()
			if err := decode(value, entry); err != nil {
				return FieldError{Value: val, Field: key, Inner: err}
			}

			// Then set the map index.
			res.SetMapIndex(reflect.ValueOf(key), entry)
		}
		rt.Set(res)

	default:
		panic(fmt.Sprintf("river/value: unexpected Go object target type %s", rt.Kind()))
	}

	return nil
}

func cloneGoValue(v reflect.Value) reflect.Value {
	switch v.Kind() {
	case reflect.Array:
		return cloneGoArray(v)
	case reflect.Slice:
		return cloneGoSlice(v)
	case reflect.Map:
		return cloneGoMap(v)
	}

	return v
}

func needsCloned(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Array, reflect.Slice, reflect.Map:
		return true
	default:
		return false
	}
}

func cloneGoArray(in reflect.Value) reflect.Value {
	res := reflect.New(in.Type()).Elem()

	if !needsCloned(in.Type().Elem()) {
		// Optimization: we can use reflect.Copy if the inner type doesn't need to
		// be cloned.
		reflect.Copy(res, in)
		return res
	}

	for i := 0; i < in.Len(); i++ {
		res.Index(i).Set(cloneGoValue(in.Index(i)))
	}
	return res
}

func cloneGoSlice(in reflect.Value) reflect.Value {
	res := reflect.MakeSlice(in.Type(), in.Len(), in.Len())

	if !needsCloned(in.Type().Elem()) {
		// Optimization: we can use reflect.Copy if the inner type doesn't need to
		// be cloned.
		reflect.Copy(res, in)
		return res
	}

	for i := 0; i < in.Len(); i++ {
		res.Index(i).Set(cloneGoValue(in.Index(i)))
	}
	return res
}

func cloneGoMap(in reflect.Value) reflect.Value {
	res := reflect.MakeMapWithSize(in.Type(), in.Len())
	iter := in.MapRange()
	for iter.Next() {
		res.SetMapIndex(cloneGoValue(iter.Key()), cloneGoValue(iter.Value()))
	}
	return res
}
