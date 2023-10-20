package pipelinetests

import (
	"testing"
	"time"

	"github.com/grafana/agent/cmd/internal/pipelinetests/internal/framework"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestPipeline_Prometheus_SelfScrapeAndWrite(topT *testing.T) {
	framework.PipelineTest{
		ConfigFile: "testdata/self_scrape_and_write.river",
		Timeout:    1 * time.Minute, // prometheus tests are slower due to remote_write/wal issues
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
	}.RunTest(topT)
}

func TestPipeline_Prometheus_TargetScrapeAndWrite(topT *testing.T) {
	framework.PipelineTest{
		ConfigFile:       "testdata/target_scrape_and_write.river",
		Timeout:          1 * time.Minute, // prometheus tests are slower due to remote_write/wal issues
		EventuallyAssert: verifyDifferentTypesOfMetricsWithTestTarget(),
	}.RunTest(topT)
}

func TestPipeline_Prometheus_TargetScrapeAndWrite_WithOTELConversion(topT *testing.T) {
	framework.PipelineTest{
		ConfigFile:       "testdata/target_scrape_and_write_otel_conversion.river",
		Timeout:          1 * time.Minute, // prometheus tests are slower due to remote_write/wal issues
		EventuallyAssert: verifyDifferentTypesOfMetricsWithTestTarget(),
	}.RunTest(topT)
}

// verifyDifferentTypesOfMetricsWithTestTarget exposes different metrics using the context.TestTarget and then
// verifies that they all arrived eventually to context.DataSentToProm. This test can be used to verify various
// pipelines that are expected to ship the metrics from a test target to the prometheus remote write endpoint.
func verifyDifferentTypesOfMetricsWithTestTarget() func(t *assert.CollectT, context *framework.RuntimeContext) {
	registered := false
	gauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "test_target_gauge",
		ConstLabels: map[string]string{
			"foo": "bar",
		},
	})
	counter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "test_target_counter",
		ConstLabels: map[string]string{
			"owner": "count_von_count",
		},
	})
	hist := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "test_target_histogram",
		ConstLabels: map[string]string{
			"type": "histogram",
		},
	})

	return func(t *assert.CollectT, context *framework.RuntimeContext) {
		if !registered {
			// Register and set the test gauge only once
			context.TestTarget.Register(gauge)
			gauge.Set(123)

			context.TestTarget.Register(counter)
			counter.Add(321)

			context.TestTarget.Register(hist)
			for i := 0; i < 100; i++ {
				hist.Observe(float64(i))
			}

			registered = true
		}

		assert.NotEmptyf(t, context.DataSentToProm.WritesCount(), "must receive at least one prom write request")
		// Check the gauge as expected value and has const labels
		assert.Equal(t, float64(123), context.DataSentToProm.FindLastSampleMatching(
			"test_target_gauge",
			"foo", "bar",
		))

		// Check the counter labels
		assert.Equal(t, float64(321), context.DataSentToProm.FindLastSampleMatching(
			"test_target_counter",
			"owner", "count_von_count",
		))

		// Check the histogram metrics
		assert.Equal(t, float64(100), context.DataSentToProm.FindLastSampleMatching(
			"test_target_histogram_count",
			"type", "histogram",
		))
		assert.Equal(t, float64(4950), context.DataSentToProm.FindLastSampleMatching(
			"test_target_histogram_sum",
			"type", "histogram",
		))
		assert.Equal(t, float64(11), context.DataSentToProm.FindLastSampleMatching(
			"test_target_histogram_bucket",
			"type", "histogram",
			"le", "10",
		))
	}
}
