package auto

import (
	"fmt"
	"reflect"
	"strings"
)

type ConversionCfg struct {
	FromTags string
	ToTags   string

	FromTagNameExtractor func(string) string
	ToTagNameExtractor   func(string) string
}

var YamlToRiver = ConversionCfg{
	FromTags:             "yaml",
	ToTags:               "river",
	FromTagNameExtractor: FistInCSV,
	ToTagNameExtractor:   FistInCSV,
}

var RiverToYaml = ConversionCfg{
	FromTags:             "river",
	ToTags:               "yaml",
	FromTagNameExtractor: FistInCSV,
	ToTagNameExtractor:   FistInCSV,
}

type setter func(value reflect.Value) error

func ConvertByFieldNames(from interface{}, to interface{}, cfg ConversionCfg) error {

	fromVal := getDereferencedValue(from)
	if fromVal.Kind() == reflect.Slice {
		toVal := getDereferencedValue(to)
		return convertSlice(fromVal, toVal, cfg)
	} else {
		return convertNonSlice(from, to, cfg)

	}
}

func convertSlice(fromSlice reflect.Value, toSlice reflect.Value, cfg ConversionCfg) error {
	if fromSlice.Kind() != reflect.Slice || toSlice.Kind() != reflect.Slice {
		return fmt.Errorf(
			"both source and target values should be a slice, got: %v, %v",
			fromSlice.Kind(),
			toSlice.Kind(),
		)
	}

	// Iterate over the slice of A and call the conversion function on each element
	for i := 0; i < fromSlice.Len(); i++ {
		a := fromSlice.Index(i)
		var b reflect.Value
		if a.Kind() == reflect.Ptr {
			b = reflect.New(toSlice.Type().Elem().Elem())
			err := ConvertByFieldNames(a.Interface(), b.Interface(), cfg)
			if err != nil {
				return err
			}
			toSlice.Index(i).Set(b.Elem().Addr())
		} else {
			a = a.Addr()
			b = reflect.New(toSlice.Type().Elem())
			err := ConvertByFieldNames(a.Interface(), b.Interface(), cfg)
			if err != nil {
				return err
			}
			toSlice.Index(i).Set(b.Elem())
		}
	}

	return nil
}

func convertNonSlice(from interface{}, to interface{}, cfg ConversionCfg) error {
	return walkFields(from, func(fieldValue reflect.Value, structField reflect.StructField) error {

		frName := determineFieldName(structField, cfg.FromTags, cfg.FromTagNameExtractor)

		setter, err := findSetter(to, frName, cfg)
		if err != nil {
			fmt.Printf("error finding setter for %q: %s\n", frName, err)
			return nil
		}

		err = setter(fieldValue)
		if err != nil {
			fmt.Printf("error setting fromVal for %q: %s\n", frName, err)
			return nil
		}

		fmt.Printf("set field %q to %v\n", frName, fieldValue)
		return nil
	})
}

func determineFieldName(
	structField reflect.StructField,
	tags string,
	tagNameExtractor func(string) string,
) string {
	frName := structField.Tag.Get(tags)
	if tagNameExtractor != nil {
		frName = tagNameExtractor(frName)
	}
	if frName == "" {
		frName = structField.Name
	}
	return frName
}

func findSetter(to interface{}, name string, cfg ConversionCfg) (setter, error) {
	var setter setter
	err := walkFields(to, func(fieldValue reflect.Value, structField reflect.StructField) error {
		toName := determineFieldName(structField, cfg.ToTags, cfg.ToTagNameExtractor)
		if toName == name {
			setter = func(v reflect.Value) error {
				return setValue(v, fieldValue, cfg)
			}
			return nil
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if setter == nil {
		return nil, fmt.Errorf("no setter found for %q", name)
	}
	return setter, nil
}

func setValue(val reflect.Value, target reflect.Value, cfg ConversionCfg) error {
	if !val.CanSet() {
		return fmt.Errorf("value not settable: %v", val.Kind())
	}

	if !val.Type().AssignableTo(target.Type()) {
		fmt.Printf("value %v not assignable to %v\n", val.Type(), target.Type())
		nestedTarget := target
		nestedValue := val
		if target.Kind() != reflect.Ptr {
			nestedTarget = target.Addr()
		}
		if val.Kind() != reflect.Ptr {
			nestedValue = val.Addr()
		}

		if nestedValue.Elem().Kind() == reflect.Slice {
			targetSlice := reflect.MakeSlice(nestedTarget.Type().Elem(), nestedValue.Elem().Len(), nestedValue.Elem().Len())
			nestedTarget.Elem().Set(targetSlice)
		}

		if nestedTarget.IsNil() && !nestedValue.IsNil() {
			nestedTarget.Set(reflect.New(nestedTarget.Type().Elem()))
		}

		return ConvertByFieldNames(nestedValue.Interface(), nestedTarget.Interface(), cfg)
	}

	// Use the appropriate Set method based on the type of the value
	switch target.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		target.SetInt(val.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		target.SetUint(val.Uint())
	case reflect.Float32, reflect.Float64:
		target.SetFloat(val.Float())
	case reflect.Bool:
		target.SetBool(val.Bool())
	case reflect.String:
		target.SetString(val.String())
	default:
		target.Set(val)
	}
	return nil
}

type walker func(fieldValue reflect.Value, structField reflect.StructField) error

func walkFields(s interface{}, f walker) error {
	//fmt.Printf("walking fields of %+v\n", s)
	v := getDereferencedValue(s)

	if !v.CanAddr() {
		return fmt.Errorf("cannot walk unaddressable value: %+v", v)
	}

	if v.Kind() != reflect.Struct {
		return fmt.Errorf("not a struct: %v", v)
	}

	for i := 0; i < v.NumField(); i++ {
		if err := f(v.Field(i), v.Type().Field(i)); err != nil {
			return err
		}
	}
	return nil
}

func getDereferencedValue(s interface{}) reflect.Value {
	v := reflect.ValueOf(s)

	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	return v
}

// FistInCSV returns the first element of a comma-separated string.
func FistInCSV(s string) string {
	return strings.Split(s, ",")[0]
}
