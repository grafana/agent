package servicegraphprocessor

import (
	"bytes"
	"context"
	"os"
	"testing"
	"time"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/grafana/agent/pkg/tempo/contextkeys"
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
			cfg:             &Config{},
			expectedMetrics: happyCaseExpectedMetrics,
		},
		{
			name:           "unpaired spans",
			sampleDataPath: unpairedTraceSamplePath,
			cfg: &Config{
				Wait: time.Millisecond,
			},
			expectedMetrics: `
				# HELP tempo_service_graph_unpaired_spans_total Total count of requests between two nodes
				# TYPE tempo_service_graph_unpaired_spans_total counter
				tempo_service_graph_unpaired_spans_total{client="",server="db"} 2
				tempo_service_graph_unpaired_spans_total{client="app",server=""} 3
				tempo_service_graph_unpaired_spans_total{client="lb",server=""} 3
        	    # HELP tempo_service_graph_untagged_spans_total Total count of spans processed that were not tagged with span.kind
        	    # TYPE tempo_service_graph_untagged_spans_total counter
        	    tempo_service_graph_untagged_spans_total{span_kind="SPAN_KIND_UNSPECIFIED"} 5
`,
		},
		{
			name:           "max items in store is reached",
			sampleDataPath: traceSamplePath,
			cfg: &Config{
				Wait:     time.Millisecond,
				MaxItems: 1, // Configure max number of items in store to 1. Only one edge will be processed.
			},
			expectedMetrics: `
				# HELP tempo_service_graph_unpaired_spans_total Total count of requests between two nodes
				# TYPE tempo_service_graph_unpaired_spans_total counter
				tempo_service_graph_unpaired_spans_total{client="lb",server=""} 1
				# HELP tempo_service_graph_untagged_spans_total Total count of spans processed that were not tagged with span.kind
				# TYPE tempo_service_graph_untagged_spans_total counter
				tempo_service_graph_untagged_spans_total{span_kind="SPAN_KIND_UNSPECIFIED"} 1
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
		# HELP tempo_service_graph_request_seconds Time for a request between two nodes
		# TYPE tempo_service_graph_request_seconds histogram
		tempo_service_graph_request_seconds_bucket{client="app",server="db",le="0.01"} 0
		tempo_service_graph_request_seconds_bucket{client="app",server="db",le="0.02"} 0
		tempo_service_graph_request_seconds_bucket{client="app",server="db",le="0.04"} 0
		tempo_service_graph_request_seconds_bucket{client="app",server="db",le="0.08"} 0
		tempo_service_graph_request_seconds_bucket{client="app",server="db",le="0.16"} 0
		tempo_service_graph_request_seconds_bucket{client="app",server="db",le="0.32"} 0
		tempo_service_graph_request_seconds_bucket{client="app",server="db",le="0.64"} 0
		tempo_service_graph_request_seconds_bucket{client="app",server="db",le="1.28"} 1
		tempo_service_graph_request_seconds_bucket{client="app",server="db",le="2.56"} 3
		tempo_service_graph_request_seconds_bucket{client="app",server="db",le="5.12"} 3
		tempo_service_graph_request_seconds_bucket{client="app",server="db",le="10.24"} 3
		tempo_service_graph_request_seconds_bucket{client="app",server="db",le="20.48"} 3
		tempo_service_graph_request_seconds_bucket{client="app",server="db",le="+Inf"} 3
		tempo_service_graph_request_seconds_sum{client="app",server="db"} 5
		tempo_service_graph_request_seconds_count{client="app",server="db"} 3
		tempo_service_graph_request_seconds_bucket{client="lb",server="app",le="0.01"} 0
		tempo_service_graph_request_seconds_bucket{client="lb",server="app",le="0.02"} 0
		tempo_service_graph_request_seconds_bucket{client="lb",server="app",le="0.04"} 0
		tempo_service_graph_request_seconds_bucket{client="lb",server="app",le="0.08"} 0
		tempo_service_graph_request_seconds_bucket{client="lb",server="app",le="0.16"} 0
		tempo_service_graph_request_seconds_bucket{client="lb",server="app",le="0.32"} 0
		tempo_service_graph_request_seconds_bucket{client="lb",server="app",le="0.64"} 0
		tempo_service_graph_request_seconds_bucket{client="lb",server="app",le="1.28"} 1
		tempo_service_graph_request_seconds_bucket{client="lb",server="app",le="2.56"} 2
		tempo_service_graph_request_seconds_bucket{client="lb",server="app",le="5.12"} 3
		tempo_service_graph_request_seconds_bucket{client="lb",server="app",le="10.24"} 3
		tempo_service_graph_request_seconds_bucket{client="lb",server="app",le="20.48"} 3
		tempo_service_graph_request_seconds_bucket{client="lb",server="app",le="+Inf"} 3
		tempo_service_graph_request_seconds_sum{client="lb",server="app"} 6.2
		tempo_service_graph_request_seconds_count{client="lb",server="app"} 3
		# HELP tempo_service_graph_request_total Total count of requests between two nodes
		# TYPE tempo_service_graph_request_total counter
		tempo_service_graph_request_total{client="app",server="db"} 3
		tempo_service_graph_request_total{client="lb",server="app"} 3
		# HELP tempo_service_graph_untagged_spans_total Total count of spans processed that were not tagged with span.kind
		# TYPE tempo_service_graph_untagged_spans_total counter
		tempo_service_graph_untagged_spans_total{span_kind="SPAN_KIND_UNSPECIFIED"} 5
`
)
