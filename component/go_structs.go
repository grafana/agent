package component

import (
	"fmt"
	"reflect"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rfratto/gohcl"
	"github.com/zclconf/go-cty/cty"
)

// RegisterGoStruct can be used to create a custom HCL type whose value is a Go
// struct pointer. This allows to expose structs which normally can't be
// encoded or decoded to HCL, such as a struct which contains a Go interface or
// channel.
//
// RegisterGoStruct should be called in the init method for the package which
// declares the custom type.
//
// The displayName of the GoStruct is what will be shown to end users as the
// type name.
//
// v must be a struct value and not a pointer. RegisterGoStruct will panic if v
// is an unexpected type.
//
// Config and Exports can then expose the registered Go struct through a
// pointer to the registered struct type. For example:
//
//     type MyCustomStruct struct{ Stream <-chan int }
//
//     func init() {
//         component.RegisterGoStruct("MyCustomStruct", MyCustomStruct{})
//     }
//
//     type Exports struct {
//         CustomValue *MyCustomStruct `hcl:"custom_value,attr"`
//     }
func RegisterGoStruct(displayName string, v interface{}) {
	ty := reflect.TypeOf(v)
	if ty.Kind() != reflect.Struct {
		panic(fmt.Sprintf("RegisterGoStruct called with %T, expected struct type", v))
	}

	capsuleTy := cty.CapsuleWithOps(displayName, ty, &cty.CapsuleOps{
		ExtensionData: func(key interface{}) interface{} {
			switch key {
			case gohcl.CapsuleTokenExtensionKey:
				// gohcl will fail to normally encode cty capsule values to HCL; this
				// extension is used to provide custom encoding.
				return gohcl.CapsuleTokenExtension(func(_ cty.Value) hclwrite.Tokens {
					return hclwrite.Tokens{
						{Type: hclsyntax.TokenComment, Bytes: []byte(fmt.Sprintf("/* %s value */", displayName))},
					}
				})
			}
			return nil
		},
	})
	gohcl.RegisterCapsuleType(capsuleTy)
}
