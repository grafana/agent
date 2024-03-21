package count

import (
	"context"
	"testing"
	"time"

	"github.com/grafana/agent/internal/component/otelcol"
	"github.com/grafana/agent/internal/component/otelcol/internal/fakeconsumer"
	"github.com/grafana/agent/internal/flow/componenttest"
	"github.com/grafana/agent/internal/util"
	"github.com/grafana/river"
	"github.com/open-telemetry/opentelemetry-collector-contrib/connector/countconnector"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"gopkg.in/yaml.v2"
)

func TestArguments_UnmarshalRiver(t *testing.T) {
	tests := []struct {
		name     string
		cfg      string
		expected *countconnector.Config
		errMsg   string
	}{
		{
			name: "Valid",
			cfg: `
      span {
        name = "spans.test"
        description = "foobar description"
        conditions = [
          "attributes[\"env\"] == \"test\"",
        ]
        attributes = [
          { "key" = "key", "default_value" = "val" },
        ]
      }
      spanevent {
        name = "spanevents.test"
        description = "foobar description"
        conditions = [
          "attributes[\"env\"] == \"test\"",
        ]
        attributes = [
          { "key" = "key", "default_value" = "val" },
        ]
      }
      metric {
        name = "metrics.test"
        description = "foobar description"
        conditions = [
          "resource.attributes[\"env\"] == \"test\"",
        ]
      }
      datapoint {
        name = "datapoints.test"
        description = "foobar description"
        conditions = [
          "attributes[\"env\"] == \"test\"",
        ]
        attributes = [
          { "key" = "key", "default_value" = "val" },
        ]
      }
      log {
        name = "logs.test"
        description = "foobar description"
        conditions = [
          "attributes[\"env\"] == \"test\"",
        ]
        attributes = [
          { "key" = "key", "default_value" = "val" },
        ]
      }
      output {}
      `,
			expected: &countconnector.Config{
				Spans:      testSpans(),
				SpanEvents: testSpanEvents(),
				Metrics:    testMetrics(),
				DataPoints: testDataPoints(),
				Logs:       testLogs(),
			},
			errMsg: "",
		},
		{
			name: "Defaults",
			cfg: `
      output {}
      `,
			expected: connectorDefaultConfig(t),
			errMsg:   "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var args Arguments
			err := river.Unmarshal([]byte(tt.cfg), &args)
			if tt.errMsg != "" {
				require.ErrorContains(t, err, tt.errMsg)
				return
			}

			require.NoError(t, err)

			for _, span := range args.Spans {
				mi, ok := tt.expected.Spans[span.Name]
				require.True(t, ok)
				require.Equal(t, mi.Description, span.Description)
				require.Equal(t, mi.Conditions, span.Conditions)
				require.Equal(t, len(mi.Attributes), len(span.Attributes))
			}
			for _, spanevent := range args.SpanEvents {
				mi, ok := tt.expected.SpanEvents[spanevent.Name]
				require.True(t, ok)
				require.Equal(t, mi.Description, spanevent.Description)
				require.Equal(t, mi.Conditions, spanevent.Conditions)
				require.Equal(t, len(mi.Attributes), len(spanevent.Attributes))
			}
			for _, metric := range args.Metrics {
				mi, ok := tt.expected.Metrics[metric.Name]
				require.True(t, ok)
				require.Equal(t, mi.Description, metric.Description)
				require.Equal(t, mi.Conditions, metric.Conditions)
				require.Equal(t, len(mi.Attributes), len(metric.Attributes))
			}
			for _, datapoint := range args.DataPoints {
				mi, ok := tt.expected.DataPoints[datapoint.Name]
				require.True(t, ok)
				require.Equal(t, mi.Description, datapoint.Description)
				require.Equal(t, mi.Conditions, datapoint.Conditions)
				require.Equal(t, len(mi.Attributes), len(datapoint.Attributes))
			}
			for _, log := range args.Logs {
				mi, ok := tt.expected.Logs[log.Name]
				require.True(t, ok)
				require.Equal(t, mi.Description, log.Description)
				require.Equal(t, mi.Conditions, log.Conditions)
				require.Equal(t, len(mi.Attributes), len(log.Attributes))
			}

		})
	}
}

func connectorDefaultConfig(t *testing.T) *countconnector.Config {
	var rawConf map[string]any
	cfg := &countconnector.Config{}
	require.NoError(t, yaml.Unmarshal([]byte(`count:`), &rawConf))
	require.NoError(t, cfg.Unmarshal(confmap.NewFromStringMap(rawConf)))
	return cfg
}

