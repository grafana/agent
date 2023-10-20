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
	defaultTimeout         = 10 * time.Second
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
	// Timeout is the maximum amount of time the test case is allowed to run. If 0, defaultTimeout is used.
	Timeout time.Duration
	// Environment is a map of environment variables to be set before running the test. It will be automatically
	// cleaned. The values can be used inside the config files using the `env("ENV_VAR")` syntax.
	Environment map[string]string
}

func (p PipelineTest) RunTest(t *testing.T) {
	if p.Timeout == 0 {
		p.Timeout = defaultTimeout
	}
	// Main context has some padding added to the timeout to allow for assertions error message to surface first.
	ctx, cancel := context.WithTimeout(context.Background(), p.Timeout+2*assertionCheckInterval)

	for k, v := range p.Environment {
		cleanUp := setEnvVariable(t, k, v)
		//goland:noinspection GoDeferInLoop
		defer cleanUp()
	}

	cleanUp := setUpGlobalRegistryForTesting(prometheus.NewRegistry())
	defer cleanUp()

	agentRuntimeCtx, cleanUpAgent := newAgentRuntimeContext(t)
	defer cleanUpAgent()

	cmd := flowmode.Command()
	cmd.SetArgs([]string{
		"run",
		p.ConfigFile,
		"--server.http.listen-addr",
		fmt.Sprintf("127.0.0.1:%d", agentRuntimeCtx.AgentPort),
		"--storage.path",
		t.TempDir(),
	})

	doneErr := make(chan error)
	go func() { doneErr <- cmd.ExecuteContext(ctx) }()

	assertionsDone := make(chan struct{})
	go func() {
		if p.EventuallyAssert != nil {
			require.EventuallyWithT(t, func(t *assert.CollectT) {
				p.EventuallyAssert(t, agentRuntimeCtx)
			}, p.Timeout, assertionCheckInterval)
		}
		assertionsDone <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		t.Fatalf("test case failed to complete within deadline")
	case <-assertionsDone:
	case err := <-doneErr:
		verifyDoneError(t, p, err)
		cancel()
		return
	}

	t.Log("assertion checks done, shutting down agent")
	cancel()
	select {
	case <-time.After(shutdownTimeout):
		if p.RequireCleanShutdown {
			t.Fatalf("agent failed to shut down within deadline")
		} else {
			t.Log("agent failed to shut down within deadline")
		}
	case err := <-doneErr:
		verifyDoneError(t, p, err)
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
