package grok_exporter //nolint:golint

import (
	v3 "github.com/fstab/grok_exporter/config/v3"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateCounterMetric(t *testing.T) {
	config := Config{
		GrokConfig: v3.Config{
			GrokPatterns: []string{"HOUR (?:2[0123]|[01]?[0-9])"},
			AllMetrics: v3.MetricsConfig{{
				Type:          "counter",
				Name:          "test_counter",
				Help:          "test_counter",
				PathsAndGlobs: v3.PathsAndGlobs{},
				Match:         "%{HOUR}",
			}},
		},
	}

	metrics, err := config.CreateMetrics()

	assert.NoError(t, err)
	if _, ok := metrics[0].Collector().(prometheus.Counter); !ok {
		assert.Fail(t, "Failed to create counter metric")
	}
	assert.Equal(t, config.GrokConfig.AllMetrics[0].Name, metrics[0].Name())
}

func TestCreateGaugeMetric(t *testing.T) {
	config := Config{
		GrokConfig: v3.Config{
			GrokPatterns: []string{"HOUR (?:2[0123]|[01]?[0-9])"},
			AllMetrics: v3.MetricsConfig{{
				Type:          "gauge",
				Name:          "test_counter",
				Help:          "test_counter",
				PathsAndGlobs: v3.PathsAndGlobs{},
				Match:         "%{HOUR}",
			}},
		},
	}

	metrics, err := config.CreateMetrics()

	assert.NoError(t, err)
	if _, ok := metrics[0].Collector().(prometheus.Gauge); !ok {
		assert.Fail(t, "Failed to create gauge metric")
	}
	assert.Equal(t, config.GrokConfig.AllMetrics[0].Name, metrics[0].Name())
}

func TestCreateHistogramMetric(t *testing.T) {
	config := Config{
		GrokConfig: v3.Config{
			GrokPatterns: []string{"HOUR (?:2[0123]|[01]?[0-9])"},
			AllMetrics: v3.MetricsConfig{{
				Type:          "histogram",
				Name:          "test_counter",
				Help:          "test_counter",
				PathsAndGlobs: v3.PathsAndGlobs{},
				Match:         "%{HOUR}",
			}},
		},
	}

	metrics, err := config.CreateMetrics()

	assert.NoError(t, err)
	if _, ok := metrics[0].Collector().(prometheus.Histogram); !ok {
		assert.Fail(t, "Failed to create histogram metric")
	}
	assert.Equal(t, config.GrokConfig.AllMetrics[0].Name, metrics[0].Name())
}

func TestCreateSummaryMetric(t *testing.T) {
	config := Config{
		GrokConfig: v3.Config{
			GrokPatterns: []string{"HOUR (?:2[0123]|[01]?[0-9])"},
			AllMetrics: v3.MetricsConfig{{
				Type:          "summary",
				Name:          "test_counter",
				Help:          "test_counter",
				PathsAndGlobs: v3.PathsAndGlobs{},
				Match:         "%{HOUR}",
			}},
		},
	}

	metrics, err := config.CreateMetrics()

	assert.NoError(t, err)
	if _, ok := metrics[0].Collector().(prometheus.Summary); !ok {
		assert.Fail(t, "Failed to create summary metric")
	}
	assert.Equal(t, config.GrokConfig.AllMetrics[0].Name, metrics[0].Name())
}

func TestCreateInvalidMetric(t *testing.T) {
	config := Config{
		GrokConfig: v3.Config{
			GrokPatterns: []string{"HOUR (?:2[0123]|[01]?[0-9])"},
			AllMetrics: v3.MetricsConfig{{
				Type:          "invalid",
				Name:          "test_counter",
				Help:          "test_counter",
				PathsAndGlobs: v3.PathsAndGlobs{},
				Match:         "%{HOUR}",
			}},
		},
	}

	metrics, err := config.CreateMetrics()

	assert.Error(t, err)
	assert.Equal(t, "failed to initialize metrics: Metric type invalid is not supported", err.Error())
	assert.Nil(t, metrics)
}

func TestCreateMetricInvalidMatch(t *testing.T) {
	config := Config{
		GrokConfig: v3.Config{
			GrokPatterns: []string{"HOUR (?:2[0123]|[01]?[0-9])"},
			AllMetrics: v3.MetricsConfig{{
				Type:          "counter",
				Name:          "test_counter",
				Help:          "test_counter",
				PathsAndGlobs: v3.PathsAndGlobs{},
				Match:         "%{INVALID}",
			}},
		},
	}

	metrics, err := config.CreateMetrics()

	assert.Error(t, err)
	assert.Equal(t, "failed to initialize metric test_counter: Pattern %{INVALID} not defined.", err.Error())
	assert.Nil(t, metrics)
}

func TestCreateMetricInvalidPatternDirectory(t *testing.T) {
	config := Config{
		GrokConfig: v3.Config{
			Imports: []v3.ImportConfig{{
				Type: "grok_patterns",
				Dir:  "/invalid",
			}},
			GrokPatterns: []string{"HOUR (?:2[0123]|[01]?[0-9])"},
			AllMetrics: v3.MetricsConfig{{
				Type:          "counter",
				Name:          "test_counter",
				Help:          "test_counter",
				PathsAndGlobs: v3.PathsAndGlobs{},
				Match:         "%{INVALID}",
			}},
		},
	}

	metrics, err := config.CreateMetrics()

	assert.Error(t, err)
	assert.Equal(t, "failed to initialize patterns: failed to read pattern directory /invalid: open /invalid: no such file or directory", err.Error())
	assert.Nil(t, metrics)
}

func TestCreateMetricInvalidPatternFile(t *testing.T) {
	config := Config{
		GrokConfig: v3.Config{
			Imports: []v3.ImportConfig{{
				Type: "grok_patterns",
				File: "invalid.txt",
			}},
			GrokPatterns: []string{"HOUR (?:2[0123]|[01]?[0-9])"},
			AllMetrics: v3.MetricsConfig{{
				Type:          "counter",
				Name:          "test_counter",
				Help:          "test_counter",
				PathsAndGlobs: v3.PathsAndGlobs{},
				Match:         "%{INVALID}",
			}},
		},
	}

	metrics, err := config.CreateMetrics()

	assert.Error(t, err)
	assert.Equal(t, "failed to initialize patterns: invalid.txt: no such file", err.Error())
	assert.Nil(t, metrics)
}
