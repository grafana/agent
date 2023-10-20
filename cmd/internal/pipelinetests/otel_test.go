package pipelinetests

import (
	"testing"

	"github.com/grafana/agent/cmd/internal/pipelinetests/internal/framework"
	"github.com/stretchr/testify/assert"
)

func TestPipeline_OTEL_TestScrapeAndWrite(topT *testing.T) {
	framework.RunPipelineTest(topT, framework.PipelineTest{
		ConfigFile: "testdata/self_scrape_and_write.river",
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
				"component_id", "prometheus.remote_write.default",
				"job", "agent",
			), float64(100))
		},
	})
}
