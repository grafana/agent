package framework

import (
	"fmt"
	"os"
	"testing"

	"github.com/phayes/freeport"
	"github.com/stretchr/testify/require"
)

type RuntimeContext struct {
	AgentPort      int
	DataSentToProm *PromData
}

func newAgentRuntimeContext(t *testing.T) (*RuntimeContext, func()) {
	agentPort, err := freeport.GetFreePort()
	require.NoError(t, err)
	cleanAgentPortVar := setEnvVariable(t, "AGENT_SELF_HTTP_PORT", fmt.Sprintf("%d", agentPort))

	agentRuntimeCtx := &RuntimeContext{
		AgentPort:      agentPort,
		DataSentToProm: &PromData{},
	}

	promServer := newTestPromServer(agentRuntimeCtx.DataSentToProm.appendPromWrite)
	cleanPromServerVar := setEnvVariable(t, "PROM_SERVER_URL", fmt.Sprintf("%s/api/v1/write", promServer.URL))

	return agentRuntimeCtx, func() {
		promServer.Close()
		cleanAgentPortVar()
		cleanPromServerVar()
	}
}

func setEnvVariable(t *testing.T, key, value string) func() {
	require.NoError(t, os.Setenv(key, value))
	return func() {
		require.NoError(t, os.Unsetenv(key))
	}
}
