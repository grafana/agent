package servicegraphprocessor

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/golang/protobuf/jsonpb"
	"github.com/grafana/tempo/pkg/tempopb"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/consumer/pdata"
)

const sampleDataRelPath = "testdata/trace-sample.json"

func TestConsumeMetrics(t *testing.T) {
	traces := traceSamples(t)

	p, err := newProcessor(&mockConsumer{}, &Config{})
	require.NoError(t, err)

	err = p.Start(context.Background(), nil)
	require.NoError(t, err)

	err = p.ConsumeTraces(context.Background(), traces)
	require.NoError(t, err)

	err = testutil.GatherAndCompare(p.reg.(prometheus.Gatherer), bytes.NewBufferString(`
		# HELP tempo_service_graph_request_seconds Time for a request between two nodes
		# TYPE tempo_service_graph_request_seconds histogram
		tempo_service_graph_request_seconds_bucket{client="app",server="db",le="0.01"} 3
		tempo_service_graph_request_seconds_bucket{client="app",server="db",le="0.02"} 3
		tempo_service_graph_request_seconds_bucket{client="app",server="db",le="0.04"} 3
		tempo_service_graph_request_seconds_bucket{client="app",server="db",le="0.08"} 3
		tempo_service_graph_request_seconds_bucket{client="app",server="db",le="0.16"} 3
		tempo_service_graph_request_seconds_bucket{client="app",server="db",le="0.32"} 3
		tempo_service_graph_request_seconds_bucket{client="app",server="db",le="0.64"} 3
		tempo_service_graph_request_seconds_bucket{client="app",server="db",le="1.28"} 3
		tempo_service_graph_request_seconds_bucket{client="app",server="db",le="2.56"} 3
		tempo_service_graph_request_seconds_bucket{client="app",server="db",le="5.12"} 3
		tempo_service_graph_request_seconds_bucket{client="app",server="db",le="10.24"} 3
		tempo_service_graph_request_seconds_bucket{client="app",server="db",le="20.48"} 3
		tempo_service_graph_request_seconds_bucket{client="app",server="db",le="+Inf"} 3
		tempo_service_graph_request_seconds_sum{client="app",server="db"} 0.00033999999999999997
		tempo_service_graph_request_seconds_count{client="app",server="db"} 3
		tempo_service_graph_request_seconds_bucket{client="lb",server="app",le="0.01"} 3
		tempo_service_graph_request_seconds_bucket{client="lb",server="app",le="0.02"} 3
		tempo_service_graph_request_seconds_bucket{client="lb",server="app",le="0.04"} 3
		tempo_service_graph_request_seconds_bucket{client="lb",server="app",le="0.08"} 3
		tempo_service_graph_request_seconds_bucket{client="lb",server="app",le="0.16"} 3
		tempo_service_graph_request_seconds_bucket{client="lb",server="app",le="0.32"} 3
		tempo_service_graph_request_seconds_bucket{client="lb",server="app",le="0.64"} 3
		tempo_service_graph_request_seconds_bucket{client="lb",server="app",le="1.28"} 3
		tempo_service_graph_request_seconds_bucket{client="lb",server="app",le="2.56"} 3
		tempo_service_graph_request_seconds_bucket{client="lb",server="app",le="5.12"} 3
		tempo_service_graph_request_seconds_bucket{client="lb",server="app",le="10.24"} 3
		tempo_service_graph_request_seconds_bucket{client="lb",server="app",le="20.48"} 3
		tempo_service_graph_request_seconds_bucket{client="lb",server="app",le="+Inf"} 3
		tempo_service_graph_request_seconds_sum{client="lb",server="app"} 0.013146
		tempo_service_graph_request_seconds_count{client="lb",server="app"} 3
		# HELP tempo_service_graph_request_total Total count of requests between two nodes
		# TYPE tempo_service_graph_request_total counter
		tempo_service_graph_request_total{client="app",server="db"} 3
		tempo_service_graph_request_total{client="lb",server="app"} 3
	`))
	require.NoError(t, err)
}

func traceSamples(t *testing.T) pdata.Traces {
	f, err := os.Open(sampleDataRelPath)
	require.NoError(t, err)

	r := &tempopb.Trace{}
	err = jsonpb.Unmarshal(f, r)
	require.NoError(t, err)

	b, err := r.Marshal()
	require.NoError(t, err)

	traces, err := pdata.TracesFromOtlpProtoBytes(b)
	require.NoError(t, err)

	return traces
}

type mockConsumer struct{}

func (m *mockConsumer) Capabilities() consumer.Capabilities { return consumer.Capabilities{} }

func (m *mockConsumer) ConsumeTraces(context.Context, pdata.Traces) error { return nil }
