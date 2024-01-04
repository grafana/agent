//go:build !windows

package main

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/grafana/agent/integration-tests/common"
	"github.com/stretchr/testify/assert"
)

const promURL = "http://localhost:9009/prometheus/api/v1/query?query="

func metricQuery(metricName string, testName string) string {
	return fmt.Sprintf("%s%s{test_name='%s'}", promURL, metricName, testName)
}

func TestOTLPMetrics(t *testing.T) {
	const testName = "otlp_metrics"
	tests := []struct {
		metric string
	}{
		// TODO: better differentiate these metric types?
		{"example_counter"},
		{"example_float_counter"},
		{"example_updowncounter"},
		{"example_float_updowncounter"},
		{"example_histogram_bucket"},
		{"example_float_histogram_bucket"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.metric, func(t *testing.T) {
			t.Parallel()
			assertMetricData(t, metricQuery(tt.metric, testName), tt.metric, testName)
		})
	}

	histogramTests := []string{
		"example_exponential_histogram",
		"example_exponential_float_histogram",
	}

	for _, metric := range histogramTests {
		metric := metric
		t.Run(metric, func(t *testing.T) {
			t.Parallel()
			assertHistogramData(t, metricQuery(metric, testName), metric, testName)
		})
	}
}

func assertHistogramData(t *testing.T, query string, expectedMetric string, testName string) {
	var metricResponse common.MetricResponse
	assert.EventuallyWithT(t, func(c *assert.CollectT) {
		err := common.FetchDataFromURL(query, &metricResponse)
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
					sum, _ := strconv.Atoi(histogram.Data.Sum)
					assert.Greater(c, sum, 10, "Sum should be at some point greater than 10.")
				}
				assert.NotEmpty(c, histogram.Data.Buckets)
				assert.Nil(c, metricResponse.Data.Result[0].Value)
			}
		}
	}, common.DefaultTimeout, common.DefaultRetryInterval, "Histogram data did not satisfy the conditions within the time limit")
}

func assertMetricData(t *testing.T, query, expectedMetric string, testName string) {
	var metricResponse common.MetricResponse
	assert.EventuallyWithT(t, func(c *assert.CollectT) {
		err := common.FetchDataFromURL(query, &metricResponse)
		assert.NoError(c, err)
		if assert.NotEmpty(c, metricResponse.Data.Result) {
			assert.Equal(c, metricResponse.Data.Result[0].Metric.Name, expectedMetric)
			assert.Equal(c, metricResponse.Data.Result[0].Metric.TestName, testName)
			assert.NotEmpty(c, metricResponse.Data.Result[0].Value.Value)
			assert.Nil(c, metricResponse.Data.Result[0].Histogram)
		}
	}, common.DefaultTimeout, common.DefaultRetryInterval, "Data did not satisfy the conditions within the time limit")
}
