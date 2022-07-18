package value

import (
	"encoding"
	"fmt"
	"reflect"
)

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
			return Error{Value: val, Inner: err}
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
	case val.Type() == TypeNull:
		// TODO(rfratto): Does it make sense for a null to always decode into the
		// zero value? Maybe only objects and arrays should support null?
		into.Set(reflect.Zero(into.Type()))
	case val.rv.Type() == into.Type():
		into.Set(cloneGoValue(val.rv))
		return nil
	case into.Type() == goAny:
		return decodeAny(val, into)
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
		return Error{
			Value: val,
			Inner: fmt.Errorf("expected function(%s), got function(%s)", into.Type(), convVal.rv.Type()),
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
		return Error{
			Value: val,
			Inner: fmt.Errorf("expected capsule(%s), got %s", into.Type(), convVal.Describe()),
		}
	default:
		panic("river/value: unexpected kind " + convVal.Type().String())
	}
}

func decodeAny(val Value, into reflect.Value) error {
	var ptr reflect.Value

	switch val.Type() {
	case TypeNull:
		into.Set(reflect.Zero(into.Type()))
		return nil

	case TypeNumber:
		switch val.Number().Kind() {
		case NumberKindFloat:
			var v float64
			ptr = reflect.ValueOf(&v)
		case NumberKindInt:
			var v int
			ptr = reflect.ValueOf(&v)
		case NumberKindUint:
			var v uint
			ptr = reflect.ValueOf(&v)
		default:
			panic("river/value: unreachable")
		}

	case TypeArray:
		var v []interface{}
		ptr = reflect.ValueOf(&v)

	case TypeObject:
		var v map[string]interface{}
		ptr = reflect.ValueOf(&v)

	case TypeBool:
		var v bool
		ptr = reflect.ValueOf(&v)

	case TypeString:
		var v string
		ptr = reflect.ValueOf(&v)

	case TypeFunction, TypeCapsule:
		// Functions and capsules must be directly assigned since there's no
		// "generic" representation for either.
		into.Set(val.rv)
		return nil

	default:
		panic("river/value: unreachable")
	}

	if err := decode(val, ptr); err != nil {
		return err
	}
	into.Set(ptr.Elem())
	return nil
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

		if val.rv.Len() != res.Len() {
			return Error{
				Value: val,
				Inner: fmt.Errorf("array must have exactly %d elements, got %d", res.Len(), val.rv.Len()),
			}
		}

		for i := 0; i < val.rv.Len(); i++ {
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
	switch rt.Kind() {
	case reflect.Struct:
		targetTags := getCachedTags(rt.Type())
		return decodeObjectToStruct(val, rt, targetTags, false)

	case reflect.Map:
		if rt.Type().Key() != goString {
			// Maps with non-string types are treated as capsules and can't be
			// decoded from maps.
			return TypeError{Value: val, Expected: RiverType(rt.Type())}
		}

		res := reflect.MakeMapWithSize(rt.Type(), val.Len())

		for _, key := range val.Keys() {
			// We ignore the ok value because we know it exists.
			value, _ := val.Key(key)

			// Create a new value to hold the entry and decode into it.
			into := reflect.New(rt.Type().Elem()).Elem()
			if err := decode(value, into); err != nil {
				return FieldError{Value: val, Field: key, Inner: err}
			}

			// Then set the map index.
			res.SetMapIndex(reflect.ValueOf(key), into)
		}

		rt.Set(res)

	default:
		panic(fmt.Sprintf("river/value: unexpected object type %s", val.rv.Kind()))
	}

	return nil
}

func decodeObjectToStruct(val Value, rt reflect.Value, fields *objectFields, decodedLabel bool) error {
	// TODO(rfratto): this needs to check for required keys being set

	for _, key := range val.Keys() {
		// We ignore the ok value because we know it exists.
		value, _ := val.Key(key)

		// Struct labels should be decoded first, since objects are wrapped in
		// labels. If we have yet to decode the label, decode it now.
		if lf, ok := fields.LabelField(); ok && !decodedLabel {
			// Safety check: if the inner field isn't an object, there's something
			// wrong here. It's unclear if a user can craft an expression that hits
			// this case, but it's left in for safety.
			if value.Type() != TypeObject {
				return FieldError{
					Value: val,
					Field: key,
					Inner: TypeError{Value: value, Expected: TypeObject},
				}
			}

			// Decode the key into the label.
			rt.FieldByIndex(lf.Index).Set(reflect.ValueOf(key))

			// ...and then code the rest of the object.
			if err := decodeObjectToStruct(value, rt, fields, true); err != nil {
				return err
			}
			continue
		}

		switch fields.Has(key) {
		case objectKeyTypeInvalid:
			return MissingKeyError{Value: value, Missing: key}
		case objectKeyTypeNestedField:
			next, _ := fields.NestedField(key)
			// Recruse the call with the inner value.
			if err := decodeObjectToStruct(value, rt, next, decodedLabel); err != nil {
				return err
			}
		case objectKeyTypeField:
			targetField, _ := fields.Field(key)
			if err := decodeToField(value, rt, targetField.Index); err != nil {
				return FieldError{Value: val, Field: key, Inner: err}
			}
		}
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
