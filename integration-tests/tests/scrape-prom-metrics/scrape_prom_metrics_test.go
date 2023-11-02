package main

import (
	"fmt"
	"testing"

	"github.com/grafana/agent/integration-tests/common"
	"github.com/stretchr/testify/assert"
)

const promURL = "http://localhost:9009/prometheus/api/v1/query?query="

func metricQuery(metricName, testName string) string {
	return fmt.Sprintf("%s%s{test_name='%s'}", promURL, metricName, testName)
}

func TestScrapePromMetrics(t *testing.T) {
	tests := []struct {
		query  string
		metric string
	}{
		{metricQuery("golang_counter", "scrape_prom_metrics"), "golang_counter"},
		{metricQuery("golang_gauge", "scrape_prom_metrics"), "golang_gauge"},
		{metricQuery("golang_histogram_bucket", "scrape_prom_metrics"), "golang_histogram_bucket"},
		{metricQuery("golang_summary", "scrape_prom_metrics"), "golang_summary"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.metric, func(t *testing.T) {
			t.Parallel()
			assertMetricData(t, tt.query, tt.metric)
		})
	}
	t.Run("golang_native_histogram", func(t *testing.T) {
		t.Parallel()
		assertHistogramData(t, metricQuery("golang_native_histogram", "scrape_prom_metrics"), "golang_native_histogram")
	})
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
				assert.NotEmpty(c, histogram.Data.Count)
				assert.NotEmpty(c, histogram.Data.Sum)
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
