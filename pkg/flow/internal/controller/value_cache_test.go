package controller

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValueCache(t *testing.T) {
	vc := newValueCache()

	type fooArgs struct {
		Something bool `river:"something,attr"`
	}
	type fooExports struct {
		SomethingElse bool `river:"something_else,attr"`
	}

	type barArgs struct {
		Number int `river:"number,attr"`
	}

	// Emulate values from the following River file:
	//
	//     foo {
	//       something = true
	//
	//       // Exported fields:
	//       // something_else = true
	//     }
	//
	//     bar "label_a" {
	//       number = 12
	//     }
	//
	//     bar "label_b" {
	//       number = 34
	//     }
	//
	// and expects to generate the equivalent to the following River object:
	//
	//     {
	//      	foo = {
	//      		something_else = true,
	//      	},
	//
	//      	bar = {
	//      		label_a = {},
	//      		label_b = {},
	//      	}
	//     }
	//
	// For now, only exports are placed in generated objects, which is why the
	// bar values are empty and the foo object only contains the exports.

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

	type object = map[string]interface{}

	expectFoo := fooExports{SomethingElse: true}
	expectBar := object{
		"label_a": object{}, // no exports for bar
		"label_b": object{}, // no exports for bar
	}
	require.Equal(t, expectFoo, res.Variables["foo"])
	require.Equal(t, expectBar, res.Variables["bar"])
}

func TestExportValueCache(t *testing.T) {
	vc := newValueCache()
	vc.CacheModuleExportValue("t1", 1)
	require.True(t, vc.HasModuleExportsChangedSinceLastCall())
	// The call to HasModuleExportsChangedSinceLastCall sets the value to false.
	require.False(t, vc.HasModuleExportsChangedSinceLastCall())

	vc.CacheModuleExportValue("t1", 2)
	require.True(t, vc.HasModuleExportsChangedSinceLastCall())
	require.False(t, vc.HasModuleExportsChangedSinceLastCall())

	vc.CacheModuleExportValue("t1", 2)
	require.False(t, vc.HasModuleExportsChangedSinceLastCall())

	vc.CacheModuleExportValue("t2", "test")
	require.True(t, vc.HasModuleExportsChangedSinceLastCall())

	vc.ClearModuleExports()
	require.True(t, vc.HasModuleExportsChangedSinceLastCall())
}
