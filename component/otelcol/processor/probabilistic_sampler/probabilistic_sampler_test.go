package probabilistic_sampler_test

import (
	"context"
	"testing"

	probabilisticsampler "github.com/grafana/agent/component/otelcol/processor/probabilistic_sampler"
	"github.com/grafana/agent/component/otelcol/processor/processortest"
	"github.com/grafana/agent/pkg/flow/componenttest"
	"github.com/grafana/agent/pkg/util"
	"github.com/grafana/river"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/probabilisticsamplerprocessor"
	"github.com/stretchr/testify/require"
)

func TestArguments_UnmarshalRiver(t *testing.T) {
	tests := []struct {
		testName string
		cfg      string
		expected probabilisticsamplerprocessor.Config
		errorMsg string
	}{
		{
			testName: "Defaults",
			cfg: `
					output {}
				`,
			expected: probabilisticsamplerprocessor.Config{
				SamplingPercentage: 0,
				HashSeed:           0,
				AttributeSource:    "traceID",
				FromAttribute:      "",
				SamplingPriority:   "",
			},
		},
		{
			testName: "ExplicitValues",
			cfg: `
					sampling_percentage = 10
					hash_seed = 123
					attribute_source = "record"
					from_attribute = "logID"
					sampling_priority = "priority"
					output {}
				`,
			expected: probabilisticsamplerprocessor.Config{
				SamplingPercentage: 10,
				HashSeed:           123,
				AttributeSource:    "record",
				FromAttribute:      "logID",
				SamplingPriority:   "priority",
			},
		},
		{
			testName: "Negative SamplingPercentage",
			cfg: `
					sampling_percentage = -1
					output {}
				`,
			errorMsg: "negative sampling rate: -1.00",
		},
		{
			testName: "Invalid AttributeSource",
			cfg: `
					sampling_percentage = 0.1
					attribute_source = "example"
					output {}
				`,
			errorMsg: "invalid attribute source: example. Expected: traceID or record",
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			var args probabilisticsampler.Arguments
			err := river.Unmarshal([]byte(tc.cfg), &args)
			if tc.errorMsg != "" {
				require.EqualError(t, err, tc.errorMsg)
				return
			}
			require.NoError(t, err)

			actualPtr, err := args.Convert()
			require.NoError(t, err)

			actual := actualPtr.(*probabilisticsamplerprocessor.Config)
			require.Equal(t, tc.expected, *actual)
		})
	}
}

func testRunProcessor(t *testing.T, processorConfig string, testSignal processortest.Signal) {
	ctx := componenttest.TestContext(t)
	testRunProcessorWithContext(ctx, t, processorConfig, testSignal)
}

func testRunProcessorWithContext(ctx context.Context, t *testing.T, processorConfig string, testSignal processortest.Signal) {
	l := util.TestLogger(t)

	ctrl, err := componenttest.NewControllerFromID(l, "otelcol.processor.probabilistic_sampler")
	require.NoError(t, err)

	var args probabilisticsampler.Arguments
	require.NoError(t, river.Unmarshal([]byte(processorConfig), &args))

	// Override the arguments so signals get forwarded to the test channel.
	args.Output = testSignal.MakeOutput()

	prc := processortest.ProcessorRunConfig{
		Ctx:        ctx,
		T:          t,
		Args:       args,
		TestSignal: testSignal,
		Ctrl:       ctrl,
		L:          l,
	}
	processortest.TestRunProcessor(prc)
}

func TestLogProcessing(t *testing.T) {
	cfg := `
			sampling_percentage = 100
			hash_seed = 123
			output {
				// no-op: will be overridden by test code.
			}
		`
	var args probabilisticsampler.Arguments
	require.NoError(t, river.Unmarshal([]byte(cfg), &args))

	var inputLogs = `{
		"resourceLogs": [{
			"scopeLogs": [{
				"logRecords": [{
					"attributes": [{
						"key": "foo",
						"value": {
							"stringValue": "bar"
						}
					}]
				}]
			}]
		}]
	}`

	var expectedOutputLogs = `{
		"resourceLogs": [{
			"scopeLogs": [{
				"logRecords": [{
					"attributes": [{
						"key": "foo",
						"value": {
							"stringValue": "bar"
						}
					}]
				}]
			}]
		}]
	}`

	testRunProcessor(t, cfg, processortest.NewLogSignal(inputLogs, expectedOutputLogs))
}

func TestTraceProcessing(t *testing.T) {
	cfg := `
		sampling_percentage = 100
		hash_seed = 123
		output {
			// no-op: will be overridden by test code.
		}
	`

	var args probabilisticsampler.Arguments
	require.NoError(t, river.Unmarshal([]byte(cfg), &args))

	var inputTraces = `{
		"resourceSpans": [{
			"scopeSpans": [{
				"spans": [{
					"name": "TestSpan"
				}]
			}]
		}]
	}`

	expectedOutputTraces := `{
		"resourceSpans": [{
			"scopeSpans": [{
				"spans": [{
					"name": "TestSpan"
				}]
			}]
		}]
	}`

	testRunProcessor(t, cfg, processortest.NewTraceSignal(inputTraces, expectedOutputTraces))
}
