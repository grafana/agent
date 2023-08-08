package riverjson_test

import (
	"testing"

	"github.com/grafana/agent/pkg/agentstate"
	"github.com/grafana/agent/pkg/river"
	"github.com/grafana/agent/pkg/river/encoding/riverjson"
	"github.com/grafana/agent/pkg/river/rivertypes"
	"github.com/stretchr/testify/require"
)

// TODO: Add a map block. Does it also make sense to have an int block or a string block?
type testBlock2 struct {
	Number  int            `river:"number,attr,optional"`
	String  string         `river:"string,attr,optional"`
	Boolean bool           `river:"boolean,attr,optional"`
	Array   []any          `river:"array,attr,optional"`
	Object  map[string]any `river:"object,attr,optional"`

	Labeled []labeledBlock2 `river:"labeled_block,block,optional"`
}

type labeledBlock2 struct {
	TestBlock testBlock3 `river:",squash"`
	Label     string     `river:",label"`
}

type testBlock3 struct {
	Number int    `river:"number,attr,optional"`
	String string `river:"string,attr,optional"`
}

func TestGetComponentDetail(t *testing.T) {
	// Zero values should be omitted from result.

	val := testBlock2{
		Number:  5,
		String:  "example string value",
		Boolean: true,
		Array:   []any{1, 2, 3},
		Object: map[string]any{
			"key1": "val1",
			"key2": "val2",
		},
		Labeled: []labeledBlock2{
			{
				TestBlock: testBlock3{
					Number: 33,
					String: "asdf",
				},
				Label: "label_a",
			},
			{
				TestBlock: testBlock3{
					Number: 77,
					String: "qwerty",
				},
				Label: "label_b",
			},
		},
	}

	expect := []agentstate.ComponentDetail{
		{
			ID:         1,
			ParentID:   0,
			Name:       "number",
			Label:      "",
			RiverType:  "attr",
			RiverValue: []byte(`{"type":"number","value":5}`),
		},
		{
			ID:         2,
			ParentID:   0,
			Name:       "string",
			Label:      "",
			RiverType:  "attr",
			RiverValue: []byte(`{"type":"string","value":"example string value"}`),
		},
		{
			ID:         3,
			ParentID:   0,
			Name:       "boolean",
			Label:      "",
			RiverType:  "attr",
			RiverValue: []byte(`{"type":"bool","value":true}`),
		},
		{
			ID:         4,
			ParentID:   0,
			Name:       "array",
			Label:      "",
			RiverType:  "attr",
			RiverValue: []byte(`{"type":"array","value":[{"type":"number","value":1},{"type":"number","value":2},{"type":"number","value":3}]}`),
		},
		{
			ID:         5,
			ParentID:   0,
			Name:       "object",
			Label:      "",
			RiverType:  "attr",
			RiverValue: []byte(`{"type":"object","value":[{"key":"key1","value":{"type":"string","value":"val1"}},{"key":"key2","value":{"type":"string","value":"val2"}}]}`),
		},
		{
			ID:        6,
			ParentID:  0,
			Name:      "labeled_block",
			Label:     "label_a",
			RiverType: "block",
		},
		{
			ID:         7,
			ParentID:   6,
			Name:       "number",
			Label:      "",
			RiverType:  "attr",
			RiverValue: []byte(`{"type":"number","value":33}`),
		},
		{
			ID:         8,
			ParentID:   6,
			Name:       "string",
			Label:      "",
			RiverType:  "attr",
			RiverValue: []byte(`{"type":"string","value":"asdf"}`),
		},
		{
			ID:        9,
			ParentID:  0,
			Name:      "labeled_block",
			Label:     "label_b",
			RiverType: "block",
		},
		{
			ID:         10,
			ParentID:   9,
			Name:       "number",
			Label:      "",
			RiverType:  "attr",
			RiverValue: []byte(`{"type":"number","value":77}`),
		},
		{
			ID:         11,
			ParentID:   9,
			Name:       "string",
			Label:      "",
			RiverType:  "attr",
			RiverValue: []byte(`{"type":"string","value":"qwerty"}`),
		},
	}
	actual := riverjson.GetComponentDetail(val)
	require.Equal(t, expect, actual)
}

