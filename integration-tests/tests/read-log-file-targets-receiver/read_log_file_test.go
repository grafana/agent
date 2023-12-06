package main

import (
	"testing"

	"github.com/grafana/agent/integration-tests/common"
	"github.com/stretchr/testify/assert"
)

func TestReadLogFile(t *testing.T) {
	const queryLogFile = "http://localhost:3100/loki/api/v1/query?query={test_name=%22read_log_file_targets_receiver%22,%20filename=~%22.*example.log%22}"
	expectedLine := "[2023-10-02 14:25:43] INFO: source=example.log Starting the web application..."
	testQuery(t, queryLogFile, expectedLine)
}

func TestReadTxtFile(t *testing.T) {
	const queryTxtFile = "http://localhost:3100/loki/api/v1/query?query={test_name=%22read_log_file_targets_receiver%22,%20filename=~%22.*logs.txt%22}"
	expectedLine := "[2023-10-02 14:25:43] INFO: Starting the web application..."
	testQuery(t, queryTxtFile, expectedLine)
}

func testQuery(t *testing.T, query string, requiredLogLine string) {
	var logResponse common.LogResponse
	assert.EventuallyWithT(t, func(c *assert.CollectT) {
		err := common.FetchDataFromURL(query, &logResponse)
		assert.NoError(c, err)
		if assert.NotEmpty(c, logResponse.Data.Result) {
			logs := make([]string, len(logResponse.Data.Result[0].Values))
			for i, valuePair := range logResponse.Data.Result[0].Values {
				logs[i] = valuePair[1]
			}
			assert.Contains(c, logs, requiredLogLine)
		}
	}, common.DefaultTimeout, common.DefaultRetryInterval, "Data did not satisfy the conditions within the time limit")
}
