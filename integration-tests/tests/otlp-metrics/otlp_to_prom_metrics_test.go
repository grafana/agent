//go:build !windows

package main

import (
	"testing"
)

func TestOTLPToPromMetrics(t *testing.T) {
	const testName = "otlp_to_prom_metrics"
	tests := []struct {
		metric string
	}{
		{"example_counter_total"},
		{"example_float_counter_total"},
		{"example_updowncounter"},
		{"example_float_updowncounter"},
		{"example_histogram_bucket"},
		{"example_float_histogram_bucket"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.metric, func(t *testing.T) {
			t.Parallel()
			assertMetricData(t, metricQuery(tt.metric, testName), tt.metric, testName)
		})
	}

	histogramTests := []string{
		"example_exponential_histogram",
		"example_exponential_float_histogram",
	}

	for _, metric := range histogramTests {
		metric := metric
		t.Run(metric, func(t *testing.T) {
			t.Parallel()
			assertHistogramData(t, metricQuery(metric, testName), metric, testName)
		})
	}
}
