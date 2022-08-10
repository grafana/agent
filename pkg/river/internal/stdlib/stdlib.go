// Package stdlib contains standard library functions exposed to River configs.
package stdlib

import (
	"encoding/json"
	"os"
	"reflect"

	"github.com/grafana/agent/pkg/river/internal/value"
)

var goAny = reflect.TypeOf((*interface{})(nil)).Elem()

// Functions returns the list of stdlib functions by name. The interface{}
// value is always a River-compatible function value, where functions have at
// least one non-error return value, with an optionally supported error return
// value as the second return value.
var Functions = map[string]interface{}{
	"env": os.Getenv,

	// concat is implemented as a raw function so it can bypass allocations
	// converting arguments into []interface{}. concat is optimized to allow it
	// to perform well when it is in the hot path for combining targets from many
	// other blocks.
	"concat": value.RawFunction(func(funcValue value.Value, args ...value.Value) (value.Value, error) {
		if len(args) == 0 {
			return value.Array(), nil
		}

		// finalSize is the final size of the resulting concatenated array. We type
		// check our arguments while computing what finalSize will be.
		var finalSize int
		for i, arg := range args {
			if arg.Type() != value.TypeArray {
				return value.Null, value.ArgError{
					Function: funcValue,
					Argument: arg,
					Index:    i,
					Inner: value.TypeError{
						Value:    arg,
						Expected: value.TypeArray,
					},
				}
			}

			finalSize += arg.Len()
		}

		// Optimization: if there's only one array, we can just return it directly.
		// This is done *after* the previous loop to ensure that args[0] is a River
		// array.
		if len(args) == 1 {
			return args[0], nil
		}

		// If the imcoming Go slices have the same type, we can have our resulting
		// slice use the same type. This will allow decoding to use the direct
		// assignment optimization.
		//
		// However, if the types don't match, then we're forced to fall back to
		// returning []interface{}.
		//
		// TODO(rfratto): This could fall back to checking the elements if the
		// array/slice types don't match. It would be slower, but the direct
		// assignment optimization probably justifies it.
		useType := args[0].Reflect().Type()
		for i := 1; i < len(args); i++ {
			if args[i].Reflect().Type() != useType {
				useType = reflect.SliceOf(goAny)
				break
			}
		}

		// Build out the final array.
		raw := reflect.MakeSlice(useType, finalSize, finalSize)

		var argNum int
		for _, arg := range args {
			for i := 0; i < arg.Len(); i++ {
				elem := arg.Index(i)
				if elem.Type() != value.TypeNull {
					raw.Index(argNum).Set(elem.Reflect())
				}
				argNum++
			}
		}

		return value.Encode(raw.Interface()), nil
	}),

	"unmarshal_json": func(in string) (interface{}, error) {
		var res interface{}
		err := json.Unmarshal([]byte(in), &res)
		if err != nil {
			return nil, err
		}
		return res, nil
	},
}
