package main

import (
	"testing"

	"github.com/grafana/agent/integration-tests/common"
)

func TestScrapePromMetricsModuleGit(t *testing.T) {
	common.MimirMetricsTest(t, common.PromDefaultMetrics, common.PromDefaultHistogramMetric, "scrape_prom_metrics_module_git")
}
