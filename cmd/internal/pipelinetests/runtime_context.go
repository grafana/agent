package pipelinetests

import (
	"fmt"
	"os"
	"testing"

	"github.com/phayes/freeport"
	"github.com/stretchr/testify/require"
)

type runtimeContext struct {
	agentPort int
	promData  *promData
}

func newAgentRuntimeContext(t *testing.T) (*runtimeContext, func()) {
	sinkPort, err := freeport.GetFreePort()
	require.NoError(t, err)
	cleanSinkVar := setEnvVariable(t, "HTTP_SINK_URL", fmt.Sprintf("http://127.0.0.1:%d", sinkPort))

	agentPort, err := freeport.GetFreePort()
	require.NoError(t, err)
	cleanAgentPortVar := setEnvVariable(t, "AGENT_SELF_HTTP_PORT", fmt.Sprintf("%d", agentPort))

	agentRuntimeCtx := &runtimeContext{
		agentPort: agentPort,
		promData:  &promData{},
	}

	promServer := newTestPromServer(agentRuntimeCtx.promData.appendPromWrite)
	cleanPromServerVar := setEnvVariable(t, "PROM_SERVER_URL", fmt.Sprintf("%s/api/v1/write", promServer.URL))

	return agentRuntimeCtx, func() {
		promServer.Close()
		cleanSinkVar()
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
