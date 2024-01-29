//go:build !windows

package main

import (
	"testing"

	"github.com/grafana/agent/integration-tests/common"
)

func TestRedisMetrics(t *testing.T) {
	var redisMetrics = []string{
		"redis_up",
		"redis_memory_used_bytes",
	}
	common.MimirMetricsTest(t, redisMetrics, []string{}, "redis_metrics")
}
