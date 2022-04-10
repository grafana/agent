package component

import (
	"fmt"
	"reflect"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rfratto/gohcl"
	"github.com/zclconf/go-cty/cty"
)

var (
	registeredTypeNames = map[string]struct{}{}
)

// RegisterComplexType is used to register a complex type for which values will
// be passed around to components. By default, components are able to pass around
// anything except for structs with unexported fields, channels, or interfaces.
//
// Values of such types may be passed around by registering them here. The
// displayName will be used when rendering them in HCL.
//
// Values of ty must be addressable.
func RegisterComplexType(displayName string, ty reflect.Type) {
	if ty == nil {
		panic("RegisterComplexType called with nil")
	}
	if _, registered := registeredTypeNames[displayName]; registered {
		panic(fmt.Sprintf("Type displayName %q is already in used", displayName))
	}

	capsuleTy := cty.CapsuleWithOps(displayName, ty, &cty.CapsuleOps{
		ExtensionData: func(key interface{}) interface{} {
			switch key {
			case gohcl.CapsuleTokenExtensionKey:
				return gohcl.CapsuleTokenExtension(func(v cty.Value) hclwrite.Tokens {
					return hclwrite.Tokens{
						{Type: hclsyntax.TokenComment, Bytes: []byte(fmt.Sprintf("/* internal %s value */", displayName))},
					}
				})
			}
			return nil
		},
	})
	gohcl.RegisterCapsuleType(capsuleTy)
	registeredTypeNames[displayName] = struct{}{}
}
