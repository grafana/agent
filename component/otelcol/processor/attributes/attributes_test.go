package attributes_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component/otelcol"
	"github.com/grafana/agent/component/otelcol/internal/fakeconsumer"
	"github.com/grafana/agent/component/otelcol/processor/attributes"
	"github.com/grafana/agent/pkg/flow/componenttest"
	"github.com/grafana/agent/pkg/river"
	"github.com/grafana/agent/pkg/util"
	"github.com/grafana/dskit/backoff"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

// Test performs a basic integration test which runs the
// otelcol.processor.attributes component and ensures that it can accept, process, and forward data.
func Test(t *testing.T) {
	ctx := componenttest.TestContext(t)
	l := util.TestLogger(t)

	ctrl, err := componenttest.NewControllerFromID(l, "otelcol.processor.attributes")
	require.NoError(t, err)

	//TODO: Try testing with an invalid config, e.g. action = "dfdfgdsfg" ?
	//TODO: Run all these test configs automatically

	// cfg := `
	// processor_types {
	// 	traces = true
	// }
	// actions {
	// 	action {
	// 		key = "db.table"
	// 		action = "delete"
	// 	}
	// 	action {
	// 		key = "redacted_span"
	// 		value = "true"
	// 		action = "upsert"
	// 	}
	// 	action {
	// 		key = "copy_key"
	// 		from_attribute = "key_original"
	// 		action = "update"
	// 	}
	// 	action {
	// 		key = "account_id"
	// 		value = "2245"
	// 		action = "insert"
	// 	}
	// 	action {
	// 		key = "account_password"
	// 		action = "delete"
	// 	}
	// 	action {
	// 		key = "account_email"
	// 		action = "hash"
	// 	}
	// 	action {
	// 		key = "http.status_code"
	// 		action = "convert"
	// 		converted_type = "int"
	// 	}
	// }
	// match {
	// 	include {
	// 		match_type = "strict"
	// 		services = ["svcA", "svcB"]
	// 	}
	// 	exclude {
	// 		match_type = "strict"
	// 		attributes {
	// 			attribute {
	// 				key = "redact_trace"
	// 				value = false
	// 			}
	// 		}
	// 	}
	// }
	// output {
	// 	// no-op: will be overridden by test code.
	// }
	// `

	// cfg := `
	// processor_types {
	// 	traces = true
	// }
	// match {
	// 	include {
	// 		match_type = "strict"
	// 		services = ["svcA", "svcB"]
	// 	}
	// 	exclude {
	// 		match_type = "strict"
	// 		attributes {
	// 			attribute {
	// 				key = "redact_trace"
	// 				value = false
	// 			}
	// 		}
	// 	}
	// }
	// output {
	// 	// no-op: will be overridden by test code.
	// }
	// `

	// cfg := `
	// processor_types {
	// 	traces = true
	// }
	// actions {
	// 	action {
	// 		key = "credit_card"
	// 		action = "delete"
	// 	}
	// }
	// match {
	// 	exclude {
	// 		match_type = "strict"
	// 		libraries {
	// 			library{
	// 				name = "mongo-java-driver"
	// 				version = "3.8.0"
	// 			}
	// 		}
	// 	}
	// }
	// output {
	// 	// no-op: will be overridden by test code.
	// }
	// `

	cfg := `
	processor_types {
		traces = true
	}

	actions {
		action {
			key = "credit_card"
			action = "delete"
		}
	}

	match {
		exclude {
			match_type = "strict"
			resources {
				attribute {
					key = "host.type"
					value = "n1-standard-1"
				}
			} 
		}
	}

	output {
		// no-op: will be overridden by test code.
	}
	`

	var args attributes.Arguments
	require.NoError(t, river.Unmarshal([]byte(cfg), &args))

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
	case <-time.After(time.Second):
		require.FailNow(t, "failed waiting for traces")
	case tr := <-traceCh:
		require.Equal(t, 1, tr.SpanCount())
		require.Equal(t, 1, tr.ResourceSpans().Len())

		//TODO: Can we get the below to work?
		// attr := tr.ResourceSpans().At(0).Resource().Attributes()
		// require.Equal(t, 6, attr.Len())
		// {
		// 	val, err := attr.Get("account_id")
		// 	if !err {
		// 		require.Equal(t, 2245, val)
		// 	}
		// }
		// {
		// 	val, err := attr.Get("redacted_span")
		// 	if !err {
		// 		require.Equal(t, true, val)
		// 	}
		// }
		// {
		// 	val, err := attr.Get("copy_key")
		// 	if !err {
		// 		require.Equal(t, "original_val", val)
		// 	}
		// }
		// {
		// 	val, err := attr.Get("key_original")
		// 	if !err {
		// 		require.Equal(t, "original_val", val)
		// 	}
		// }
		// {
		// 	val, err := attr.Get("account_email")
		// 	if !err {
		// 		require.NotEqual(t, "val_to_be_hashed", val)
		// 	}
		// }
		// {
		// 	val, err := attr.Get("http.status_code")
		// 	if !err {
		// 		require.Equal(t, 500, val)
		// 	}
		// }
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
					"name": "TestSpan",
					"attributes": [{
						"db.table": "val_to_be_deleted",
						"redacted_span": false,
						"copy_key": "val_to_be_replaced",
						"key_original": "original_val",
						"account_password": "val_to_be_deleted",
						"account_email": "val_to_be_hashed",
						"http.status_code": "500"
					}]
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
