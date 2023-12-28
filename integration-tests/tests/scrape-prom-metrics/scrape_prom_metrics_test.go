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

func metricQuery(metricName string) string {
	return fmt.Sprintf("%s%s{test_name='scrape_prom_metrics'}", promURL, metricName)
}

func TestScrapePromMetrics(t *testing.T) {
	metrics := []string{
		// TODO: better differentiate these metric types?
		"golang_counter",
		"golang_gauge",
		"golang_histogram_bucket",
		"golang_summary",
		"golang_native_histogram",
	}

	for _, metric := range metrics {
		metric := metric
		t.Run(metric, func(t *testing.T) {
			t.Parallel()
			if metric == "golang_native_histogram" {
				assertHistogramData(t, metricQuery(metric), metric)
			} else {
				assertMetricData(t, metricQuery(metric), metric)
			}
		})
	}
}

func assertHistogramData(t *testing.T, query string, expectedMetric string) {
	var metricResponse common.MetricResponse
	assert.EventuallyWithT(t, func(c *assert.CollectT) {
		err := common.FetchDataFromURL(query, &metricResponse)
		assert.NoError(c, err)
		if assert.NotEmpty(c, metricResponse.Data.Result) {
			assert.Equal(c, metricResponse.Data.Result[0].Metric.Name, expectedMetric)
			assert.Equal(c, metricResponse.Data.Result[0].Metric.TestName, "scrape_prom_metrics")
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
	}, common.DefaultTimeout, common.DefaultRetryInterval, "Histogram data did not satisfy the conditions within the time limit")
}

func assertMetricData(t *testing.T, query, expectedMetric string) {
	var metricResponse common.MetricResponse
	assert.EventuallyWithT(t, func(c *assert.CollectT) {
		err := common.FetchDataFromURL(query, &metricResponse)
		assert.NoError(c, err)
		if assert.NotEmpty(c, metricResponse.Data.Result) {
			assert.Equal(c, metricResponse.Data.Result[0].Metric.Name, expectedMetric)
			assert.Equal(c, metricResponse.Data.Result[0].Metric.TestName, "scrape_prom_metrics")
			assert.NotEmpty(c, metricResponse.Data.Result[0].Value.Value)
			assert.Nil(c, metricResponse.Data.Result[0].Histogram)
		}
	}, common.DefaultTimeout, common.DefaultRetryInterval, "Data did not satisfy the conditions within the time limit")
}
