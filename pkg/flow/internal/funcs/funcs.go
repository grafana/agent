// Package funcs defines extra HCL functions.
package funcs

import (
	"os"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

// EnvFunc returns the value of an environment variable by name. If the
// environment variable doesn't exist, an empty string is returned.
var EnvFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name: "var_name",
			Type: cty.String,
		},
	},
	Type: function.StaticReturnType(cty.String),
	Impl: func(args []cty.Value, _ cty.Type) (cty.Value, error) {
		ret := os.Getenv(args[0].AsString())
		return cty.StringVal(ret), nil
	},
})
