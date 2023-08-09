package riveragentstate_test

import (
	"testing"

	"github.com/grafana/agent/pkg/river/encoding/riveragentstate"
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

	expect := []riveragentstate.ComponentDetail{
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
	actual := riveragentstate.GetComponentDetail(val)
	require.Equal(t, expect, actual)
}
