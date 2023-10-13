package main

import (
	"testing"

	"github.com/grafana/agent/integration-tests/common"
	"github.com/stretchr/testify/assert"
)

const query = "http://localhost:9009/prometheus/api/v1/query?query=avalanche_metric_mmmmm_0_9{job='prometheus.scrape.scrap_prom_metrics9'}"

func TestScrapPromMetrics(t *testing.T) {
	var metricResponse common.MetricResponse
	assert.EventuallyWithT(t, func(c *assert.CollectT) {
		err := common.FetchDataFromURL(query, &metricResponse)
		assert.NoError(c, err)
		if assert.NotEmpty(c, metricResponse.Data.Result) {
			assert.Equal(c, metricResponse.Data.Result[0].Metric.Name, "avalanche_metric_mmmmm_0_9")
			assert.Equal(c, metricResponse.Data.Result[0].Metric.Job, "prometheus.scrape.scrap_prom_metrics9")
		}
	}, common.DefaultTimeout, common.DefaultRetryInterval, "Data did not satisfy the conditions within the time limit")
}
