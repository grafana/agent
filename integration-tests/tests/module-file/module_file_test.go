package main

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/grafana/agent/integration-tests/common"
	"github.com/stretchr/testify/assert"
)

const promURL = "http://localhost:9009/prometheus/api/v1/query?query="
const lokiUrl = "http://localhost:3100/loki/api/v1/query?query={test_name=%22module_file%22}"

func metricQuery(metricName string) string {
	return fmt.Sprintf("%s%s{test_name='module_file'}", promURL, metricName)
}

func TestScrapePromMetricsModuleFile(t *testing.T) {
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
			assert.Equal(c, metricResponse.Data.Result[0].Metric.TestName, "module_file")
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

func TestReadLogFile(t *testing.T) {
	var logResponse common.LogResponse
	assert.EventuallyWithT(t, func(c *assert.CollectT) {
		err := common.FetchDataFromURL(lokiUrl, &logResponse)
		assert.NoError(c, err)
		if assert.NotEmpty(c, logResponse.Data.Result) {
			assert.Equal(c, logResponse.Data.Result[0].Stream["filename"], "logs.txt")
			logs := make([]string, len(logResponse.Data.Result[0].Values))
			for i, valuePair := range logResponse.Data.Result[0].Values {
				logs[i] = valuePair[1]
			}
			assert.Contains(c, logs, "[2023-10-02 14:25:43] INFO: Starting the web application...")
		}
	}, common.DefaultTimeout, common.DefaultRetryInterval, "Data did not satisfy the conditions within the time limit")
}

func assertMetricData(t *testing.T, query, expectedMetric string) {
	var metricResponse common.MetricResponse
	assert.EventuallyWithT(t, func(c *assert.CollectT) {
		err := common.FetchDataFromURL(query, &metricResponse)
		assert.NoError(c, err)
		if assert.NotEmpty(c, metricResponse.Data.Result) {
			assert.Equal(c, metricResponse.Data.Result[0].Metric.Name, expectedMetric)
			assert.Equal(c, metricResponse.Data.Result[0].Metric.TestName, "module_file")
			assert.NotEmpty(c, metricResponse.Data.Result[0].Value.Value)
			assert.Nil(c, metricResponse.Data.Result[0].Histogram)
		}
	}, common.DefaultTimeout, common.DefaultRetryInterval, "Data did not satisfy the conditions within the time limit")
}
