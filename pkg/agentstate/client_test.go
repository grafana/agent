package agentstate_test

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/grafana/agent/pkg/agentstate"
	"github.com/parquet-go/parquet-go"
	"github.com/stretchr/testify/require"
)

var agentState agentstate.AgentState = agentstate.AgentState{
	ID: "agent-1",
	Labels: map[string]string{
		"app1": ".net",
		"app2": ".net",
	},
}

var componentState []agentstate.Component = []agentstate.Component{
	{
		ID: "module.file.default",
		Health: agentstate.Health{
			Health:     "healthy",
			Message:    "Everything is fine",
			UpdateTime: time.Now().UTC(),
		},
		ComponentDetail: []agentstate.ComponentDetail{
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
		Health: agentstate.Health{
			Health:     "healthy",
			Message:    "Everything is fine",
			UpdateTime: time.Now().UTC(),
		},
		ComponentDetail: []agentstate.ComponentDetail{},
	},
	{
		ID: "prometheus.scrape.first",
		Health: agentstate.Health{
			Health:     "healthy",
			Message:    "Everything is fine",
			UpdateTime: time.Now().UTC(),
		},
		ComponentDetail: []agentstate.ComponentDetail{},
	},
	{
		ID:       "module.file.nested",
		ModuleID: "module.file.default",
		Health: agentstate.Health{
			Health:     "healthy",
			Message:    "Everything is fine",
			UpdateTime: time.Now().UTC(),
		},
		ComponentDetail: []agentstate.ComponentDetail{},
	},
	{
		ID: "prometheus.scrape.second",
		Health: agentstate.Health{
			Health:     "healthy",
			Message:    "Everything is fine",
			UpdateTime: time.Now().UTC(),
		},
		ComponentDetail: []agentstate.ComponentDetail{},
	},
}

func TestClient(t *testing.T) {
	client := agentstate.NewClient(agentState, componentState)
	err := client.Write()
	require.NoError(t, err)
	err = client.Write()
	require.NoError(t, err)

	validateMetadata(t, client.Buf(), agentState)
	validateComponentState(t, client.Buf(), componentState)
	validateFakeComponentState(t, client.Buf(), componentState)
	validateFakeComponent2State(t, client.Buf(), componentState)
}

func validateMetadata(t *testing.T, buf bytes.Buffer, expected agentstate.AgentState) {
	f, err := parquet.OpenFile(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	require.NoError(t, err)

	value, found := f.Lookup("ID")
	require.True(t, found)
	require.Equal(t, expected.ID, value)

	for key, label := range expected.Labels {
		value, found = f.Lookup(key)
		require.True(t, found)
		require.Equal(t, label, value)
	}
}

func validateComponentState(t *testing.T, buf bytes.Buffer, expected []agentstate.Component) {
	actual, err := parquet.Read[agentstate.Component](bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	require.NoError(t, err)

	require.Equal(t, expected, actual)
}

func validateFakeComponentState(t *testing.T, buf bytes.Buffer, expected []agentstate.Component) {
	type FakeComponent struct {
		ComponentDetail []agentstate.ComponentDetail `parquet:"component_detail"`
	}

	var fakeComponent []FakeComponent
	for _, component := range expected {
		fakeComponent = append(fakeComponent, FakeComponent{ComponentDetail: component.ComponentDetail})
	}

	actual, err := parquet.Read[FakeComponent](bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	require.NoError(t, err)

	require.Equal(t, fakeComponent, actual)
}

func validateFakeComponent2State(t *testing.T, buf bytes.Buffer, expected []agentstate.Component) {
	type FakeComponent struct {
		ID       string            `parquet:"id"`
		ModuleID string            `parquet:"module_id"`
		Health   agentstate.Health `parquet:"health"`
	}

	var fakeComponent []FakeComponent
	for _, component := range expected {
		fakeComponent = append(fakeComponent, FakeComponent{ID: component.ID, ModuleID: component.ModuleID, Health: component.Health})
	}

	actual, err := parquet.Read[FakeComponent](bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	require.NoError(t, err)

	require.Equal(t, fakeComponent, actual)
}
