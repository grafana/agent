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

type setter func(value reflect.Value) error

func Convert(from interface{}, to interface{}, cfg ConversionCfg) error {
	return walkFields(from, func(fieldValue reflect.Value, structField reflect.StructField) error {

		frName := determineFieldName(structField, cfg.FromTags, cfg.FromTagNameExtractor)

		setter, err := findSetter(to, frName, cfg)
		if err != nil {
			fmt.Printf("error finding setter for %q: %s\n", frName, err)
			return nil
		}

		err = setter(fieldValue)
		if err != nil {
			fmt.Printf("error setting value for %q: %s\n", frName, err)
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
				return setValue(fieldValue, v)
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

func setValue(target reflect.Value, val reflect.Value) error {
	if !val.CanSet() {
		return fmt.Errorf("value not settable: %v", val.Kind())
	}

	if !val.Type().AssignableTo(target.Type()) {
		return fmt.Errorf("provided value is not assignable to %v", target.Type())
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
	v := reflect.ValueOf(s)

	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct && v.Kind() != reflect.Interface {
		return fmt.Errorf("not a struct or interface: %v", v)
	}

	for i := 0; i < v.NumField(); i++ {
		if err := f(v.Field(i), v.Type().Field(i)); err != nil {
			return err
		}
	}
	return nil
}

// FistInCSV returns the first element of a comma-separated string.
func FistInCSV(s string) string {
	return strings.Split(s, ",")[0]
}
