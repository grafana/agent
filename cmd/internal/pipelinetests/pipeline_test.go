package pipelinetests

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

type pipelineTest struct {
	configFile           string
	eventuallyAssert     func(t *assert.CollectT, context *runtimeContext)
	cmdErrContains       string
	requireCleanShutdown bool
}

func TestPipeline_WithEmptyConfig(t *testing.T) {
	runTestCase(t, pipelineTest{
		configFile:           "testdata/empty.river",
		requireCleanShutdown: true,
	})
}

func TestPipeline_FileNotExists(t *testing.T) {
	runTestCase(t, pipelineTest{
		configFile:           "does_not_exist.river",
		cmdErrContains:       "does_not_exist.river: no such file or directory",
		requireCleanShutdown: true,
	})
}

func TestPipeline_FileInvalid(t *testing.T) {
	runTestCase(t, pipelineTest{
		configFile:           "testdata/invalid.river",
		cmdErrContains:       "could not perform the initial load successfully",
		requireCleanShutdown: true,
	})
}

func TestPipeline_Prometheus_SelfScrapeAndWrite(topT *testing.T) {
	runTestCase(topT, pipelineTest{
		configFile: "testdata/scrape_and_write.river",
		eventuallyAssert: func(t *assert.CollectT, context *runtimeContext) {
			writes := context.promData.getPromWrites()
			assert.NotEmptyf(t, writes, "must receive at least one prom write request")
			assert.Greater(t, context.promData.sampleValueForSeries("agent_prometheus_forwarded_samples_total"), float64(1000))
			assert.Greater(t, context.promData.sampleValueForSeries("agent_wal_samples_appended_total"), float64(1000))
			assert.Equal(t, context.promData.sampleValueForSeries("agent_prometheus_scrape_targets_gauge"), float64(1))
		},
	})
}

func runTestCase(t *testing.T, testCase pipelineTest) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)

	cleanUp := setUpGlobalRegistryForTesting(prometheus.NewRegistry())
	defer cleanUp()

	agentRuntimeCtx, cleanUpAgent := newAgentRuntimeContext(t, ctx)
	defer cleanUpAgent()

	cmd := flowmode.Command()
	cmd.SetArgs([]string{
		"run",
		testCase.configFile,
		"--server.http.listen-addr",
		fmt.Sprintf("127.0.0.1:%d", agentRuntimeCtx.agentPort),
		"--storage.path",
		t.TempDir(),
	})

	doneErr := make(chan error)
	go func() { doneErr <- cmd.ExecuteContext(ctx) }()

	assertionsDone := make(chan struct{})
	go func() {
		if testCase.eventuallyAssert != nil {
			require.EventuallyWithT(t, func(t *assert.CollectT) {
				testCase.eventuallyAssert(t, agentRuntimeCtx)
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
		if testCase.requireCleanShutdown {
			t.Fatalf("agent failed to shut down within deadline")
		} else {
			t.Log("agent failed to shut down within deadline")
		}
	case err := <-doneErr:
		verifyDoneError(t, testCase, err)
	}
}

func verifyDoneError(t *testing.T, testCase pipelineTest, err error) {
	if testCase.cmdErrContains != "" {
		require.ErrorContains(t, err, testCase.cmdErrContains, "command must return error containing the string specified in test case")
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
