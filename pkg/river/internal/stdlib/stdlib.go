// Package stdlib contains standard library functions exposed to River configs.
package stdlib

import (
	"encoding/json"
	"os"
)

// Functions returns the list of stdlib functions by name. The interface{}
// value is always a River-compatible function value, where functions have at
// least one non-error return value, with an optionally supported error return
// value as the second return value.
var Functions = map[string]interface{}{
	"env": os.Getenv,

	"concat": func(arrays ...[]interface{}) []interface{} {
		var res []interface{}
		for _, array := range arrays {
			res = append(res, array...)
		}
		return res
	},

	"unmarshal_json": func(in string) (interface{}, error) {
		var res interface{}
		err := json.Unmarshal([]byte(in), &res)
		if err != nil {
			return nil, err
		}
		return res, nil
	},
}
