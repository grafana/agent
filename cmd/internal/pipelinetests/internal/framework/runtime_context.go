package framework

import (
	"fmt"
	"net"
	"os"
	"testing"

	"github.com/phayes/freeport"
	"github.com/stretchr/testify/require"
)

type RuntimeContext struct {
	// AgentPort is the port the agent's HTTP server is listening on.
	AgentPort int
	// DataSentToProm is a collection of data sent to the fake test Prometheus server.
	DataSentToProm *PromData
	// DataSentToLoki is a collection of data sent to the fake test Loki server.
	DataSentToLoki *LokiData
	// TestTarget is a fake test target that can be used to expose metrics that can be scraped by the agent.
	TestTarget *TestTarget
}

func newAgentRuntimeContext(t *testing.T) (*RuntimeContext, func()) {
	agentPort, err := freeport.GetFreePort()
	require.NoError(t, err)
	cleanAgentPortVar := setEnvVariable(t, "AGENT_SELF_HTTP_PORT", fmt.Sprintf("%d", agentPort))

	testTarget := newTestTarget()
	cleanTestTargetVar := setEnvVariable(t, "TEST_TARGET_ADDRESS", fmt.Sprintf("127.0.0.1:%d", testTarget.server.Listener.Addr().(*net.TCPAddr).Port))

	agentRuntimeCtx := &RuntimeContext{
		AgentPort:      agentPort,
		DataSentToProm: &PromData{},
		DataSentToLoki: &LokiData{},
		TestTarget:     testTarget,
	}

	promServer := newTestPromServer(agentRuntimeCtx.DataSentToProm.appendPromWrite)
	cleanPromServerVar := setEnvVariable(t, "PROM_SERVER_URL", fmt.Sprintf("%s/api/v1/write", promServer.URL))

	lokiServer := newTestLokiServer(agentRuntimeCtx.DataSentToLoki.appendLokiWrite)
	cleanLokiServerVar := setEnvVariable(t, "LOKI_SERVER_URL", lokiServer.URL)

	return agentRuntimeCtx, func() {
		cleanAgentPortVar()
		testTarget.server.Close()
		cleanTestTargetVar()
		promServer.Close()
		cleanPromServerVar()
		lokiServer.Close()
		cleanLokiServerVar()
	}
}

func setEnvVariable(t *testing.T, key, value string) func() {
	require.NoError(t, os.Setenv(key, value))
	return func() {
		require.NoError(t, os.Unsetenv(key))
	}
}
