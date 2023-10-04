package main

import (
	"encoding/json"
	"testing"

	"github.com/grafana/agent/integration-tests/common"
	"github.com/stretchr/testify/assert"
)

const query = "http://localhost:3100/loki/api/v1/series"

type data struct {
	Filename string `json:"filename"`
}

type LogsResponse struct {
	Status string `json:"status"`
	Data   []data `json:"data"`
}

func TestReadLogFile(t *testing.T) {
	var logsResponse LogsResponse
	assert.EventuallyWithT(t, func(c *assert.CollectT) {
		err := common.FetchDataFromURL(query, &logsResponse)
		assert.NoError(c, err)
		assert.NotEmpty(c, logsResponse.Data)
		assert.Equal(c, logsResponse.Data[0].Filename, "logs.txt")
	}, common.DefaultTimeout, common.DefaultRetryInterval, "Data did not satisfy the conditions within the time limit")
}

func (m *LogsResponse) Unmarshal(data []byte) error {
	return json.Unmarshal(data, m)
}