func TestArguments_Convert(t *testing.T) {
	tests := []struct {
		name     string
		args     Arguments
		expected *countconnector.Config
		errMsg   string
	}{
		{
			name: "convert success",
			args: Arguments{
				Spans: []MetricInfo{
					{
						Name:        "spans.test",
						Description: "foobar description",
						Conditions:  []string{`attributes["env"] == "test"`},
						Attributes: []AttributeConfig{
							{
								Key:          "key",
								DefaultValue: "val",
							},
						},
					},
				},
				SpanEvents: []MetricInfo{
					{
						Name:        "spanevents.test",
						Description: "foobar description",
						Conditions:  []string{`attributes["env"] == "test"`},
						Attributes: []AttributeConfig{
							{
								Key:          "key",
								DefaultValue: "val",
							},
						},
					},
				},
				Metrics: []MetricInfo{
					{
						Name:        "metrics.test",
						Description: "foobar description",
						Conditions:  []string{`resource.attributes["env"] == "test"`},
					},
				},
				DataPoints: []MetricInfo{
					{
						Name:        "datapoints.test",
						Description: "foobar description",
						Conditions:  []string{`attributes["env"] == "test"`},
						Attributes: []AttributeConfig{
							{
								Key:          "key",
								DefaultValue: "val",
							},
						},
					},
				},
				Logs: []MetricInfo{
					{
						Name:        "logs.test",
						Description: "foobar description",
						Conditions:  []string{`attributes["env"] == "test"`},
						Attributes: []AttributeConfig{
							{
								Key:          "key",
								DefaultValue: "val",
							},
						},
					},
				},
			},
			expected: &countconnector.Config{
				Spans:      testSpans(),
				SpanEvents: testSpanEvents(),
				Metrics:    testMetrics(),
				DataPoints: testDataPoints(),
				Logs:       testLogs(),
			},
			errMsg: "",
		},
		{
			name: "duplicate spans",
			args: Arguments{
				Spans: []MetricInfo{
					{
						Name: "foobar",
					},
					{
						Name: "foobar",
					},
				},
			},
			errMsg: "duplicate span name: foobar",
		},
		{
			name: "duplicate spanevents",
			args: Arguments{
				SpanEvents: []MetricInfo{
					{
						Name: "foobar",
					},
					{
						Name: "foobar",
					},
				},
			},
			errMsg: "duplicate spanevent name: foobar",
		},
		{
			name: "duplicate metrics",
			args: Arguments{
				Metrics: []MetricInfo{
					{
						Name: "foobar",
					},
					{
						Name: "foobar",
					},
				},
			},
			errMsg: "duplicate metric name: foobar",
		},
		{
			name: "duplicate datapoints",
			args: Arguments{
				DataPoints: []MetricInfo{
					{
						Name: "foobar",
					},
					{
						Name: "foobar",
					},
				},
			},
			errMsg: "duplicate datapoint name: foobar",
		},
		{
			name: "duplicate logs",
			args: Arguments{
				Logs: []MetricInfo{
					{
						Name: "foobar",
					},
					{
						Name: "foobar",
					},
				},
			},
			errMsg: "duplicate log name: foobar",
		},
		{
			name: "multiple duplicates",
			args: Arguments{
				Spans: []MetricInfo{
					{
						Name: "barbaz",
					},
					{
						Name: "barbaz",
					},
				},
				Logs: []MetricInfo{
					{
						Name: "foobar",
					},
					{
						Name: "foobar",
					},
				},
			},
			errMsg: "duplicate span name: barbaz; duplicate log name: foobar",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := tt.args.Convert()
			if tt.errMsg != "" {
				require.EqualError(t, err, tt.errMsg)
			}
			if cfg != nil {
				for _, span := range tt.args.Spans {
					mi, ok := tt.expected.Spans[span.Name]
					require.True(t, ok)
					require.Equal(t, mi.Description, span.Description)
					require.Equal(t, mi.Conditions, span.Conditions)
					require.Equal(t, len(mi.Attributes), len(span.Attributes))
				}
				for _, spanevent := range tt.args.SpanEvents {
					mi, ok := tt.expected.SpanEvents[spanevent.Name]
					require.True(t, ok)
					require.Equal(t, mi.Description, spanevent.Description)
					require.Equal(t, mi.Conditions, spanevent.Conditions)
					require.Equal(t, len(mi.Attributes), len(spanevent.Attributes))
				}
				for _, metric := range tt.args.Metrics {
					mi, ok := tt.expected.Metrics[metric.Name]
					require.True(t, ok)
					require.Equal(t, mi.Description, metric.Description)
					require.Equal(t, mi.Conditions, metric.Conditions)
					require.Equal(t, len(mi.Attributes), len(metric.Attributes))
				}
				for _, datapoint := range tt.args.DataPoints {
					mi, ok := tt.expected.DataPoints[datapoint.Name]
					require.True(t, ok)
					require.Equal(t, mi.Description, datapoint.Description)
					require.Equal(t, mi.Conditions, datapoint.Conditions)
					require.Equal(t, len(mi.Attributes), len(datapoint.Attributes))
				}
				for _, log := range tt.args.Logs {
					mi, ok := tt.expected.Logs[log.Name]
					require.True(t, ok)
					require.Equal(t, mi.Description, log.Description)
					require.Equal(t, mi.Conditions, log.Conditions)
					require.Equal(t, len(mi.Attributes), len(log.Attributes))
				}
			}
		})
	}
}

