package servicegraphprocessor

import (
	"bytes"
	"context"
	"os"
	"testing"
	"time"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/grafana/agent/pkg/traces/contextkeys"
	"github.com/grafana/tempo/pkg/tempopb"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/model/otlp"
	"go.opentelemetry.io/collector/model/pdata"
)

const (
	traceSamplePath         = "testdata/trace-sample.json"
	unpairedTraceSamplePath = "testdata/unpaired-trace-sample.json"
)

func TestConsumeMetrics(t *testing.T) {
	for _, tc := range []struct {
		name            string
		sampleDataPath  string
		cfg             *Config
		expectedMetrics string
	}{
		{
			name:            "happy case",
			sampleDataPath:  traceSamplePath,
			cfg:             &Config{
				Wait: time.Hour,
			},
			expectedMetrics: happyCaseExpectedMetrics,
		},
		{
			name:           "unpaired spans",
			sampleDataPath: unpairedTraceSamplePath,
			cfg: &Config{
				Wait: time.Millisecond,
			},
			expectedMetrics: `
				# HELP traces_service_graph_unpaired_spans_total Total count of unpaired spans
				# TYPE traces_service_graph_unpaired_spans_total counter
				traces_service_graph_unpaired_spans_total{client="",server="db"} 2
				traces_service_graph_unpaired_spans_total{client="app",server=""} 3
				traces_service_graph_unpaired_spans_total{client="lb",server=""} 3
`,
		},
		{
			name:           "max items in storeMap is reached",
			sampleDataPath: traceSamplePath,
			cfg: &Config{
				Wait:     time.Millisecond,
				MaxItems: 1, // Configure max number of items in storeMap to 1. Only one edge will be processed.
			},
			expectedMetrics: `
        	    # HELP traces_service_graph_dropped_spans_total Total count of dropped spans
        	    # TYPE traces_service_graph_dropped_spans_total counter
        	    traces_service_graph_dropped_spans_total{service="lb"} 1
				# HELP traces_service_graph_unpaired_spans_total Total count of unpaired spans
				# TYPE traces_service_graph_unpaired_spans_total counter
				traces_service_graph_unpaired_spans_total{client="lb",server=""} 1
`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := newProcessor(&mockConsumer{}, tc.cfg)

			reg := prometheus.NewRegistry()
			ctx := context.WithValue(context.Background(), contextkeys.PrometheusRegisterer, reg)

			err := p.Start(ctx, nil)
			require.NoError(t, err)

			traces := traceSamples(t, tc.sampleDataPath)
			err = p.ConsumeTraces(context.Background(), traces)
			require.NoError(t, err)

			p.collectMetrics()

			assert.Eventually(t, func() bool {
				return testutil.GatherAndCompare(reg, bytes.NewBufferString(tc.expectedMetrics)) == nil
			}, time.Second, time.Millisecond*100)
			err = testutil.GatherAndCompare(reg, bytes.NewBufferString(tc.expectedMetrics))
			require.NoError(t, err)
		})
	}

}

func traceSamples(t *testing.T, path string) pdata.Traces {
	f, err := os.Open(path)
	require.NoError(t, err)

	r := &tempopb.Trace{}
	err = jsonpb.Unmarshal(f, r)
	require.NoError(t, err)

	b, err := r.Marshal()
	require.NoError(t, err)

	traces, err := otlp.NewProtobufTracesUnmarshaler().UnmarshalTraces(b)
	require.NoError(t, err)

	return traces
}

type mockConsumer struct{}

func (m *mockConsumer) Capabilities() consumer.Capabilities { return consumer.Capabilities{} }

func (m *mockConsumer) ConsumeTraces(context.Context, pdata.Traces) error { return nil }