func TestValues(t *testing.T) {
	tt := []struct {
		name       string
		input      interface{}
		expectJSON string
	}{
		{
			name:       "null",
			input:      nil,
			expectJSON: `{ "type": "null", "value": null }`,
		},
		{
			name:       "number",
			input:      54,
			expectJSON: `{ "type": "number", "value": 54 }`,
		},
		{
			name:       "string",
			input:      "Hello, world!",
			expectJSON: `{ "type": "string", "value": "Hello, world!" }`,
		},
		{
			name:       "bool",
			input:      true,
			expectJSON: `{ "type": "bool", "value": true }`,
		},
		{
			name:  "simple array",
			input: []int{0, 1, 2, 3, 4},
			expectJSON: `{
				"type": "array",
				"value": [
						{ "type": "number", "value": 0 },
						{ "type": "number", "value": 1 },
						{ "type": "number", "value": 2 },
						{ "type": "number", "value": 3 },
						{ "type": "number", "value": 4 }
				]
			}`,
		},
		{
			name:  "nested array",
			input: []interface{}{"testing", []int{0, 1, 2}},
			expectJSON: `{
				"type": "array",
				"value": [
						{ "type": "string", "value": "testing" },
						{
							"type": "array",
							"value": [
								{ "type": "number", "value": 0 },
								{ "type": "number", "value": 1 },
								{ "type": "number", "value": 2 }
							]
						}
				]
			}`,
		},
		{
			name:  "object",
			input: map[string]any{"foo": "bar", "fizz": "buzz", "year": 2023},
			expectJSON: `{
				"type": "object",
				"value": [
					{ "key": "fizz", "value": { "type": "string", "value": "buzz" }},
					{ "key": "foo", "value": { "type": "string", "value": "bar" }},
					{ "key": "year", "value": { "type": "number", "value": 2023 }}
				]
			}`,
		},
		{
			name:       "function",
			input:      func(i int) int { return i * 2 },
			expectJSON: `{ "type": "function", "value": "function" }`,
		},
		{
			name:       "capsule",
			input:      rivertypes.Secret("foo"),
			expectJSON: `{ "type": "capsule", "value": "(secret)" }`,
		},
		{
			// nil arrays and objects must always be [] instead of null as that's
			// what the API definition says they should be.
			name:       "nil array",
			input:      ([]any)(nil),
			expectJSON: `{ "type": "array", "value": [] }`,
		},
		{
			// nil arrays and objects must always be [] instead of null as that's
			// what the API definition says they should be.
			name:       "nil object",
			input:      (map[string]any)(nil),
			expectJSON: `{ "type": "object", "value": [] }`,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := riverjson.MarshalValue(tc.input)
			require.NoError(t, err)
			require.JSONEq(t, tc.expectJSON, string(actual))
		})
	}
}

func TestBlock(t *testing.T) {
	// Zero values should be omitted from result.

	val := testBlock{
		Number: 5,
		Array:  []any{1, 2, 3},
		Labeled: []labeledBlock{
			{
				TestBlock: testBlock{Boolean: true},
				Label:     "label_a",
			},
			{
				TestBlock: testBlock{String: "foo"},
				Label:     "label_b",
			},
		},
		Blocks: []testBlock{
			{String: "hello"},
			{String: "world"},
		},
	}

	expect := `[
		{ 
			"name": "number", 
			"type": "attr", 
			"value": { "type": "number", "value": 5 }
		},
		{
			"name": "array",
			"type": "attr",
			"value": { 
				"type": "array",
				"value": [
					{ "type": "number", "value": 1 },
					{ "type": "number", "value": 2 },
					{ "type": "number", "value": 3 }
				]
			}
		},
		{
			"name": "labeled_block",
			"type": "block",
			"label": "label_a",
			"body": [{
				"name": "boolean",
				"type": "attr",
				"value": { "type": "bool", "value": true }
			}]
		},
		{
			"name": "labeled_block",
			"type": "block",
			"label": "label_b",
			"body": [{
				"name": "string",
				"type": "attr",
				"value": { "type": "string", "value": "foo" }
			}]
		},
		{
			"name": "inner_block",
			"type": "block",
			"body": [{
				"name": "string",
				"type": "attr",
				"value": { "type": "string", "value": "hello" }
			}]
		},
		{
			"name": "inner_block",
			"type": "block",
			"body": [{
				"name": "string",
				"type": "attr",
				"value": { "type": "string", "value": "world" }
			}]
		}
	]`

	actual, err := riverjson.MarshalBody(val)
	require.NoError(t, err)
	require.JSONEq(t, expect, string(actual))
}