func testSpans() map[string]countconnector.MetricInfo {
	return testMetricInfoMap("spans.test")
}

func testSpanEvents() map[string]countconnector.MetricInfo {
	return testMetricInfoMap("spanevents.test")
}

func testMetrics() map[string]countconnector.MetricInfo {
	return map[string]countconnector.MetricInfo{
		"metrics.test": {
			Description: "foobar description",
			Conditions:  []string{`resource.attributes["env"] == "test"`},
		},
	}
}

func testDataPoints() map[string]countconnector.MetricInfo {
	return testMetricInfoMap("datapoints.test")
}

func testLogs() map[string]countconnector.MetricInfo {
	return testMetricInfoMap("logs.test")
}

func testMetricInfoMap(name string) map[string]countconnector.MetricInfo {
	return map[string]countconnector.MetricInfo{
		name: {
			Description: "foobar description",
			Conditions:  []string{`attributes["env"] == "test"`},
			Attributes: []countconnector.AttributeConfig{
				{
					Key:          "key",
					DefaultValue: "val",
				},
			},
		},
	}
}

func Test_ComponentIO(t *testing.T) {
	tests := []struct {
		name         string
		args         Arguments
		inputMetrics pmetric.Metrics
		inputLogs    plog.Logs
		inputTraces  ptrace.Traces
	}{
		{
			name:         "Defaults",
			args:         DefaultArguments,
			inputMetrics: newTestMetrics(),
			inputLogs:    newTestLogs(),
			inputTraces:  newTestTraces(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl, err := componenttest.NewControllerFromID(util.TestLogger(t), "otelcol.connector.count")
			require.NoError(t, err)

			metricsCh := make(chan pmetric.Metrics)
			done := make(chan struct{})

			myConsumer := fakeconsumer.Consumer{
				ConsumeMetricsFunc: func(_ context.Context, m pmetric.Metrics) error {
					metricsCh <- m
					return nil
				},
			}

			tt.args.Output = &otelcol.ConsumerArguments{
				Metrics: []otelcol.Consumer{&myConsumer},
			}

			go func() {
				err := ctrl.Run(context.Background(), tt.args)
				require.NoError(t, err)
			}()

			require.NoError(t, ctrl.WaitRunning(time.Second), "component never started")
			require.NoError(t, ctrl.WaitExports(time.Second), "component never exported anything")

			var outputMetrics []pmetric.Metrics

			go func() {
			loop:
				for {
					select {
					case <-done:
						break loop
					case metrics := <-metricsCh:
						if metrics.ResourceMetrics().Len() > 0 {
							outputMetrics = append(outputMetrics, metrics)
						}
					}
				}
			}()

			exports := ctrl.Exports().(otelcol.ConsumerExports)

			require.NoError(t, exports.Input.ConsumeTraces(context.Background(), tt.inputTraces))
			require.NoError(t, exports.Input.ConsumeLogs(context.Background(), tt.inputLogs))
			require.NoError(t, exports.Input.ConsumeMetrics(context.Background(), tt.inputMetrics))

			done <- struct{}{}

			require.GreaterOrEqual(t, len(outputMetrics), 3)

		})
	}
}

func newTestMetrics() pmetric.Metrics {
	metrics := pmetric.NewMetrics()
	ilm := metrics.ResourceMetrics().AppendEmpty().ScopeMetrics().AppendEmpty()
	ilm.Scope().SetName("unit.test")
	m := ilm.Metrics().AppendEmpty()
	m.SetName("metrics.test")
	m.SetEmptyGauge()
	dps := m.Gauge().DataPoints()
	dps.EnsureCapacity(5)
	timestamp := pcommon.NewTimestampFromTime(time.Now())
	for i := 0; i < 5; i++ {
		dpCalls := dps.AppendEmpty()
		dpCalls.SetStartTimestamp(timestamp)
		dpCalls.SetTimestamp(timestamp)
		dpCalls.SetIntValue(int64(1))
	}
	return metrics
}

func newTestLogs() plog.Logs {
	logs := plog.NewLogs()
	lr := logs.ResourceLogs().AppendEmpty().ScopeLogs().AppendEmpty().LogRecords().AppendEmpty()
	lr.Body().SetStr("test log record")
	return logs
}

func newTestTraces() ptrace.Traces {
	traces := ptrace.NewTraces()
	rs := traces.ResourceSpans().AppendEmpty()
	span := rs.ScopeSpans().AppendEmpty().Spans().AppendEmpty()
	span.SetName("test span")
	span.SetKind(ptrace.SpanKindInternal)
	return traces
}
