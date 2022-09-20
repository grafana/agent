package vm_test

import (
	"testing"

	"github.com/grafana/agent/pkg/river/ast"
	"github.com/grafana/agent/pkg/river/parser"
	"github.com/grafana/agent/pkg/river/vm"
	"github.com/stretchr/testify/require"
)

// This file contains tests for decoding blocks.

func TestVM_File(t *testing.T) {
	type block struct {
		String string `river:"string,attr"`
		Number int    `river:"number,attr,optional"`
	}
	type file struct {
		SettingA int `river:"setting_a,attr"`
		SettingB int `river:"setting_b,attr,optional"`

		Block block `river:"some_block,block,optional"`
	}

	input := `
	setting_a = 15 

	some_block {
		string = "Hello, world!"
	}
	`

	expect := file{
		SettingA: 15,
		Block: block{
			String: "Hello, world!",
		},
	}

	res, err := parser.ParseFile(t.Name(), []byte(input))
	require.NoError(t, err)

	eval := vm.New(res)

	var actual file
	require.NoError(t, eval.Evaluate(nil, &actual))
	require.Equal(t, expect, actual)
}

func TestVM_Block_Attributes(t *testing.T) {
	t.Run("Decodes attributes", func(t *testing.T) {
		type block struct {
			Number int    `river:"number,attr"`
			String string `river:"string,attr"`
		}

		input := `some_block {
			number = 15 
			string = "Hello, world!"
		}`
		eval := vm.New(parseBlock(t, input))

		var actual block
		require.NoError(t, eval.Evaluate(nil, &actual))
		require.Equal(t, 15, actual.Number)
		require.Equal(t, "Hello, world!", actual.String)
	})

	t.Run("Fails if attribute used as block", func(t *testing.T) {
		type block struct {
			Number int `river:"number,attr"`
		}

		input := `some_block {
			number {} 
		}`
		eval := vm.New(parseBlock(t, input))

		err := eval.Evaluate(nil, &block{})
		require.EqualError(t, err, `"number" must be an attribute, but is used as a block`)
	})

	t.Run("Fails if required attributes are not present", func(t *testing.T) {
		type block struct {
			Number int    `river:"number,attr"`
			String string `river:"string,attr"`
		}

		input := `some_block {
			number = 15 
		}`
		eval := vm.New(parseBlock(t, input))

		err := eval.Evaluate(nil, &block{})
		require.EqualError(t, err, `missing required attribute "string"`)
	})

	t.Run("Succeeds if optional attributes are not present", func(t *testing.T) {
		type block struct {
			Number int    `river:"number,attr"`
			String string `river:"string,attr,optional"`
		}

		input := `some_block {
			number = 15 
		}`
		eval := vm.New(parseBlock(t, input))

		var actual block
		require.NoError(t, eval.Evaluate(nil, &actual))
		require.Equal(t, 15, actual.Number)
		require.Equal(t, "", actual.String)
	})

	t.Run("Fails if attribute is not defined in struct", func(t *testing.T) {
		type block struct {
			Number int `river:"number,attr"`
		}

		input := `some_block {
			number  = 15 
			invalid = "This attribute does not exist!"
		}`
		eval := vm.New(parseBlock(t, input))

		err := eval.Evaluate(nil, &block{})
		require.EqualError(t, err, `3:4: unrecognized attribute name "invalid"`)
	})

	t.Run("Supports arbitrarily nested struct pointer fields", func(t *testing.T) {
		type block struct {
			NumberA int    `river:"number_a,attr"`
			NumberB *int   `river:"number_b,attr"`
			NumberC **int  `river:"number_c,attr"`
			NumberD ***int `river:"number_d,attr"`
		}

		input := `some_block {
			number_a = 15 
			number_b = 20
			number_c = 25
			number_d = 30
		}`
		eval := vm.New(parseBlock(t, input))

		var actual block
		require.NoError(t, eval.Evaluate(nil, &actual))
		require.Equal(t, 15, actual.NumberA)
		require.Equal(t, 20, *actual.NumberB)
		require.Equal(t, 25, **actual.NumberC)
		require.Equal(t, 30, ***actual.NumberD)
	})
}

