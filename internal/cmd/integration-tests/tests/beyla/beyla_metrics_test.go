//go:build !windows

package main

import (
	"net/http"
	"testing"
	"time"

	"github.com/grafana/agent/internal/cmd/integration-tests/common"
)

func TestBeylaMetrics(t *testing.T) {
	var beylaMetrics = []string{
		"http_server_request_duration_seconds_count",
	}
	http.Get("http://localhost:9001/metrics")
	time.Sleep(2 * time.Second)
	common.MimirMetricsTest(t, beylaMetrics, []string{}, "beyla_metrics")
}
