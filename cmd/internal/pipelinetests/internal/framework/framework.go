package framework

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/grafana/agent/cmd/internal/flowmode"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	defaultTimeout         = 1 * time.Minute
	assertionCheckInterval = 100 * time.Millisecond
	shutdownTimeout        = 5 * time.Second
)

type PipelineTest struct {
	// ConfigFile is the path to the config file to be used for the test.
	ConfigFile string
	// EventuallyAssert is a function that will be called after the agent has started, repeatedly until all assertions
	// are satisfied or the default timeout is reached. The provided context contains all the extra information that
	// the framework has collected, such as data received by the fake prometheus server.
	EventuallyAssert func(t *assert.CollectT, context *RuntimeContext)
	// CmdErrContains is a string that must be contained in the error returned by the command. If empty, no error is
	// expected.
	CmdErrContains string
	// RequireCleanShutdown indicates whether the test framework should verify that the agent shut down cleanly after
	// the test case has completed.
	RequireCleanShutdown bool
}

func RunPipelineTest(t *testing.T, testCase PipelineTest) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)

	cleanUp := setUpGlobalRegistryForTesting(prometheus.NewRegistry())
	defer cleanUp()

	agentRuntimeCtx, cleanUpAgent := newAgentRuntimeContext(t)
	defer cleanUpAgent()

	cmd := flowmode.Command()
	cmd.SetArgs([]string{
		"run",
		testCase.ConfigFile,
		"--server.http.listen-addr",
		fmt.Sprintf("127.0.0.1:%d", agentRuntimeCtx.AgentPort),
		"--storage.path",
		t.TempDir(),
	})

	doneErr := make(chan error)
	go func() { doneErr <- cmd.ExecuteContext(ctx) }()

	assertionsDone := make(chan struct{})
	go func() {
		if testCase.EventuallyAssert != nil {
			require.EventuallyWithT(t, func(t *assert.CollectT) {
				testCase.EventuallyAssert(t, agentRuntimeCtx)
			}, defaultTimeout, assertionCheckInterval)
		}
		assertionsDone <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		t.Fatalf("test case failed to complete within deadline")
	case <-assertionsDone:
	case err := <-doneErr:
		verifyDoneError(t, testCase, err)
		cancel()
		return
	}

	t.Log("assertion checks done, shutting down agent")
	cancel()
	select {
	case <-time.After(shutdownTimeout):
		if testCase.RequireCleanShutdown {
			t.Fatalf("agent failed to shut down within deadline")
		} else {
			t.Log("agent failed to shut down within deadline")
		}
	case err := <-doneErr:
		verifyDoneError(t, testCase, err)
	}
}

func verifyDoneError(t *testing.T, testCase PipelineTest, err error) {
	if testCase.CmdErrContains != "" {
		require.ErrorContains(t, err, testCase.CmdErrContains, "command must return error containing the string specified in test case")
	} else {
		require.NoError(t, err)
	}
}

func setUpGlobalRegistryForTesting(registry *prometheus.Registry) func() {
	prevRegisterer, prevGatherer := prometheus.DefaultRegisterer, prometheus.DefaultGatherer
	prometheus.DefaultRegisterer, prometheus.DefaultGatherer = registry, registry
	return func() {
		prometheus.DefaultRegisterer, prometheus.DefaultGatherer = prevRegisterer, prevGatherer
	}
}
