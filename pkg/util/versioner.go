package util

import (
	"fmt"

	"gopkg.in/yaml.v2"
)

// Versioned implements yaml.Unmarshaler for any object, needed when
// yaml.UnmarshalStrict is used. Looks for a "version" field with a
// string value in the object.
type Versioned string

func (v *Versioned) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var fields yaml.MapSlice
	if err := unmarshal(&fields); err != nil {
		return err
	}

	for _, f := range fields {
		field, ok := f.Key.(string)
		if !ok || field != "version" {
			continue
		}

		val, ok := f.Value.(string)
		if !ok {
			return fmt.Errorf("version field must be a string, got %T", f.Value)
		}

		*v = Versioned(val)
		return nil
	}

	return nil
}
