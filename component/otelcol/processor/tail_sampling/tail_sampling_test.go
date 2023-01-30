package tail_sampling

import (
	"context"
	"testing"
	"time"

	"github.com/grafana/agent/component/otelcol"
	"github.com/grafana/agent/component/otelcol/internal/fakeconsumer"
	"github.com/grafana/agent/pkg/flow/componenttest"
	"github.com/grafana/agent/pkg/river"
	"github.com/grafana/agent/pkg/util"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

const exampleConfig = `
  decision_wait               = "10s"
  num_traces                  = 100
  expected_new_traces_per_sec = 10
  policy {
    name = "test-policy-1"
    type = "always_sample"
  }
  policy {
    name = "test-policy-2"
  	type = "latency"
    latency {
      threshold_ms = 5000
    }
  }
  policy {
    name = "test-policy-3"
    type = "numeric_attribute"
    numeric_attribute {
      key = "key1"
      min_value = 50
      max_value = 100
    }
  }
  policy {
    name = "test-policy-4"
    type = "probabilistic"
    probabilistic {
      sampling_percentage = 10
    }
  }
  policy {
    name = "test-policy-5"
    type = "status_code"
    status_code {
      status_codes = ["ERROR", "UNSET"]
    }
  }
  policy {
    name = "test-policy-6"
    type = "string_attribute"
    string_attribute {
      key = "key2"
      values = ["value1", "value2"]
    }
  }
  policy {
    name = "test-policy-7"
    type = "string_attribute"
    string_attribute {
      key = "key2"
      values = ["value1", "val*"]
      enabled_regex_matching = true
      cache_max_size = 10
    }
  }
  policy {
    name = "test-policy-8"
    type = "rate_limiting"
    rate_limiting {
      spans_per_second = 35
    }
  }
  policy {
    name = "test-policy-9"
    type = "string_attribute"
    string_attribute {
      key = "http.url"
      values = ["/health", "/metrics"]
      enabled_regex_matching = true
      invert_match = true
    }
  }
  policy {
    name = "test-policy-10"
    type = "span_count"
    span_count {
      min_spans = 2
    }
  }
  policy {
    name = "test-policy-11"
    type = "trace_state"
    trace_state {
      key = "key3"
      values = ["value1", "value2"]
    }
  }
  policy{
    name = "and-policy-1"
    type = "and"
    and {
      and_sub_policy {
        name = "test-and-policy-1"
        type = "numeric_attribute"
        numeric_attribute {
          key = "key1"
          min_value = 50
          max_value = 100
        }
      }
      and_sub_policy {
        name = "test-and-policy-2"
        type = "string_attribute"
        string_attribute {
          key = "key1"
          values = ["value1", "value2"]
        }
      }
    }
  }
  policy{
    name = "composite-policy-1"
    type = "composite"
    composite {
      max_total_spans_per_second = 1000
      policy_order = ["test-composite-policy-1", "test-composite-policy-2", "test-composite-policy-3"]
      composite_sub_policy {
        name = "test-composite-policy-1"
        type = "numeric_attribute"
        numeric_attribute {
          key = "key1"
          min_value = 50
          max_value = 100
        }
      }
      composite_sub_policy {
        name = "test-composite-policy-2"
        type = "string_attribute"
        string_attribute {
          key = "key1"
          values = ["value1", "value2"]
        }
      }
      composite_sub_policy {
        name = "test-composite-policy-3"
        type = "always_sample"
      }
      rate_allocation {
        policy = "test-composite-policy-1"
        percent = 50
      }
      rate_allocation {
        policy = "test-composite-policy-2"
        percent = 50
      }
    }
  }

  output {
	// no-op: will be overridden by test code.
  }
`

func Test(t *testing.T) {
	ctx := componenttest.TestContext(t)
	l := util.TestLogger(t)

	ctrl, err := componenttest.NewControllerFromID(l, "otelcol.processor.tail_sampling")
	require.NoError(t, err)

	var args Arguments
	require.NoError(t, river.Unmarshal([]byte(exampleConfig), &args))

	// Override our arguments so traces get forwarded to traceCh.
	traceCh := make(chan ptrace.Traces)
	args.Output = makeTracesOutput(traceCh)

	go func() {
		err := ctrl.Run(ctx, args)
		require.NoError(t, err)
	}()

	require.NoError(t, ctrl.WaitRunning(time.Second), "component never started")
	require.NoError(t, ctrl.WaitExports(time.Second), "component never exported anything")
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
