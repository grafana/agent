package hcltypes

import (
	"fmt"
	"reflect"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rfratto/gohcl"
	"github.com/zclconf/go-cty/cty"
)

// OptionalSecret holds a potentially sensitive value. When Sensitive is true,
// Value will be treated as a Secret and its value will be hidden from users.
//
// HCL expressions permit converting both Strings and Secrets may be converted
// into OptionalSecret, which will set the Sensitive field accordingly.
//
// HCL expressions may also convert OptionalSecret into a Secret regardless of
// the value of Sensitive. However, OptionalSecret may only be converted into a
// String if Sensitive is false.
type OptionalSecret struct {
	Sensitive bool
	Value     string
}

var optionalSecretTy cty.Type

func init() {
	optionalSecretTy = cty.CapsuleWithOps("OptionalSecret", reflect.TypeOf(OptionalSecret{}), &cty.CapsuleOps{
		ConversionFrom: func(src cty.Type) func(interface{}, cty.Path) (cty.Value, error) {
			switch {
			case src.Equals(cty.String): // OptionalSecret -> Secret
				return func(v interface{}, _ cty.Path) (cty.Value, error) {
					os := v.(*OptionalSecret)
					if os.Sensitive {
						// Only allow conversion to string if the OptionalSecret is non-sensitive.
						return cty.NilVal, fmt.Errorf("cannot convert secret to string")
					}
					return cty.StringVal(os.Value), nil
				}
			case src.Equals(secretTy): // OptionalSecret -> Secret
				// Always allow conversion to a Secret.
				return func(v interface{}, _ cty.Path) (cty.Value, error) {
					os := v.(*OptionalSecret)
					s := Secret(os.Value)
					return cty.CapsuleVal(secretTy, &s), nil
				}
			default:
				return nil
			}
		},

		ConversionTo: func(dst cty.Type) func(cty.Value, cty.Path) (interface{}, error) {
			switch {
			case dst.Equals(cty.String): // String -> OptionalSecret
				return func(v cty.Value, _ cty.Path) (interface{}, error) {
					return &OptionalSecret{Sensitive: false, Value: v.AsString()}, nil
				}
			case dst.Equals(secretTy): // Secret -> OptionalSecret
				return func(v cty.Value, _ cty.Path) (interface{}, error) {
					secret := v.EncapsulatedValue().(*Secret)
					return &OptionalSecret{Sensitive: true, Value: string(*secret)}, nil
				}

			default:
				return nil
			}
		},

		ExtensionData: func(key interface{}) interface{} {
			switch key {
			case gohcl.CapsuleTokenExtensionKey:
				return gohcl.CapsuleTokenExtension(func(v cty.Value) hclwrite.Tokens {
					os := v.EncapsulatedValue().(*OptionalSecret)

					if os.Sensitive {
						return hclwrite.Tokens{
							{Type: hclsyntax.TokenOParen, Bytes: []byte("(")},
							{Type: hclsyntax.TokenIdent, Bytes: []byte("secret")},
							{Type: hclsyntax.TokenCParen, Bytes: []byte(")")},
						}
					}

					return hclwrite.Tokens{
						// NOTE(rfratto): the space before the %q below is intentional; for
						// some reason, for only TokenQuotedLit, no space between = and the
						// value is added unless the token itself contains one.
						{Type: hclsyntax.TokenQuotedLit, Bytes: []byte(fmt.Sprintf(" %q", os.Value))},
					}
				})
			}
			return nil
		},
	})

	gohcl.RegisterCapsuleType(optionalSecretTy)
}
