package main

import (
	"testing"

	"github.com/grafana/agent/integration-tests/common"
	"github.com/stretchr/testify/assert"
)

const query = "http://localhost:3100/loki/api/v1/series"

func TestReadLogFile(t *testing.T) {
	var logResponse common.LogResponse
	assert.EventuallyWithT(t, func(c *assert.CollectT) {
		err := common.FetchDataFromURL(query, &logResponse)
		assert.NoError(c, err)
		assert.NotEmpty(c, logResponse.Data)
		assert.Equal(c, logResponse.Data[0].Filename, "logs.txt")
	}, common.DefaultTimeout, common.DefaultRetryInterval, "Data did not satisfy the conditions within the time limit")
}
