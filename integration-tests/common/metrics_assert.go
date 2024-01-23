package common

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

const promURL = "http://localhost:9009/prometheus/api/v1/query?query="

// Default metrics list according to what the prom-gen app is generating.
var PromDefaultMetrics = []string{
	"golang_counter",
	"golang_gauge",
	"golang_histogram_bucket",
	"golang_summary",
}

// Default histogram metrics list according to what the prom-gen app is generating.
var PromDefaultHistogramMetric = []string{
	"golang_native_histogram",
}

// Default metrics list according to what the otel-metrics-gen app is generating.
var OtelDefaultMetrics = []string{
	"example_counter",
	"example_float_counter",
	"example_updowncounter",
	"example_float_updowncounter",
	"example_histogram_bucket",
	"example_float_histogram_bucket",
}

// Default histogram metrics list according to what the otel-metrics-gen app is generating.
var OtelDefaultHistogramMetrics = []string{
	"example_exponential_histogram",
	"example_exponential_float_histogram",
}

// MetricQuery returns a formatted Prometheus metric query with a given metricName and a given label.
func MetricQuery(metricName string, testName string) string {
	return fmt.Sprintf("%s%s{test_name='%s'}", promURL, metricName, testName)
}

// MimirMetricsTest checks that all given metrics are stored in Mimir.
func MimirMetricsTest(t *testing.T, metrics []string, histogramMetrics []string, testName string) {
	for _, metric := range metrics {
		metric := metric
		t.Run(metric, func(t *testing.T) {
			t.Parallel()
			AssertMetricData(t, MetricQuery(metric, testName), metric, testName)
		})
	}
	for _, metric := range histogramMetrics {
		metric := metric
		t.Run(metric, func(t *testing.T) {
			t.Parallel()
			AssertHistogramData(t, MetricQuery(metric, testName), metric, testName)
		})
	}
}

// AssertHistogramData performs a Prometheus query and expect the result to eventually contain the expected histogram.
// The count and sum metrics should be greater than 10 before the timeout triggers.
func AssertHistogramData(t *testing.T, query string, expectedMetric string, testName string) {
	var metricResponse MetricResponse
	assert.EventuallyWithT(t, func(c *assert.CollectT) {
		err := FetchDataFromURL(query, &metricResponse)
		assert.NoError(c, err)
		if assert.NotEmpty(c, metricResponse.Data.Result) {
			assert.Equal(c, metricResponse.Data.Result[0].Metric.Name, expectedMetric)
			assert.Equal(c, metricResponse.Data.Result[0].Metric.TestName, testName)
			if assert.NotNil(c, metricResponse.Data.Result[0].Histogram) {
				histogram := metricResponse.Data.Result[0].Histogram
				if assert.NotEmpty(c, histogram.Data.Count) {
					count, _ := strconv.Atoi(histogram.Data.Count)
					assert.Greater(c, count, 10, "Count should be at some point greater than 10.")
				}
				if assert.NotEmpty(c, histogram.Data.Sum) {
					sum, _ := strconv.ParseFloat(histogram.Data.Sum, 64)
					assert.Greater(c, sum, 10., "Sum should be at some point greater than 10.")
				}
				assert.NotEmpty(c, histogram.Data.Buckets)
				assert.Nil(c, metricResponse.Data.Result[0].Value)
			}
		}
	}, DefaultTimeout, DefaultRetryInterval, "Histogram data did not satisfy the conditions within the time limit")
}

// AssertMetricData performs a Prometheus query and expect the result to eventually contain the expected metric.
func AssertMetricData(t *testing.T, query, expectedMetric string, testName string) {
	var metricResponse MetricResponse
	assert.EventuallyWithT(t, func(c *assert.CollectT) {
		err := FetchDataFromURL(query, &metricResponse)
		assert.NoError(c, err)
		if assert.NotEmpty(c, metricResponse.Data.Result) {
			assert.Equal(c, metricResponse.Data.Result[0].Metric.Name, expectedMetric)
			assert.Equal(c, metricResponse.Data.Result[0].Metric.TestName, testName)
			fmt.Println(metricResponse.Data.Result[0])
			fmt.Println(metricResponse.Data.Result[0].Value)
			assert.NotEmpty(c, metricResponse.Data.Result[0].Value.Value)
			assert.Nil(c, metricResponse.Data.Result[0].Histogram)
		}
	}, DefaultTimeout, DefaultRetryInterval, "Data did not satisfy the conditions within the time limit")
}
