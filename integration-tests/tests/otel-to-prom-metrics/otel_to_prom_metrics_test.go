package main

import (
	"encoding/json"
	"testing"

	"github.com/grafana/agent/integration-tests/common"
	"github.com/stretchr/testify/assert"
)

const query = "http://localhost:9009/prometheus/api/v1/label/__name__/values"

type MetricResponse struct {
	Status string   `json:"status"`
	Data   []string `json:"data"`
}

func TestOtelToPromMetrics(t *testing.T) {
	var metricResponse MetricResponse
	assert.EventuallyWithT(t, func(c *assert.CollectT) {
		err := common.FetchDataFromURL(query, &metricResponse)
		assert.NoError(c, err)
		assert.Contains(c, metricResponse.Data, "span_metrics_duration_bucket")
	}, common.DefaultTimeout, common.DefaultRetryInterval, "Data did not satisfy the conditions within the time limit")
}

func (m *MetricResponse) Unmarshal(data []byte) error {
	return json.Unmarshal(data, m)
}
