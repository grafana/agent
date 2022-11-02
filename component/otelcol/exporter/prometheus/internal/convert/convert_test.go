package convert_test

import (
	"context"
	"testing"

	"github.com/grafana/agent/component/otelcol/exporter/prometheus/internal/convert"
	"github.com/grafana/agent/pkg/util"
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
			name: "Gauge",
			input: `{
				"resource_metrics": [{
					"resource": {
						"attributes": [{
							"key": "service.name",
							"value": { "stringValue": "myservice" }
						}, {
							"key": "service.instance.id",
							"value": { "stringValue": "instance" }
						}]
					},
					"scope_metrics": [{
						"metrics": [{
							"name": "test_metric_seconds",
							"description": "Metric for testing",
							"unit": "seconds",
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
				# HELP test_metric_seconds Metric for testing
				# TYPE test_metric_seconds gauge
				test_metric_seconds{instance="instance",job="myservice"} 1234.56
			`,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			payload, err := pmetric.NewJSONUnmarshaler().UnmarshalMetrics([]byte(tc.input))
			require.NoError(t, err)

			var app testappender.Appender
			app.HideTimestamps = true

			l := util.TestLogger(t)
			conv := convert.New(l, appenderAppendable{Inner: &app}, convert.Options{})
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
