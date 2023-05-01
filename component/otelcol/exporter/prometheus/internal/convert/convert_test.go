package convert_test

import (
	"context"
	"testing"

	"github.com/grafana/agent/pkg/util"

	"github.com/grafana/agent/component/otelcol/exporter/prometheus/internal/convert"
	"github.com/grafana/agent/pkg/util/testappender"
	"github.com/prometheus/prometheus/storage"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

func TestConverter(t *testing.T) {
	tt := []struct {
		name   string
		input  string
		expect string

		showTimestamps    bool
		includeTargetInfo bool
		includeScopeInfo  bool
	}{
		{
			name: "Gauge",
			input: `{
				"resource_metrics": [{
					"scope_metrics": [{
						"metrics": [{
							"name": "test_metric_seconds",
							"gauge": {
								"data_points": [{
									"start_time_unix_nano": 1000000000,
									"time_unix_nano": 1000000000,
									"as_double": 1234.56
								}]
							}
						}]
					}]
				}]
			}`,
			expect: `
				# TYPE test_metric_seconds gauge
				test_metric_seconds 1234.56
			`,
		},
		{
			name: "Monotonic sum",
			input: `{
				"resource_metrics": [{
					"scope_metrics": [{
						"metrics": [{
							"name": "test_metric_seconds_total",
							"sum": {
								"aggregation_temporality": 2,
								"is_monotonic": true,
								"data_points": [{
									"start_time_unix_nano": 1000000000,
									"time_unix_nano": 1000000000,
									"as_double": 15
								}]
							}
						}]
					}]
				}]
			}`,
			expect: `
				# TYPE test_metric_seconds counter
				test_metric_seconds_total 15.0
			`,
		},
		{
			name: "Non-monotonic sum",
			input: `{
				"resource_metrics": [{
					"scope_metrics": [{
						"metrics": [{
							"name": "test_metric_seconds",
							"sum": {
								"aggregation_temporality": 2,
								"is_monotonic": false,
								"data_points": [{
									"start_time_unix_nano": 1000000000,
									"time_unix_nano": 1000000000,
									"as_double": 15
								}]
							}
						}]
					}]
				}]
			}`,
			expect: `
				# TYPE test_metric_seconds gauge
				test_metric_seconds 15.0
			`,
		},
		{
			name: "Histogram",
			input: `{
				"resource_metrics": [{
					"scope_metrics": [{
						"metrics": [{
							"name": "test_metric_seconds",
							"histogram": {
								"aggregation_temporality": 2,
								"data_points": [{
									"start_time_unix_nano": 1000000000,
									"time_unix_nano": 1000000000,
									"count": 333,
									"sum": 100,
									"bucket_counts": [0, 111, 0, 222],
									"explicit_bounds": [0.25, 0.5, 0.75, 1.0]
								}]
							}
						}]
					}]
				}]
			}`,
			expect: `
				# TYPE test_metric_seconds histogram
				test_metric_seconds_bucket{le="0.25"} 0
				test_metric_seconds_bucket{le="0.5"} 111
				test_metric_seconds_bucket{le="0.75"} 0
				test_metric_seconds_bucket{le="1.0"} 222
				test_metric_seconds_bucket{le="+Inf"} 333
				test_metric_seconds_sum 100.0
				test_metric_seconds_count 333
			`,
		},
		{
			name: "Summary",
			input: `{
				"resource_metrics": [{
					"scope_metrics": [{
						"metrics": [{
							"name": "test_metric_seconds",
							"summary": {
								"data_points": [{
									"start_time_unix_nano": 1000000000,
									"time_unix_nano": 1000000000,
									"count": 333,
									"sum": 100,
									"quantile_values": [
										{ "quantile": 0, "value": 100 },
										{ "quantile": 0.25, "value": 200 },
										{ "quantile": 0.5, "value": 300 },
										{ "quantile": 0.75, "value": 400 },
										{ "quantile": 1, "value": 500 }
									]
								}]
							}
						}]
					}]
				}]
			}`,
			expect: `
				# TYPE test_metric_seconds summary
				test_metric_seconds{quantile="0.0"} 100.0
				test_metric_seconds{quantile="0.25"} 200.0
				test_metric_seconds{quantile="0.5"} 300.0
				test_metric_seconds{quantile="0.75"} 400.0
				test_metric_seconds{quantile="1.0"} 500.0
				test_metric_seconds_sum 100.0
				test_metric_seconds_count 333
			`,
		},
		{
			name: "Timestamps",
			input: `{
				"resource_metrics": [{
					"scope_metrics": [{
						"metrics": [{
							"name": "test_metric_seconds",
							"gauge": {
								"data_points": [{
									"start_time_unix_nano": 1000000000,
									"time_unix_nano": 1000000000,
									"as_double": 1234.56
								}]
							}
						}]
					}]
				}]
			}`,
			showTimestamps: true,
			expect: `
				# TYPE test_metric_seconds gauge
				test_metric_seconds 1234.56 1.0
			`,
		},
		{
			name: "Labels from resource attributes",
			input: `{
				"resource_metrics": [{
					"resource": {
						"attributes": [{
							"key": "service.name",
							"value": { "stringValue": "myservice" }
						}, {
							"key": "service.instance.id",
							"value": { "stringValue": "instance" }
						}, {
							"key": "do_not_display",
							"value": { "stringValue": "test" }
						}]
					},
					"scope_metrics": [{
						"metrics": [{
							"name": "test_metric_seconds",
							"gauge": {
								"data_points": [{
									"as_double": 1234.56
								}]
							}
						}]
					}]
				}]
			}`,
			expect: `
				# TYPE test_metric_seconds gauge
				test_metric_seconds{instance="instance",job="myservice"} 1234.56
			`,
		},
		{
			name: "Labels from scope name and version",
			input: `{
				"resource_metrics": [{
					"scope_metrics": [{
						"scope": {
							"name": "a-name",
							"version": "a-version",
							"attributes": [{
								"key": "something.extra",
								"value": { "stringValue": "zzz-extra-value" }
							}]
						},
						"metrics": [{
							"name": "test_metric_seconds",
							"gauge": {
								"data_points": [{
									"as_double": 1234.56
								}]
							}
						}]
					}]
				}]
			}`,
			includeScopeInfo: true,
			expect: `
				# TYPE otel_scope_info gauge
				otel_scope_info{name="a-name",version="a-version",something_extra="zzz-extra-value"} 1.0
				# TYPE test_metric_seconds gauge
				test_metric_seconds{otel_scope_name="a-name",otel_scope_version="a-version"} 1234.56
			`,
		},
		{
			name: "Labels from data point",
			input: `{
				"resource_metrics": [{
					"scope_metrics": [{
						"metrics": [{
							"name": "test_metric_seconds",
							"gauge": {
								"data_points": [{
									"attributes": [{
										"key": "foo",
										"value": { "stringValue": "bar" }
									}],
									"as_double": 1234.56
								}]
							}
						}]
					}]
				}]
			}`,
			expect: `
				# TYPE test_metric_seconds gauge
				test_metric_seconds{foo="bar"} 1234.56
			`,
		},
		{
			name: "Target info metric",
			input: `{
				"resource_metrics": [{
					"resource": {
						"attributes": [{
							"key": "service.name",
							"value": { "stringValue": "myservice" }
						}, {
							"key": "service.instance.id",
							"value": { "stringValue": "instance" }
						}, {
							"key": "custom_attr",
							"value": { "stringValue": "test" }
						}]
					},
					"scope_metrics": [{
						"metrics": [{
							"name": "test_metric_seconds",
							"gauge": {
								"data_points": [{
									"as_double": 1234.56
								}]
							}
						}]
					}]
				}]
			}`,
			includeTargetInfo: true,
			expect: `
				# HELP target_info Target metadata
				# TYPE target_info gauge
				target_info{instance="instance",job="myservice",custom_attr="test"} 1.0
				# TYPE test_metric_seconds gauge
				test_metric_seconds{instance="instance",job="myservice"} 1234.56
			`,
		},
	}

	decoder := &pmetric.JSONUnmarshaler{}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			payload, err := decoder.UnmarshalMetrics([]byte(tc.input))
			require.NoError(t, err)

			var app testappender.Appender
			app.HideTimestamps = !tc.showTimestamps

			l := util.TestFlowLogger(t)
			conv := convert.New(l, appenderAppendable{Inner: &app}, convert.Options{
				IncludeTargetInfo: tc.includeTargetInfo,
				IncludeScopeInfo:  tc.includeScopeInfo,
			})
			require.NoError(t, conv.ConsumeMetrics(context.Background(), payload))

			families, err := app.MetricFamilies()
			require.NoError(t, err)

			c := testappender.Comparer{OpenMetrics: true}
			require.NoError(t, c.Compare(families, tc.expect))
		})
	}
}

// appenderAppendable always returns the same Appender.
type appenderAppendable struct {
	Inner storage.Appender
}

var _ storage.Appendable = appenderAppendable{}

func (aa appenderAppendable) Appender(context.Context) storage.Appender {
	return aa.Inner
}