func TestVM_Block_Children_Blocks(t *testing.T) {
	type childBlock struct {
		Attr bool `river:"attr,attr"`
	}

	t.Run("Decodes children blocks", func(t *testing.T) {
		type block struct {
			Value int        `river:"value,attr"`
			Child childBlock `river:"child.block,block"`
		}

		input := `some_block {
			value = 15 

			child.block { attr = true }
		}`
		eval := vm.New(parseBlock(t, input))

		var actual block
		require.NoError(t, eval.Evaluate(nil, &actual))
		require.Equal(t, 15, actual.Value)
		require.True(t, actual.Child.Attr)
	})

	t.Run("Decodes multiple instances of children blocks", func(t *testing.T) {
		type block struct {
			Value    int          `river:"value,attr"`
			Children []childBlock `river:"child.block,block"`
		}

		input := `some_block {
			value = 10 

			child.block { attr = true }
			child.block { attr = false }
			child.block { attr = true }
		}`
		eval := vm.New(parseBlock(t, input))

		var actual block
		require.NoError(t, eval.Evaluate(nil, &actual))
		require.Equal(t, 10, actual.Value)
		require.Len(t, actual.Children, 3)
		require.Equal(t, true, actual.Children[0].Attr)
		require.Equal(t, false, actual.Children[1].Attr)
		require.Equal(t, true, actual.Children[2].Attr)
	})

	t.Run("Decodes multiple instances of children blocks into an array", func(t *testing.T) {
		type block struct {
			Value    int           `river:"value,attr"`
			Children [3]childBlock `river:"child.block,block"`
		}

		input := `some_block {
			value = 15

			child.block { attr = true }
			child.block { attr = false }
			child.block { attr = true }
		}`
		eval := vm.New(parseBlock(t, input))

		var actual block
		require.NoError(t, eval.Evaluate(nil, &actual))
		require.Equal(t, 15, actual.Value)
		require.Equal(t, true, actual.Children[0].Attr)
		require.Equal(t, false, actual.Children[1].Attr)
		require.Equal(t, true, actual.Children[2].Attr)
	})

	t.Run("Fails if block used as an attribute", func(t *testing.T) {
		type block struct {
			Child childBlock `river:"child,block"`
		}

		input := `some_block {
			child = 15
		}`
		eval := vm.New(parseBlock(t, input))

		err := eval.Evaluate(nil, &block{})
		require.EqualError(t, err, `"child" must be a block, but is used as an attribute`)
	})

	t.Run("Fails if required children blocks are not present", func(t *testing.T) {
		type block struct {
			Value int        `river:"value,attr"`
			Child childBlock `river:"child.block,block"`
		}

		input := `some_block {
			value = 15
		}`
		eval := vm.New(parseBlock(t, input))

		err := eval.Evaluate(nil, &block{})
		require.EqualError(t, err, `missing required block "child.block"`)
	})

	t.Run("Succeeds if optional children blocks are not present", func(t *testing.T) {
		type block struct {
			Value int        `river:"value,attr"`
			Child childBlock `river:"child.block,block,optional"`
		}

		input := `some_block {
			value = 15 
		}`
		eval := vm.New(parseBlock(t, input))

		var actual block
		require.NoError(t, eval.Evaluate(nil, &actual))
		require.Equal(t, 15, actual.Value)
	})

	t.Run("Fails if child block is not defined in struct", func(t *testing.T) {
		type block struct {
			Value int `river:"value,attr"`
		}

		input := `some_block {
			value = 15

			child.block { attr = true }
		}`
		eval := vm.New(parseBlock(t, input))

		err := eval.Evaluate(nil, &block{})
		require.EqualError(t, err, `4:4: unrecognized block name "child.block"`)
	})

	t.Run("Supports arbitrarily nested struct pointer fields", func(t *testing.T) {
		type block struct {
			BlockA childBlock    `river:"block_a,block"`
			BlockB *childBlock   `river:"block_b,block"`
			BlockC **childBlock  `river:"block_c,block"`
			BlockD ***childBlock `river:"block_d,block"`
		}

		input := `some_block {
			block_a { attr = true } 
			block_b { attr = false } 
			block_c { attr = true } 
			block_d { attr = false } 
		}`
		eval := vm.New(parseBlock(t, input))

		var actual block
		require.NoError(t, eval.Evaluate(nil, &actual))
		require.Equal(t, true, (actual.BlockA).Attr)
		require.Equal(t, false, (*actual.BlockB).Attr)
		require.Equal(t, true, (**actual.BlockC).Attr)
		require.Equal(t, false, (***actual.BlockD).Attr)
	})

	// TODO(rfratto): decode all blocks into a []*ast.BlockStmt field.
}

