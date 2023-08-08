package observer

import (
	"testing"

	"github.com/grafana/agent/component"
	"github.com/stretchr/testify/require"
)

func TestGetAgentState(t *testing.T) {
	expected := []Component{}

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
			Arguments:    nil,
			Exports:      nil,
			DebugInfo:    nil,
		},
	}
	actual := getAgentState(components)

	require.Equal(t, expected, actual)
}
