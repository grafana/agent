package pipelinetests

import (
	"fmt"
	"os"
	"testing"

	"github.com/phayes/freeport"
	"github.com/stretchr/testify/require"
)

type runtimeContext struct {
	agentPort      int
	dataSentToProm *promData
}

func newAgentRuntimeContext(t *testing.T) (*runtimeContext, func()) {
	agentPort, err := freeport.GetFreePort()
	require.NoError(t, err)
	cleanAgentPortVar := setEnvVariable(t, "AGENT_SELF_HTTP_PORT", fmt.Sprintf("%d", agentPort))

	agentRuntimeCtx := &runtimeContext{
		agentPort:      agentPort,
		dataSentToProm: &promData{},
	}

	promServer := newTestPromServer(agentRuntimeCtx.dataSentToProm.appendPromWrite)
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
