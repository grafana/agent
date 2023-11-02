package spanmetrics_test

import (
	"testing"
	"time"

	"github.com/grafana/agent/component/otelcol/connector/spanmetrics"
	"github.com/grafana/river"
	"github.com/open-telemetry/opentelemetry-collector-contrib/connector/spanmetricsconnector"
	"github.com/stretchr/testify/require"
)

func getStringPtr(str string) *string {
	newStr := str
	return &newStr
}

func TestArguments_UnmarshalRiver(t *testing.T) {
	tests := []struct {
		testName string
		cfg      string
		expected spanmetricsconnector.Config
		errorMsg string
	}{
		{
			testName: "defaultConfigExplicitHistogram",
			cfg: `
			histogram {
				explicit {}
			}

			output {}
			`,
			expected: spanmetricsconnector.Config{
				Dimensions:             []spanmetricsconnector.Dimension{},
				ExcludeDimensions:      nil,
				DimensionsCacheSize:    1000,
				AggregationTemporality: "AGGREGATION_TEMPORALITY_CUMULATIVE",
				Histogram: spanmetricsconnector.HistogramConfig{
					Disable:     false,
					Unit:        0,
					Exponential: nil,
					Explicit: &spanmetricsconnector.ExplicitHistogramConfig{
						Buckets: []time.Duration{
							2 * time.Millisecond,
							4 * time.Millisecond,
							6 * time.Millisecond,
							8 * time.Millisecond,
							10 * time.Millisecond,
							50 * time.Millisecond,
							100 * time.Millisecond,
							200 * time.Millisecond,
							400 * time.Millisecond,
							800 * time.Millisecond,
							1 * time.Second,
							1400 * time.Millisecond,
							2 * time.Second,
							5 * time.Second,
							10 * time.Second,
							15 * time.Second,
						},
					},
				},
				MetricsFlushInterval: 15 * time.Second,
				Namespace:            "",
				Exemplars: spanmetricsconnector.ExemplarsConfig{
					Enabled: false,
				},
			},
		},
		{
			testName: "defaultConfigExponentialHistogram",
			cfg: `
			histogram {
				exponential {}
			}

			output {}
			`,
			expected: spanmetricsconnector.Config{
				Dimensions:             []spanmetricsconnector.Dimension{},
				DimensionsCacheSize:    1000,
				ExcludeDimensions:      nil,
				AggregationTemporality: "AGGREGATION_TEMPORALITY_CUMULATIVE",
				Histogram: spanmetricsconnector.HistogramConfig{
					Disable:     false,
					Unit:        0,
					Exponential: &spanmetricsconnector.ExponentialHistogramConfig{MaxSize: 160},
					Explicit:    nil,
				},
				MetricsFlushInterval: 15 * time.Second,
				Namespace:            "",
			},
		},
		{
			testName: "explicitConfig",
			cfg: `
			dimension {
				name = "http.status_code"
			}
			dimension {
				name = "http.method"
				default = "GET"
			}
			exclude_dimensions = ["test_exclude_dim1", "test_exclude_dim2"]
			dimensions_cache_size = 333
			aggregation_temporality = "DELTA"
			histogram {
				disable = true
				unit = "s"
				explicit {
					buckets = ["333ms", "777s", "999h"]
				}
			}
			metrics_flush_interval = "33s"
			namespace = "test.namespace"
			exemplars {
				enabled = true
			}

			output {}
			`,
			expected: spanmetricsconnector.Config{
				Dimensions: []spanmetricsconnector.Dimension{
					{Name: "http.status_code", Default: nil},
					{Name: "http.method", Default: getStringPtr("GET")},
				},
				ExcludeDimensions:      []string{"test_exclude_dim1", "test_exclude_dim2"},
				DimensionsCacheSize:    333,
				AggregationTemporality: "AGGREGATION_TEMPORALITY_DELTA",
				Histogram: spanmetricsconnector.HistogramConfig{
					Disable:     true,
					Unit:        1,
					Exponential: nil,
					Explicit: &spanmetricsconnector.ExplicitHistogramConfig{
						Buckets: []time.Duration{
							333 * time.Millisecond,
							777 * time.Second,
							999 * time.Hour,
						},
					},
				},
				MetricsFlushInterval: 33 * time.Second,
				Namespace:            "test.namespace",
				Exemplars: spanmetricsconnector.ExemplarsConfig{
					Enabled: true,
				},
			},
		},
		{
			testName: "exponentialHistogramMs",
			cfg: `
			histogram {
				unit = "ms"
				exponential {
					max_size = 123
				}
			}

			output {}
			`,
			expected: spanmetricsconnector.Config{
				Dimensions:             []spanmetricsconnector.Dimension{},
				DimensionsCacheSize:    1000,
				AggregationTemporality: "AGGREGATION_TEMPORALITY_CUMULATIVE",
				Histogram: spanmetricsconnector.HistogramConfig{
					Unit:        0,
					Exponential: &spanmetricsconnector.ExponentialHistogramConfig{MaxSize: 123},
					Explicit:    nil,
				},
				MetricsFlushInterval: 15 * time.Second,
				Namespace:            "",
			},
		},
		{
			testName: "invalidAggregationTemporality",
			cfg: `
			aggregation_temporality = "badVal"

			histogram {
				explicit {}
			}

			output {}
			`,
			errorMsg: `invalid aggregation_temporality: badVal`,
		},
		{
			testName: "invalidDimensionCache1",
			cfg: `
			dimensions_cache_size = -1

			histogram {
				explicit {}
			}

			output {}
			`,
			errorMsg: `invalid cache size: -1, the maximum number of the items in the cache should be positive`,
		},
		{
			testName: "invalidDimensionCache2",
			cfg: `
			dimensions_cache_size = 0

			histogram {
				explicit {}
			}

			output {}
			`,
			errorMsg: `invalid cache size: 0, the maximum number of the items in the cache should be positive`,
		},
		{
			testName: "invalidMetricsFlushInterval1",
			cfg: `
			metrics_flush_interval = "0s"

			histogram {
				explicit {}
			}

			output {}
			`,
			errorMsg: `metrics_flush_interval must be greater than 0`,
		},
		{
			testName: "invalidMetricsFlushInterval2",
			cfg: `
			metrics_flush_interval = "-1s"

			histogram {
				explicit {}
			}

			output {}
			`,
			errorMsg: `metrics_flush_interval must be greater than 0`,
		},
		{
			testName: "invalidDuplicateHistogramConfig",
			cfg: `
			histogram {
				explicit {
					buckets = ["333ms", "777s", "999h"]
				}
				exponential {
					max_size = 123
				}
			}

			output {}
			`,
			errorMsg: `only one of exponential or explicit histogram configuration can be specified`,
		},
		{
			testName: "invalidHistogramExplicitUnit",
			cfg: `
			histogram {
				explicit {
					buckets = ["333fakeunit", "777s", "999h"]
				}
			}

			output {}
			`,
			errorMsg: `4:17: "333fakeunit" time: unknown unit "fakeunit" in duration "333fakeunit"`,
		},
		{
			testName: "invalidHistogramExponentialSize",
			cfg: `
			histogram {
				exponential {
					max_size = -1
				}
			}

			output {}
			`,
			errorMsg: `max_size must be greater than 0`,
		},
		{
			testName: "invalidHistogramUnit",
			cfg: `
			histogram {
				unit = "badUnit"
				explicit {}
			}

			output {}
			`,
			errorMsg: `unknown unit "badUnit", allowed units are "ms" and "s"`,
		},
		{
			testName: "invalidHistogramNoConfig",
			cfg: `
			histogram {}

			output {}
			`,
			errorMsg: `either exponential or explicit histogram configuration must be specified`,
		},
		{
			testName: "invalidNoHistogram",
			cfg: `
			output {}
			`,
			errorMsg: `missing required block "histogram"`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			var args spanmetrics.Arguments
			err := river.Unmarshal([]byte(tc.cfg), &args)
			if tc.errorMsg != "" {
				require.ErrorContains(t, err, tc.errorMsg)
				return
			}

			require.NoError(t, err)

			actualPtr, err := args.Convert()
			require.NoError(t, err)

			actual := actualPtr.(*spanmetricsconnector.Config)

			require.NoError(t, actual.Validate())

			require.Equal(t, tc.expected, *actual)
		})
	}
}
