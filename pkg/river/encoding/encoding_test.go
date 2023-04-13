package encoding_test

import (
	"io"
	"testing"

	"github.com/grafana/agent/pkg/river/encoding"
	"github.com/stretchr/testify/require"
)

func TestConvertRiverBodyToJSON_CapsuleValue(t *testing.T) {
	type Content struct {
		Field io.Closer `river:"writer,attr"`
	}

	out, err := encoding.ConvertRiverBodyToJSON(Content{Field: io.NopCloser(nil)})
	require.NoError(t, err)

	expect := `[
		{
			"name": "writer",
			"type": "attr",
			"value": {
				"type": "capsule",
				"value": "capsule(\"io.Closer\")"
			}
		}
	]`

	require.JSONEq(t, expect, string(out))
}

func TestConvertRiverBodyToJSON_Label(t *testing.T) {
	// In this test since Name is a label it should NOT be represented in the convertRiverBody
	type Content struct {
		Name     string `river:",label"`
		DummyVal string `river:"dummy,attr,optional"`
	}

	out, err := encoding.ConvertRiverBodyToJSON(Content{
		Name:     "label_test",
		DummyVal: "dummy_test"})
	require.NoError(t, err)

	expect := `[
		{
			"name": "dummy",
			"type": "attr",
			"value": {"type":"string","value":"dummy_test"}
		}
	]`

	require.JSONEq(t, expect, string(out))
}

func TestConvertRiverBodyToJSON_BlockWithZeroValue(t *testing.T) {
	type t1 struct {
		Age int `river:"age,attr"`
	}
	type parent struct {
		Person *t1 `river:"person,block"`
	}

	out, err := encoding.ConvertRiverBodyToJSON(parent{Person: &t1{Age: 0}})
	require.NoError(t, err)

	expect := `[{
		"name": "person",
		"type": "block",
		"body": [{
			"name": "age",
			"type": "attr",
			"value": {
				"type": "number",
				"value": 0
			}
		}]
	}]`

	require.JSONEq(t, expect, string(out))
}

func TestConvertRiverBodyToJSON_Enum_Block(t *testing.T) {
	type InnerBlock struct {
		Number int `river:"number,attr"`
	}

	type EnumBlock struct {
		BlockA *InnerBlock `river:"a,block,optional"`
		BlockB *InnerBlock `river:"b,block,optional"`
		BlockC *InnerBlock `river:"c,block,optional"`
	}

	type Structure struct {
		Field string `river:"field,attr"`

		OtherBlocks []EnumBlock `river:"block,enum"`
	}

	in := Structure{
		Field: "some_value",
		OtherBlocks: []EnumBlock{
			{BlockC: &InnerBlock{Number: 1}},
			{BlockB: &InnerBlock{Number: 2}},
			{BlockC: &InnerBlock{Number: 3}},
		},
	}

	actual, err := encoding.ConvertRiverBodyToJSON(in)
	require.NoError(t, err)

	expect := `[{
		"name": "field",
		"type": "attr",
		"value": {
			"type": "string",
			"value": "some_value"
		}
	}, {
		"name": "block.c", 
		"type": "block",
		"body": [{
			"name": "number",
			"type": "attr",
			"value": { "type": "number", "value": 1 }
		}]
	}, {
		"name": "block.b", 
		"type": "block",
		"body": [{
			"name": "number",
			"type": "attr",
			"value": { "type": "number", "value": 2 }
		}]
	}, {
		"name": "block.c", 
		"type": "block",
		"body": [{
			"name": "number",
			"type": "attr",
			"value": { "type": "number", "value": 3 }
		}]
	}]`

	require.JSONEq(t, expect, string(actual))
}

func TestMapBlocks(t *testing.T) {
	type Body struct {
		Block map[string]string `river:"some_block,block,optional"`
	}

	val := Body{
		Block: map[string]string{"key": "value"},
	}

	actual, err := encoding.ConvertRiverBodyToJSON(val)
	require.NoError(t, err)

	expect := `[{
		"name": "some_block",
		"type": "block",
		"body": [{
			"name": "key",
			"type": "attr",
			"value": { "type": "string", "value": "value" }
		}]
	}]`

	require.JSONEq(t, expect, string(actual))
}
