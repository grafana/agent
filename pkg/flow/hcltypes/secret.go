package hcltypes

import (
	"reflect"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rfratto/gohcl"
	"github.com/zclconf/go-cty/cty"
)

// Secret holds a sensitive value. Secrets are never displayed to the user when
// rendering HCL.
//
// HCL expressions permit implicitly converting string values to a Secret, but
// not the inverse. This ensures that a user can't accidentally leak a
// sensitive value.
type Secret string

var secretTy cty.Type

func init() {
	secretTy = cty.CapsuleWithOps("secret", reflect.TypeOf(Secret("")), &cty.CapsuleOps{
		// We allow strings to be converted into secrets, but don't allow secrets
		// to be converted back to strings to prevent them from being leaked.

		ConversionFrom: func(src cty.Type) func(interface{}, cty.Path) (cty.Value, error) {
			switch {
			case src.Equals(optionalSecretTy): // Secret -> OptionalSecret
				return func(v interface{}, _ cty.Path) (cty.Value, error) {
					return cty.CapsuleVal(optionalSecretTy, &OptionalSecret{
						Sensitive: true,
						Value:     string(*v.(*Secret)),
					}), nil
				}
			default:
				return nil
			}
		},

		ConversionTo: func(dst cty.Type) func(cty.Value, cty.Path) (interface{}, error) {
			switch {
			case dst.Equals(cty.String): // string -> Secret
				return func(v cty.Value, _ cty.Path) (interface{}, error) {
					res := Secret(v.AsString())
					return &res, nil
				}
			case dst.Equals(optionalSecretTy): // OptionalSecret -> Secret
				return func(v cty.Value, _ cty.Path) (interface{}, error) {
					os := v.EncapsulatedValue().(*OptionalSecret)
					res := Secret(os.Value)
					return &res, nil
				}
			default:
				return nil
			}
		},

		ExtensionData: func(key interface{}) interface{} {
			switch key {
			case gohcl.CapsuleTokenExtensionKey:
				return gohcl.CapsuleTokenExtension(func(v cty.Value) hclwrite.Tokens {
					return hclwrite.Tokens{
						{Type: hclsyntax.TokenOParen, Bytes: []byte("(")},
						{Type: hclsyntax.TokenIdent, Bytes: []byte("secret")},
						{Type: hclsyntax.TokenCParen, Bytes: []byte(")")},
					}
				})
			}
			return nil
		},
	})

	gohcl.RegisterCapsuleType(secretTy)
}