func TestVM_Block_Label(t *testing.T) {
	t.Run("Decodes label into string field", func(t *testing.T) {
		type block struct {
			Label string `river:",label"`
		}

		input := `some_block "label_value_1" {}`
		eval := vm.New(parseBlock(t, input))

		var actual block
		require.NoError(t, eval.Evaluate(nil, &actual))
		require.Equal(t, "label_value_1", actual.Label)
	})

	t.Run("Struct must have label field if block is labeled", func(t *testing.T) {
		type block struct{}

		input := `some_block "label_value_2" {}`
		eval := vm.New(parseBlock(t, input))

		err := eval.Evaluate(nil, &block{})
		require.EqualError(t, err, `1:1: block "some_block" does not support specifying labels`)
	})

	t.Run("Block must have label if struct accepts label", func(t *testing.T) {
		type block struct {
			Label string `river:",label"`
		}

		input := `some_block {}`
		eval := vm.New(parseBlock(t, input))

		err := eval.Evaluate(nil, &block{})
		require.EqualError(t, err, `1:1: block "some_block" requires non-empty label`)
	})

	t.Run("Block must have non-empty label if struct accepts label", func(t *testing.T) {
		type block struct {
			Label string `river:",label"`
		}

		input := `some_block "" {}`
		eval := vm.New(parseBlock(t, input))

		err := eval.Evaluate(nil, &block{})
		require.EqualError(t, err, `1:1: block "some_block" requires non-empty label`)
	})
}

func TestVM_Block_Unmarshaler(t *testing.T) {
	type OuterBlock struct {
		FieldA   string  `river:"field_a,attr"`
		Settings Setting `river:"some.settings,block"`
	}

	input := `
		field_a = "foobar"
		some.settings {
			field_a = "fizzbuzz"
			field_b = "helloworld"
		}
	`

	file, err := parser.ParseFile(t.Name(), []byte(input))
	require.NoError(t, err)

	eval := vm.New(file)

	var actual OuterBlock
	require.NoError(t, eval.Evaluate(nil, &actual))
	require.True(t, actual.Settings.Called, "UnmarshalRiver did not get invoked")
}

type Setting struct {
	FieldA string `river:"field_a,attr"`
	FieldB string `river:"field_b,attr"`

	Called bool
}

func (s *Setting) UnmarshalRiver(f func(interface{}) error) error {
	s.Called = true
	type setting Setting
	return f((*setting)(s))
}

func parseBlock(t *testing.T, input string) *ast.BlockStmt {
	t.Helper()

	res, err := parser.ParseFile("", []byte(input))
	require.NoError(t, err)
	require.Len(t, res.Body, 1)

	stmt, ok := res.Body[0].(*ast.BlockStmt)
	require.True(t, ok, "Expected stmt to be a ast.BlockStmt, got %T", res.Body[0])
	return stmt
}
