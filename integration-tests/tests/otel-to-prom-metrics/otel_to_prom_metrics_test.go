package main

import (
	"testing"

	"github.com/grafana/agent/integration-tests/common"
	"github.com/stretchr/testify/assert"
)

const query = "http://localhost:9009/prometheus/api/v1/query?query=span_metrics_duration_bucket{job='otel_to_prom_metrics'}"

func TestOtelToPromMetrics(t *testing.T) {
	var metricResponse common.MetricResponse
	assert.EventuallyWithT(t, func(c *assert.CollectT) {
		err := common.FetchDataFromURL(query, &metricResponse)
		assert.NoError(c, err)
		if assert.NotEmpty(c, metricResponse.Data.Result) {
			assert.Equal(c, metricResponse.Data.Result[0].Metric.Name, "span_metrics_duration_bucket")
			assert.Equal(c, metricResponse.Data.Result[0].Metric.Job, "otel_to_prom_metrics")
		}
	}, common.DefaultTimeout, common.DefaultRetryInterval, "Data did not satisfy the conditions within the time limit")
}