func TestBlock_Empty_Required_Block_Slice(t *testing.T) {
	type wrapper struct {
		Blocks []testBlock `river:"some_block,block"`
	}

	tt := []struct {
		name string
		val  any
	}{
		{"nil block slice", wrapper{Blocks: nil}},
		{"empty block slice", wrapper{Blocks: []testBlock{}}},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			expect := `[]`

			actual, err := riverjson.MarshalBody(tc.val)
			require.NoError(t, err)
			require.JSONEq(t, expect, string(actual))
		})
	}
}

type testBlock struct {
	Number  int            `river:"number,attr,optional"`
	String  string         `river:"string,attr,optional"`
	Boolean bool           `river:"boolean,attr,optional"`
	Array   []any          `river:"array,attr,optional"`
	Object  map[string]any `river:"object,attr,optional"`

	Labeled []labeledBlock `river:"labeled_block,block,optional"`
	Blocks  []testBlock    `river:"inner_block,block,optional"`
}

type labeledBlock struct {
	TestBlock testBlock `river:",squash"`
	Label     string    `river:",label"`
}

func TestNilBody(t *testing.T) {
	actual, err := riverjson.MarshalBody(nil)
	require.NoError(t, err)
	require.JSONEq(t, `[]`, string(actual))
}

func TestEmptyBody(t *testing.T) {
	type block struct{}

	actual, err := riverjson.MarshalBody(block{})
	require.NoError(t, err)
	require.JSONEq(t, `[]`, string(actual))
}

func TestHideDefaults(t *testing.T) {
	tt := []struct {
		name       string
		val        defaultsBlock
		expectJSON string
	}{
		{
			name: "no defaults",
			val: defaultsBlock{
				Name: "Jane",
				Age:  41,
			},
			expectJSON: `[
				{ "name": "name", "type": "attr", "value": { "type": "string", "value": "Jane" }},
				{ "name": "age", "type": "attr", "value": { "type": "number", "value": 41 }}
			]`,
		},
		{
			name: "some defaults",
			val: defaultsBlock{
				Name: "John Doe",
				Age:  41,
			},
			expectJSON: `[
				{ "name": "age", "type": "attr", "value": { "type": "number", "value": 41 }}
			]`,
		},
		{
			name: "all defaults",
			val: defaultsBlock{
				Name: "John Doe",
				Age:  35,
			},
			expectJSON: `[]`,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := riverjson.MarshalBody(tc.val)
			require.NoError(t, err)
			require.JSONEq(t, tc.expectJSON, string(actual))
		})
	}
}

type defaultsBlock struct {
	Name string `river:"name,attr,optional"`
	Age  int    `river:"age,attr,optional"`
}

var _ river.Defaulter = (*defaultsBlock)(nil)

func (d *defaultsBlock) SetToDefault() {
	*d = defaultsBlock{
		Name: "John Doe",
		Age:  35,
	}
}

func TestMapBlocks(t *testing.T) {
	type block struct {
		Value map[string]any `river:"block,block,optional"`
	}
	val := block{Value: map[string]any{"field": "value"}}

	expect := `[{
		"name": "block",
		"type": "block",
		"body": [{
			"name": "field",
			"type": "attr",
			"value": { "type": "string", "value": "value" }
		}]
	}]`

	bb, err := riverjson.MarshalBody(val)
	require.NoError(t, err)
	require.JSONEq(t, expect, string(bb))
}
