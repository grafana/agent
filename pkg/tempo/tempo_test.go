package tempo

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/grafana/agent/pkg/tempo/internal/tempoutils"
	"github.com/grafana/agent/pkg/util"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	"github.com/weaveworks/common/logging"
	"go.opentelemetry.io/collector/model/pdata"
	"gopkg.in/yaml.v2"
)

func TestTempo(t *testing.T) {
	tracesCh := make(chan pdata.Traces)
	tracesAddr := tempoutils.NewTestServer(t, func(t pdata.Traces) {
		tracesCh <- t
	})

	tempoCfgText := util.Untab(fmt.Sprintf(`
configs:
- name: default
  receivers:
    jaeger:
      protocols:
        thrift_compact:
  remote_write:
  	- endpoint: %s
      insecure: true
  batch:
    timeout: 100ms
    send_batch_size: 1
	`, tracesAddr))

	var cfg Config
	dec := yaml.NewDecoder(strings.NewReader(tempoCfgText))
	dec.SetStrict(true)
	err := dec.Decode(&cfg)
	require.NoError(t, err)

	var loggingLevel logging.Level
	require.NoError(t, loggingLevel.Set("debug"))

	tempo, err := New(nil, nil, prometheus.NewRegistry(), cfg, logrus.InfoLevel)
	require.NoError(t, err)
	t.Cleanup(tempo.Stop)

	tr := testJaegerTracer(t)
	span := tr.StartSpan("test-span")
	span.Finish()

	select {
	case <-time.After(30 * time.Second):
		require.Fail(t, "failed to receive a span after 30 seconds")
	case tr := <-tracesCh:
		require.Equal(t, 1, tr.SpanCount())
		// Nothing to do, send succeeded.
	}
}

func TestTempo_ApplyConfig(t *testing.T) {
	tracesCh := make(chan pdata.Traces)
	tracesAddr := tempoutils.NewTestServer(t, func(t pdata.Traces) {
		tracesCh <- t
	})

	tempoCfgText := util.Untab(`
configs:
- name: default
  receivers:
    jaeger:
      protocols:
        thrift_compact:
  remote_write:
  	- endpoint: 127.0.0.1:80 # deliberately the wrong endpoint
  	  insecure: true
  batch:
    timeout: 100ms
    send_batch_size: 1
  service_graphs:
    enabled: true
`)

	var cfg Config
	dec := yaml.NewDecoder(strings.NewReader(tempoCfgText))
	dec.SetStrict(true)
	err := dec.Decode(&cfg)
	require.NoError(t, err)

	tempo, err := New(nil, nil, prometheus.NewRegistry(), cfg, logrus.DebugLevel)
	require.NoError(t, err)
	t.Cleanup(tempo.Stop)

	// Fix the config and apply it before sending spans.
	tempoCfgText = util.Untab(fmt.Sprintf(`
configs:
- name: default
  receivers:
    jaeger:
      protocols:
        thrift_compact:
  remote_write:
  	- endpoint: %s
  	  insecure: true
  batch:
    timeout: 100ms
    send_batch_size: 1
	`, tracesAddr))

	var fixedConfig Config
	dec = yaml.NewDecoder(strings.NewReader(tempoCfgText))
	dec.SetStrict(true)
	err = dec.Decode(&fixedConfig)
	require.NoError(t, err)

	err = tempo.ApplyConfig(nil, nil, fixedConfig, logrus.DebugLevel)
	require.NoError(t, err)

	tr := testJaegerTracer(t)
	span := tr.StartSpan("test-span")
	span.Finish()

	select {
	case <-time.After(30 * time.Second):
		require.Fail(t, "failed to receive a span after 30 seconds")
	case tr := <-tracesCh:
		require.Equal(t, 1, tr.SpanCount())
		// Nothing to do, send succeeded.
	}
}

func testJaegerTracer(t *testing.T) opentracing.Tracer {
	t.Helper()

	jaegerConfig := jaegercfg.Configuration{
		ServiceName: "TestTempo",
		Sampler: &jaegercfg.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &jaegercfg.ReporterConfig{
			LocalAgentHostPort: "127.0.0.1:6831",
			LogSpans:           true,
		},
	}
	tr, closer, err := jaegerConfig.NewTracer()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, closer.Close())
	})

	return tr
}
