package pipelinetests

import (
	"github.com/grafana/agent/cmd/internal/pipelinetests/internal/framework"
	"github.com/stretchr/testify/assert"
	"testing"
)

/*
*
//TODO(thampiotr):
- Provide fake scrape target that can be scraped?
- Think how to make this low-code and easier to use
- Make a test with logging pipeline
- Make a test with OTEL pipeline
- Make a test with loki.process
- Make a test with relabel rules
*
*/
func TestPipeline_WithEmptyConfig(t *testing.T) {
	framework.RunPipelineTest(t, framework.PipelineTest{
		ConfigFile:           "testdata/empty.river",
		RequireCleanShutdown: true,
	})
}

func TestPipeline_FileNotExists(t *testing.T) {
	framework.RunPipelineTest(t, framework.PipelineTest{
		ConfigFile:           "does_not_exist.river",
		CmdErrContains:       "does_not_exist.river: no such file or directory",
		RequireCleanShutdown: true,
	})
}

func TestPipeline_FileInvalid(t *testing.T) {
	framework.RunPipelineTest(t, framework.PipelineTest{
		ConfigFile:           "testdata/invalid.river",
		CmdErrContains:       "could not perform the initial load successfully",
		RequireCleanShutdown: true,
	})
}

func TestPipeline_Prometheus_SelfScrapeAndWrite(topT *testing.T) {
	framework.RunPipelineTest(topT, framework.PipelineTest{
		ConfigFile: "testdata/scrape_and_write.river",
		EventuallyAssert: func(t *assert.CollectT, context *framework.RuntimeContext) {
			assert.NotEmptyf(t, context.DataSentToProm.WritesCount(), "must receive at least one prom write request")
			// One target expected
			assert.Equal(t, float64(1), context.DataSentToProm.FindLastSampleMatching("agent_prometheus_scrape_targets_gauge"))
			// Fanned out at least one target
			assert.GreaterOrEqual(t, context.DataSentToProm.FindLastSampleMatching(
				"agent_prometheus_fanout_latency_count",
				"component_id",
				"prometheus.scrape.agent_self",
			), float64(1))

			// Received at least `count` samples
			count := 1000
			assert.Greater(t, context.DataSentToProm.FindLastSampleMatching(
				"agent_prometheus_forwarded_samples_total",
				"component_id",
				"prometheus.scrape.agent_self",
			), float64(count))
			assert.Greater(t, context.DataSentToProm.FindLastSampleMatching(
				"agent_wal_samples_appended_total",
				"component_id",
				"prometheus.remote_write.default",
			), float64(count))

			// At least 100 active series should be present
			assert.Greater(t, context.DataSentToProm.FindLastSampleMatching(
				"agent_wal_storage_active_series",
				"component_id",
				"prometheus.remote_write.default",
			), float64(100))
		},
	})
}
