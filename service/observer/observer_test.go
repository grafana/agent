package observer

import (
	"encoding/json"
	"testing"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/pkg/river/encoding/riverparquet"
	"github.com/stretchr/testify/require"
)

func TestGetAgentState(t *testing.T) {
	//TODO: How can we get a slice with component info?
	components := []*component.Info{
		{
			Component:    nil,
			ModuleIDs:    []string{},
			ID:           component.ID{},
			Label:        "",
			References:   []string{},
			ReferencedBy: []string{},
			Registration: component.Registration{},
			Health:       component.Health{},
			Arguments:    testBlock{Name: "John", Age: 32},
			Exports:      testBlock{Name: "Jane", Age: 35},
			DebugInfo:    testBlock{Name: "Peter", Age: 49},
		},
	}

	expected := []componentRow{{
		ID:       "",
		ModuleID: "",
		Health: componentHealth{
			Health: "unknown",
		},
		Arguments: []riverparquet.Row{
			{
				ID:         1,
				ParentID:   0,
				Name:       "name",
				RiverType:  "attr",
				RiverValue: json.RawMessage(`{"type":"string","value":"John"}`),
			},
			{
				ID:         2,
				ParentID:   0,
				Name:       "age",
				RiverType:  "attr",
				RiverValue: json.RawMessage(`{"type":"number","value":32}`),
			},
		},
		Exports: []riverparquet.Row{
			{
				ID:         1,
				ParentID:   0,
				Name:       "name",
				RiverType:  "attr",
				RiverValue: json.RawMessage(`{"type":"string","value":"Jane"}`),
			},
			{
				ID:         2,
				ParentID:   0,
				Name:       "age",
				RiverType:  "attr",
				RiverValue: json.RawMessage(`{"type":"number","value":35}`),
			},
		},
		DebugInfo: []riverparquet.Row{
			{
				ID:         1,
				ParentID:   0,
				Name:       "name",
				RiverType:  "attr",
				RiverValue: json.RawMessage(`{"type":"string","value":"Peter"}`),
			},
			{
				ID:         2,
				ParentID:   0,
				Name:       "age",
				RiverType:  "attr",
				RiverValue: json.RawMessage(`{"type":"number","value":49}`),
			},
		},
	}}
	actual := convertToParquetRows(components)

	require.Equal(t, expected, actual)
}

type testBlock struct {
	Name string `river:"name,attr"`
	Age  int    `river:"age,attr"`
}
