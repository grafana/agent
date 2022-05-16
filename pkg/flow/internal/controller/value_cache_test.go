package controller

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zclconf/go-cty/cty"
)

func TestValueCache(t *testing.T) {
	vc := NewValueCache()

	type fooArgs struct {
		Something bool `hcl:"something,attr"`
	}
	type fooExports struct {
		SomethingElse bool `hcl:"something_else,attr"`
	}

	type barArgs struct {
		Number int `hcl:"number,attr"`
	}

	// Emulate values the following HCL file:
	//
	//     foo {
	//       something = true
	//
	//       // Exported fields:
	//       // something_else = bool
	//     }
	//
	//     bar "label_a" {
	//       number = 12
	//     }
	//
	//     bar "label_b" {
	//       number = 34
	//     }

	vc.CacheArguments(ComponentID{"foo"}, fooArgs{Something: true})
	vc.CacheExports(ComponentID{"foo"}, fooExports{SomethingElse: true})
	vc.CacheArguments(ComponentID{"bar", "label_a"}, barArgs{Number: 12})
	vc.CacheArguments(ComponentID{"bar", "label_b"}, barArgs{Number: 34})

	res := vc.BuildContext(nil)

	var (
		expectKeys = []string{"foo", "bar"}
		actualKeys []string
	)
	for varName := range res.Variables {
		actualKeys = append(actualKeys, varName)
	}
	require.ElementsMatch(t, expectKeys, actualKeys)

	expectFoo := cty.ObjectVal(map[string]cty.Value{
		"something":      cty.BoolVal(true),
		"something_else": cty.BoolVal(true),
	})
	expectBar := cty.ObjectVal(map[string]cty.Value{
		"label_a": cty.ObjectVal(map[string]cty.Value{
			"number": cty.NumberIntVal(12),
		}),
		"label_b": cty.ObjectVal(map[string]cty.Value{
			"number": cty.NumberIntVal(34),
		}),
	})

	requireCtyEqual(t, expectFoo, res.Variables["foo"])
	requireCtyEqual(t, expectBar, res.Variables["bar"])
}

// requireCtyEqual() requires a and b to be equal.
func requireCtyEqual(t *testing.T, expect, actual cty.Value) {
	t.Helper()

	if expect.Equals(actual).True() {
		return
	}

	require.Fail(t, fmt.Sprintf("Not equal: \n"+
		"expected: %s\n"+
		"actual  : %s", expect.GoString(), actual.GoString(),
	))
}
