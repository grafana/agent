// Package stdlib contains standard library functions exposed to River configs.
package stdlib

import (
	"encoding/json"
	"os"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/pkg/river/internal/value"
	"github.com/prometheus/prometheus/discovery/targetgroup"
)

// Identifiers holds a list of stdlib identifiers by name. All interface{}
// values are River-compatible values.
//
// Function identifiers are Go functions with exactly one one non-error return
// value, with an optionally supported error return value as the second return
// value.
var Identifiers = map[string]interface{}{
	// See constants.go for the definition.
	"constants": constants,

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

		raw := make([]value.Value, 0, finalSize)
		for _, arg := range args {
			for i := 0; i < arg.Len(); i++ {
				raw = append(raw, arg.Index(i))
			}
		}

		return value.Array(raw...), nil
	}),

	"json_decode": func(in string) (interface{}, error) {
		var res interface{}
		err := json.Unmarshal([]byte(in), &res)
		if err != nil {
			return nil, err
		}
		return res, nil
	},

	"discovery_target_decode": func(in string) (interface{}, error) {
		var targetGroups []*targetgroup.Group
		if err := json.Unmarshal([]byte(in), &targetGroups); err != nil {
			return nil, err
		}

		var res []discovery.Target

		for _, group := range targetGroups {
			for _, target := range group.Targets {

				// Create the output target from group and target labels. Target labels
				// should override group labels.
				outputTarget := make(discovery.Target, len(group.Labels)+len(target))
				for k, v := range group.Labels {
					outputTarget[string(k)] = string(v)
				}
				for k, v := range target {
					outputTarget[string(k)] = string(v)
				}

				res = append(res, outputTarget)
			}
		}

		return res, nil
	},
}
