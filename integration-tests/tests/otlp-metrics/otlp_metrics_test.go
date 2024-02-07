//go:build !windows

package main

import (
	"testing"

	"github.com/grafana/agent/integration-tests/common"
)

func TestOTLPMetrics(t *testing.T) {
	common.MimirMetricsTest(t, common.OtelDefaultMetrics, common.OtelDefaultHistogramMetrics, "otlp_metrics")
}