const (
	happyCaseExpectedMetrics = `
		# HELP traces_service_graph_request_client_seconds Time for a request between two nodes as seen from the client
		# TYPE traces_service_graph_request_client_seconds histogram
		traces_service_graph_request_client_seconds_bucket{client="app",server="db",le="0.01"} 0
		traces_service_graph_request_client_seconds_bucket{client="app",server="db",le="0.02"} 0
		traces_service_graph_request_client_seconds_bucket{client="app",server="db",le="0.04"} 0
		traces_service_graph_request_client_seconds_bucket{client="app",server="db",le="0.08"} 0
		traces_service_graph_request_client_seconds_bucket{client="app",server="db",le="0.16"} 0
		traces_service_graph_request_client_seconds_bucket{client="app",server="db",le="0.32"} 0
		traces_service_graph_request_client_seconds_bucket{client="app",server="db",le="0.64"} 0
		traces_service_graph_request_client_seconds_bucket{client="app",server="db",le="1.28"} 2
		traces_service_graph_request_client_seconds_bucket{client="app",server="db",le="2.56"} 3
		traces_service_graph_request_client_seconds_bucket{client="app",server="db",le="5.12"} 3
		traces_service_graph_request_client_seconds_bucket{client="app",server="db",le="10.24"} 3
		traces_service_graph_request_client_seconds_bucket{client="app",server="db",le="20.48"} 3
		traces_service_graph_request_client_seconds_bucket{client="app",server="db",le="+Inf"} 3
		traces_service_graph_request_client_seconds_sum{client="app",server="db"} 4.4
		traces_service_graph_request_client_seconds_count{client="app",server="db"} 3
		traces_service_graph_request_client_seconds_bucket{client="lb",server="app",le="0.01"} 0
		traces_service_graph_request_client_seconds_bucket{client="lb",server="app",le="0.02"} 0
		traces_service_graph_request_client_seconds_bucket{client="lb",server="app",le="0.04"} 0
		traces_service_graph_request_client_seconds_bucket{client="lb",server="app",le="0.08"} 0
		traces_service_graph_request_client_seconds_bucket{client="lb",server="app",le="0.16"} 0
		traces_service_graph_request_client_seconds_bucket{client="lb",server="app",le="0.32"} 0
		traces_service_graph_request_client_seconds_bucket{client="lb",server="app",le="0.64"} 0
		traces_service_graph_request_client_seconds_bucket{client="lb",server="app",le="1.28"} 0
		traces_service_graph_request_client_seconds_bucket{client="lb",server="app",le="2.56"} 2
		traces_service_graph_request_client_seconds_bucket{client="lb",server="app",le="5.12"} 3
		traces_service_graph_request_client_seconds_bucket{client="lb",server="app",le="10.24"} 3
		traces_service_graph_request_client_seconds_bucket{client="lb",server="app",le="20.48"} 3
		traces_service_graph_request_client_seconds_bucket{client="lb",server="app",le="+Inf"} 3
		traces_service_graph_request_client_seconds_sum{client="lb",server="app"} 7.8
		traces_service_graph_request_client_seconds_count{client="lb",server="app"} 3
		# HELP traces_service_graph_request_failed_total Total count of failed requests between two nodes
		# TYPE traces_service_graph_request_failed_total counter
		traces_service_graph_request_failed_total{client="app",server="db"} 3
		traces_service_graph_request_failed_total{client="lb",server="app"} 3
		# HELP traces_service_graph_request_server_seconds Time for a request between two nodes as seen from the server
		# TYPE traces_service_graph_request_server_seconds histogram
		traces_service_graph_request_server_seconds_bucket{client="app",server="db",le="0.01"} 0
		traces_service_graph_request_server_seconds_bucket{client="app",server="db",le="0.02"} 0
		traces_service_graph_request_server_seconds_bucket{client="app",server="db",le="0.04"} 0
		traces_service_graph_request_server_seconds_bucket{client="app",server="db",le="0.08"} 0
		traces_service_graph_request_server_seconds_bucket{client="app",server="db",le="0.16"} 0
		traces_service_graph_request_server_seconds_bucket{client="app",server="db",le="0.32"} 0
		traces_service_graph_request_server_seconds_bucket{client="app",server="db",le="0.64"} 0
		traces_service_graph_request_server_seconds_bucket{client="app",server="db",le="1.28"} 1
		traces_service_graph_request_server_seconds_bucket{client="app",server="db",le="2.56"} 3
		traces_service_graph_request_server_seconds_bucket{client="app",server="db",le="5.12"} 3
		traces_service_graph_request_server_seconds_bucket{client="app",server="db",le="10.24"} 3
		traces_service_graph_request_server_seconds_bucket{client="app",server="db",le="20.48"} 3
		traces_service_graph_request_server_seconds_bucket{client="app",server="db",le="+Inf"} 3
		traces_service_graph_request_server_seconds_sum{client="app",server="db"} 5
		traces_service_graph_request_server_seconds_count{client="app",server="db"} 3
		traces_service_graph_request_server_seconds_bucket{client="lb",server="app",le="0.01"} 0
		traces_service_graph_request_server_seconds_bucket{client="lb",server="app",le="0.02"} 0
		traces_service_graph_request_server_seconds_bucket{client="lb",server="app",le="0.04"} 0
		traces_service_graph_request_server_seconds_bucket{client="lb",server="app",le="0.08"} 0
		traces_service_graph_request_server_seconds_bucket{client="lb",server="app",le="0.16"} 0
		traces_service_graph_request_server_seconds_bucket{client="lb",server="app",le="0.32"} 0
		traces_service_graph_request_server_seconds_bucket{client="lb",server="app",le="0.64"} 0
		traces_service_graph_request_server_seconds_bucket{client="lb",server="app",le="1.28"} 1
		traces_service_graph_request_server_seconds_bucket{client="lb",server="app",le="2.56"} 2
		traces_service_graph_request_server_seconds_bucket{client="lb",server="app",le="5.12"} 3
		traces_service_graph_request_server_seconds_bucket{client="lb",server="app",le="10.24"} 3
		traces_service_graph_request_server_seconds_bucket{client="lb",server="app",le="20.48"} 3
		traces_service_graph_request_server_seconds_bucket{client="lb",server="app",le="+Inf"} 3
		traces_service_graph_request_server_seconds_sum{client="lb",server="app"} 6.2
		traces_service_graph_request_server_seconds_count{client="lb",server="app"} 3
		# HELP traces_service_graph_request_total Total count of requests between two nodes
		# TYPE traces_service_graph_request_total counter
		traces_service_graph_request_total{client="app",server="db"} 3
		traces_service_graph_request_total{client="lb",server="app"} 3
`
)
