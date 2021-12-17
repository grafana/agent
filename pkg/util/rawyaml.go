package util

import "gopkg.in/yaml.v2"

// RawYAML is similar to json.RawMessage and allows for deferred YAML decoding.
type RawYAML []byte

// UnmarshalYAML implements yaml.Unmarshaler.
func (r *RawYAML) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var ms yaml.MapSlice
	if err := unmarshal(&ms); err != nil {
		return err
	}
	bb, err := yaml.Marshal(ms)
	if err != nil {
		return err
	}
	*r = bb
	return nil
}

// MarshalYAML implements yaml.Marshaler.
func (r RawYAML) MarshalYAML() (interface{}, error) {
	return string(r), nil
}
