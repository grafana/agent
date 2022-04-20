package hcltypes

import (
	"reflect"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rfratto/gohcl"
	"github.com/zclconf/go-cty/cty"
)

// Secret holds a sensitive value. Any string can be freely converted to a
// Secret, but Secrets may not be re-converted back to strings.
//
// The value of a secret is obscured when rendering HCL.
type Secret string

func init() {
	secretTy := cty.CapsuleWithOps("secret", reflect.TypeOf(Secret("")), &cty.CapsuleOps{
		// We allow strings to be converted into secrets, but don't allow secrets
		// to be converted back to strings to prevent them from being leaked.

		// ConversionTo converts a string to a secret.
		ConversionTo: func(dst cty.Type) func(cty.Value, cty.Path) (interface{}, error) {
			if !dst.Equals(cty.String) {
				return nil
			}
			return func(v cty.Value, _ cty.Path) (interface{}, error) {
				// NOTE(rfratto): capsule values must be pointers to the wrapped type.
				res := Secret(v.AsString())
				return &res, nil
			}
		},

		ExtensionData: func(key interface{}) interface{} {
			switch key {
			case gohcl.CapsuleTokenExtensionKey:
				return gohcl.CapsuleTokenExtension(func(v cty.Value) hclwrite.Tokens {
					return hclwrite.Tokens{
						{Type: hclsyntax.TokenComment, Bytes: []byte("/* secret */")},
					}
				})
			}
			return nil
		},
	})

	gohcl.RegisterCapsuleType(secretTy)
}
