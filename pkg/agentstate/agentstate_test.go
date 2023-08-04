package agentstate_test

import (
	"bytes"
	"testing"

	"github.com/grafana/agent/pkg/agentstate"
	"github.com/parquet-go/parquet-go"
	"github.com/stretchr/testify/require"
)

func TestAgentState(t *testing.T) {
	expected := []agentstate.State{
		{
			ID: "agent-1",
		},
	}

	var buf bytes.Buffer
	err := parquet.Write(&buf, expected)
	require.NoError(t, err)

	actual, err := parquet.Read[agentstate.State](bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	require.NoError(t, err)

	require.Equal(t, expected, actual)
}
