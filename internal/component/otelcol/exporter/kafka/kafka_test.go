package kafka_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/grafana/agent/internal/component/otelcol"
	"github.com/grafana/agent/internal/component/otelcol/exporter/kafka"
	"github.com/grafana/agent/internal/flow/componenttest"
	"github.com/grafana/agent/internal/flow/logging/level"
	"github.com/grafana/agent/internal/util"
	"github.com/grafana/dskit/backoff"
	"github.com/grafana/river"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/pdata/ptrace/ptraceotlp"
)

// Test performs a basic integration test which runs the otelcol.exporter.kafka
// component and ensures that it can pass data to a mock Kafka broker.
func Test(t *testing.T) {
	ch := make(chan string)
	kafkaBroker := makeKafkaBroker(t, ch)

	ctx := componenttest.TestContext(t)
	l := util.TestLogger(t)

	ctrl, err := componenttest.NewControllerFromID(l, "otelcol.exporter.kafka")
	require.NoError(t, err)

	cfg := fmt.Sprintf(`
		protocol_version = "2.0.0"
		brokers          = ["%s"]
		// brokers          = ["localhost:9002"]
		timeout          = "250ms"
		metadata {
			include_all_topics = false
		}

		debug_metrics {
			disable_high_cardinality_metrics = true
		}
	`, kafkaBroker.Addr())
	var args kafka.Arguments
	require.NoError(t, river.Unmarshal([]byte(cfg), &args))
	require.Equal(t, args.DebugMetricsConfig().DisableHighCardinalityMetrics, true)

	go func() {
		err := ctrl.Run(ctx, args)
		require.NoError(t, err)
	}()

	require.NoError(t, ctrl.WaitRunning(time.Second), "component never started")
	require.NoError(t, ctrl.WaitExports(time.Second), "component never exported anything")

	// Send traces in the background to our exporter.
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

	// Wait for our exporter to finish and pass data to our HTTP server.
	select {
	case <-time.After(time.Second):
		require.FailNow(t, "failed waiting for traces")
	case tr := <-ch:
		require.Equal(t, "&{6 [otlp_spans] true false false}", tr)
	}
}

// makeTracesServer returns a host:port which will accept Kafka messages.
func makeKafkaBroker(t *testing.T, ch chan string) *sarama.MockBroker {
	t.Helper()

	broker := sarama.NewMockBroker(t, 0)
	t.Cleanup(broker.Close)

	go func() {
		for {
			if len(broker.History()) > 0 {
				rr := broker.History()[0]
				ch <- fmt.Sprint(rr.Request)
				return
			}
		}
	}()

	return broker
}

type mockTracesReceiver struct {
	ptraceotlp.UnimplementedGRPCServer
	ch chan ptrace.Traces
}

var _ ptraceotlp.GRPCServer = (*mockTracesReceiver)(nil)

func (ms *mockTracesReceiver) Export(_ context.Context, req ptraceotlp.ExportRequest) (ptraceotlp.ExportResponse, error) {
	ms.ch <- req.Traces()
	return ptraceotlp.NewExportResponse(), nil
}

func createTestTraces() ptrace.Traces {
	// Matches format from the protobuf definition:
	// https://github.com/open-telemetry/opentelemetry-proto/blob/main/opentelemetry/proto/trace/v1/trace.proto
	bb := `{
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
