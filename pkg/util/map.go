package util

import "gopkg.in/yaml.v2"

// Map is similar to a yaml.MapSlice but exposes methods for easily adding new
// elements.
type Map []yaml.MapItem

// Set adds or updates a key.
func (m *Map) Set(key string, value interface{}) {
	for i := range *m {
		str, ok := (*m)[i].Key.(string)
		if !ok {
			continue
		}
		if str == key {
			(*m)[i].Value = value
			return
		}
	}

	*m = append(*m, yaml.MapItem{Key: key, Value: value})
}
