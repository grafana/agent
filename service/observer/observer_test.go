package observer

import (
	"encoding/json"
	"testing"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/pkg/river/encoding/riverjson"
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

	expected := []Component{{
		ID:       "",
		ModuleID: "",
		Health: Health{
			Health: "unknown",
		},
		ComponentDetail: []riverjson.ComponentDetail{
			{
				ID:        1,
				ParentID:  0,
				Name:      "arguments",
				RiverType: "block",
			},
			{
				ID:         2,
				ParentID:   1,
				Name:       "name",
				RiverType:  "attr",
				RiverValue: json.RawMessage(`{"type":"string","value":"John"}`),
			},
			{
				ID:         3,
				ParentID:   1,
				Name:       "age",
				RiverType:  "attr",
				RiverValue: json.RawMessage(`{"type":"number","value":32}`),
			},
			{
				ID:        4,
				ParentID:  0,
				Name:      "exports",
				RiverType: "block",
			},
			{
				ID:         5,
				ParentID:   4,
				Name:       "name",
				RiverType:  "attr",
				RiverValue: json.RawMessage(`{"type":"string","value":"Jane"}`),
			},
			{
				ID:         6,
				ParentID:   4,
				Name:       "age",
				RiverType:  "attr",
				RiverValue: json.RawMessage(`{"type":"number","value":35}`),
			},
			{
				ID:        7,
				ParentID:  0,
				Name:      "debug_info",
				RiverType: "block",
			},
			{
				ID:         8,
				ParentID:   7,
				Name:       "name",
				RiverType:  "attr",
				RiverValue: json.RawMessage(`{"type":"string","value":"Peter"}`),
			},
			{
				ID:         9,
				ParentID:   7,
				Name:       "age",
				RiverType:  "attr",
				RiverValue: json.RawMessage(`{"type":"number","value":49}`),
			},
		},
	}}
	actual := getAgentState(components)

	require.Equal(t, expected, actual)
}

type testBlock struct {
	Name string `river:"name,attr"`
	Age  int    `river:"age,attr"`
}
