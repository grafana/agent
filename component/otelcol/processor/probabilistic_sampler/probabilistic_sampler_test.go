//go:build !race

package probabilistic_sampler

import (
	"context"
	"testing"
	"time"

	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component/otelcol/internal/fakeconsumer"
	"github.com/grafana/agent/pkg/util"

	"github.com/grafana/agent/component/otelcol"
	"github.com/grafana/agent/pkg/flow/componenttest"
	"github.com/grafana/dskit/backoff"
	"github.com/grafana/river"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

func TestBadRiverConfigNegativeSamplingRate(t *testing.T) {
	exampleBadRiverConfig := `
    sampling_percentage = -1
    output { 
	    // no-op: will be overridden by test code.
    }
`
	var args Arguments
	require.EqualError(t, river.Unmarshal([]byte(exampleBadRiverConfig), &args), "negative sampling rate: -1.00")
}

func TestBadRiverConfigInvalidAttributeSource(t *testing.T) {
	exampleBadRiverConfig := `
    sampling_percentage = 0.1
    attribute_source = "example"
    output { 
	    // no-op: will be overridden by test code.
    }
`
	var args Arguments
	require.EqualError(t, river.Unmarshal([]byte(exampleBadRiverConfig), &args), "invalid attribute source: example. Expected: traceID or record")
}

func TestLogProcessing(t *testing.T) {
	exampleSmallConfig := `
	sampling_percentage        = 100
    hash_seed                  = 123
    
    output { 
	    // no-op: will be overridden by test code.
    }
  `
	ctx := componenttest.TestContext(t)
	l := util.TestLogger(t)

	ctrl, err := componenttest.NewControllerFromID(l, "otelcol.processor.probabilistic_sampler")
	require.NoError(t, err)

	var args Arguments
	require.NoError(t, river.Unmarshal([]byte(exampleSmallConfig), &args))

	// Override our arguments so logs get forwarded to logsCh.
	logsCh := make(chan plog.Logs)
	args.Output = makeLogsOutput(logsCh)

	go func() {
		err := ctrl.Run(ctx, args)
		require.NoError(t, err)
	}()

	require.NoError(t, ctrl.WaitRunning(time.Second), "component never started")
	require.NoError(t, ctrl.WaitExports(time.Second), "component never exported anything")

	// Send traces in the background to our processor.
	go func() {
		exports := ctrl.Exports().(otelcol.ConsumerExports)

		exports.Input.Capabilities()

		bo := backoff.New(ctx, backoff.Config{
			MinBackoff: 10 * time.Millisecond,
			MaxBackoff: 100 * time.Millisecond,
		})
		for bo.Ongoing() {
			err := exports.Input.ConsumeLogs(ctx, createTestLogs())
			if err != nil {
				level.Error(l).Log("msg", "failed to send logs", "err", err)
				bo.Wait()
				continue
			}

			return
		}
	}()

	// Wait for our processor to finish and forward data to logCh.
	select {
	case <-time.After(time.Second * 10):
		require.FailNow(t, "failed waiting for logs")
	case tr := <-logsCh:
		require.Equal(t, 1, tr.LogRecordCount())
	}
}

func TestTraceProcessing(t *testing.T) {
	exampleSmallConfig := `
    sampling_percentage        = 100
    hash_seed                  = 123
    
    output { 
	    // no-op: will be overridden by test code.
    }
  `
	ctx := componenttest.TestContext(t)
	l := util.TestLogger(t)

	ctrl, err := componenttest.NewControllerFromID(l, "otelcol.processor.probabilistic_sampler")
	require.NoError(t, err)

	var args Arguments
	require.NoError(t, river.Unmarshal([]byte(exampleSmallConfig), &args))

	// Override our arguments so traces get forwarded to traceCh.
	traceCh := make(chan ptrace.Traces)
	args.Output = makeTracesOutput(traceCh)

	go func() {
		err := ctrl.Run(ctx, args)
		require.NoError(t, err)
	}()

	require.NoError(t, ctrl.WaitRunning(time.Second), "component never started")
	require.NoError(t, ctrl.WaitExports(time.Second), "component never exported anything")

	// Send traces in the background to our processor.
	go func() {
		exports := ctrl.Exports().(otelcol.ConsumerExports)

		exports.Input.Capabilities()

		bo := backoff.New(ctx, backoff.Config{
			MinBackoff: 10 * time.Millisecond,
			MaxBackoff: 100 * time.Millisecond,
		})
		for bo.Ongoing() {
			err := exports.Input.ConsumeTraces(ctx, createTestTraces())
			if err != nil {
				level.Error(l).Log("msg", "failed to send traces", "err", err)
				bo.Wait()
				continue
			}

			return
		}
	}()

	// Wait for our processor to finish and forward data to traceCh.
	select {
	case <-time.After(time.Second * 10):
		require.FailNow(t, "failed waiting for traces")
	case tr := <-traceCh:
		require.Equal(t, 1, tr.SpanCount())
	}
}

// makeTracesOutput returns ConsumerArguments which will forward traces to the
// provided channel.
func makeTracesOutput(ch chan ptrace.Traces) *otelcol.ConsumerArguments {
	traceConsumer := fakeconsumer.Consumer{
		ConsumeTracesFunc: func(ctx context.Context, t ptrace.Traces) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case ch <- t:
				return nil
			}
		},
	}

	return &otelcol.ConsumerArguments{
		Traces: []otelcol.Consumer{&traceConsumer},
	}
}

func createTestTraces() ptrace.Traces {
	// Matches format from the protobuf definition:
	// https://github.com/open-telemetry/opentelemetry-proto/blob/main/opentelemetry/proto/trace/v1/trace.proto
	var bb = `{
		"resource_spans": [{
			"scope_spans": [{
				"spans": [{
					"name": "TestSpan"
				}]
			}]
		}]
	}`

	decoder := &ptrace.JSONUnmarshaler{}
	data, err := decoder.UnmarshalTraces([]byte(bb))
	if err != nil {
		panic(err)
	}
	return data
}

// makeLogsOutput returns ConsumerArguments which will forward logs to the
// provided channel.
func makeLogsOutput(ch chan plog.Logs) *otelcol.ConsumerArguments {
	logConsumer := fakeconsumer.Consumer{
		ConsumeLogsFunc: func(ctx context.Context, t plog.Logs) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case ch <- t:
				return nil
			}
		},
	}

	return &otelcol.ConsumerArguments{
		Logs: []otelcol.Consumer{&logConsumer},
	}
}

func createTestLogs() plog.Logs {
	// Matches format from the protobuf definition:
	// https://github.com/open-telemetry/opentelemetry-proto/blob/main/opentelemetry/proto/logs/v1/logs.proto
	var bb = `{
		"resource_logs": [{
			"scope_logs": [{
				"log_records": [{
                    "attributes": [{
                    	"key": "foo",
						"value": {
							"string_value": "bar"
                        }
                    }]
				}]
			}]
		}]
	}`

	decoder := &plog.JSONUnmarshaler{}
	data, err := decoder.UnmarshalLogs([]byte(bb))
	if err != nil {
		panic(err)
	}
	return data
}
