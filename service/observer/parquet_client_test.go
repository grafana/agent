package observer

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/grafana/agent/pkg/river/encoding/riveragentstate"
	"github.com/parquet-go/parquet-go"
	"github.com/stretchr/testify/require"
)

var agentState riveragentstate.AgentState = riveragentstate.AgentState{
	Labels: map[string]string{
		"app1": ".net",
		"app2": ".net",
	},
}

var agentState2 riveragentstate.AgentState = riveragentstate.AgentState{
	Labels: map[string]string{
		"app1": ".net",
		"app2": ".net",
	},
}

var componentState []riveragentstate.Component = []riveragentstate.Component{
	{
		ID: "module.file.default",
		Health: riveragentstate.Health{
			Health:     "healthy",
			Message:    "Everything is fine",
			UpdateTime: time.Now().UTC(),
		},
		Arguments: []riveragentstate.ComponentDetail{
			{
				ID:         1,
				ParentID:   0,
				Name:       "module.file.default",
				Label:      "module.file.default",
				RiverType:  "file",
				RiverValue: json.RawMessage(`"/var/log/messages"`),
			},
			{
				ID:         2,
				ParentID:   0,
				Name:       "prometheus.remote_write.default",
				Label:      "prometheus.remote_write.default",
				RiverType:  "prometheus",
				RiverValue: json.RawMessage(`"/var/log/messages"`),
			},
			{
				ID:         3,
				ParentID:   0,
				Name:       "prometheus.scrape.first",
				Label:      "prometheus.scrape.first",
				RiverType:  "prometheus",
				RiverValue: json.RawMessage(`"/var/log/messages"`),
			},
		},
	},
	{
		ID: "prometheus.remote_write.default",
		Health: riveragentstate.Health{
			Health:     "healthy",
			Message:    "Everything is fine",
			UpdateTime: time.Now().UTC(),
		},
		Arguments: []riveragentstate.ComponentDetail{},
	},
	{
		ID: "prometheus.scrape.first",
		Health: riveragentstate.Health{
			Health:     "healthy",
			Message:    "Everything is fine",
			UpdateTime: time.Now().UTC(),
		},
		Arguments: []riveragentstate.ComponentDetail{},
	},
	{
		ID:       "module.file.nested",
		ModuleID: "module.file.default",
		Health: riveragentstate.Health{
			Health:     "healthy",
			Message:    "Everything is fine",
			UpdateTime: time.Now().UTC(),
		},
		Arguments: []riveragentstate.ComponentDetail{},
	},
	{
		ID: "prometheus.scrape.second",
		Health: riveragentstate.Health{
			Health:     "healthy",
			Message:    "Everything is fine",
			UpdateTime: time.Now().UTC(),
		},
		Arguments: []riveragentstate.ComponentDetail{},
	},
}

var componentState2 []riveragentstate.Component = []riveragentstate.Component{
	{
		ID: "prometheus.remote_write.default",
		Health: riveragentstate.Health{
			Health:     "healthy",
			Message:    "Everything is fine",
			UpdateTime: time.Now().UTC(),
		},
		Arguments: []riveragentstate.ComponentDetail{
			{
				ID:         1,
				ParentID:   0,
				Name:       "module.file.default",
				Label:      "module.file.default",
				RiverType:  "file",
				RiverValue: json.RawMessage(`"/var/log/messages"`),
			},
		},
	},
}

func TestClientWrite(t *testing.T) {
	client := NewParquetClient(agentState, componentState)
	buf, err := client.Write()
	require.NoError(t, err)
	validateMetadata(t, buf, agentState)
	validateComponentState(t, buf, componentState)
	validateFakeComponentState(t, buf, componentState)
	validateFakeComponent2State(t, buf, componentState)

	// Make sure we can write multiple times without issue.
	client.SetAgentState(agentState2)
	client.SetComponents(componentState2)
	buf, err = client.Write()
	require.NoError(t, err)
	validateMetadata(t, buf, agentState2)
	validateComponentState(t, buf, componentState2)
	validateFakeComponentState(t, buf, componentState2)
	validateFakeComponent2State(t, buf, componentState2)
}

func TestClientWriteToFile(t *testing.T) {
	client := NewParquetClient(agentState, componentState)
	filepath := t.TempDir() + "/agent_state.parquet"
	err := client.WriteToFile(filepath)
	require.NoError(t, err)
	data, err := os.ReadFile(filepath)
	var buffer bytes.Buffer
	buffer.Write(data)
	require.NoError(t, err)
	validateMetadata(t, buffer, agentState)
	validateComponentState(t, buffer, componentState)
	validateFakeComponentState(t, buffer, componentState)
	validateFakeComponent2State(t, buffer, componentState)

	// Make sure we can write multiple times without issue.
	filepath = t.TempDir() + "/agent_state2.parquet"
	client.SetAgentState(agentState2)
	client.SetComponents(componentState2)
	err = client.WriteToFile(filepath)
	require.NoError(t, err)
	data, err = os.ReadFile(filepath)
	buffer.Reset()
	buffer.Write(data)
	require.NoError(t, err)
	validateMetadata(t, buffer, agentState2)
	validateComponentState(t, buffer, componentState2)
	validateFakeComponentState(t, buffer, componentState2)
	validateFakeComponent2State(t, buffer, componentState2)
}

func validateMetadata(t *testing.T, buf bytes.Buffer, expected riveragentstate.AgentState) {
	f, err := parquet.OpenFile(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	require.NoError(t, err)

	for key, label := range expected.Labels {
		value, found := f.Lookup(key)
		require.True(t, found)
		require.Equal(t, label, value)
	}
}

func validateComponentState(t *testing.T, buf bytes.Buffer, expected []riveragentstate.Component) {
	actual, err := parquet.Read[riveragentstate.Component](bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	require.NoError(t, err)

	require.Equal(t, expected, actual)
}

func validateFakeComponentState(t *testing.T, buf bytes.Buffer, expected []riveragentstate.Component) {
	type FakeComponent struct {
		Arguments []riveragentstate.ComponentDetail `parquet:"component_detail"`
	}

	var fakeComponent []FakeComponent
	for _, component := range expected {
		fakeComponent = append(fakeComponent, FakeComponent{Arguments: component.Arguments})
	}

	actual, err := parquet.Read[FakeComponent](bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	require.NoError(t, err)

	require.Equal(t, fakeComponent, actual)
}

func validateFakeComponent2State(t *testing.T, buf bytes.Buffer, expected []riveragentstate.Component) {
	type FakeComponent struct {
		ID       string                 `parquet:"id"`
		ModuleID string                 `parquet:"module_id"`
		Health   riveragentstate.Health `parquet:"health"`
	}

	var fakeComponent []FakeComponent
	for _, component := range expected {
		fakeComponent = append(fakeComponent, FakeComponent{ID: component.ID, ModuleID: component.ModuleID, Health: component.Health})
	}

	actual, err := parquet.Read[FakeComponent](bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	require.NoError(t, err)

	require.Equal(t, fakeComponent, actual)
}
