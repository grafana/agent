package prometheus_test

import (
	"context"
	"testing"
	"time"

	"github.com/grafana/agent/component/otelcol"
	"github.com/grafana/agent/component/otelcol/internal/fakeconsumer"
	"github.com/grafana/agent/component/otelcol/receiver/prometheus"
	flowprometheus "github.com/grafana/agent/component/prometheus"
	"github.com/grafana/agent/pkg/flow/componenttest"
	"github.com/grafana/agent/pkg/river"
	"github.com/grafana/agent/pkg/util"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/scrape"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

// Test performs a basic integration test which runs the
// otelcol.receiver.prometheus component and ensures that it can receive and
// forward metric data.
func Test(t *testing.T) {
	ctx := componenttest.TestContext(t)
	l := util.TestLogger(t)

	ctrl, err := componenttest.NewControllerFromID(l, "otelcol.receiver.prometheus")
	require.NoError(t, err)

	cfg := `
		output {
			// no-op: will be overridden by test code.
		}
	`
	var args prometheus.Arguments
	require.NoError(t, river.Unmarshal([]byte(cfg), &args))

	// Override our settings so metrics get forwarded to metricCh.
	metricCh := make(chan pmetric.Metrics)
	args.Output = makeMetricsOutput(metricCh)

	go func() {
		err := ctrl.Run(ctx, args)
		require.NoError(t, err)
	}()

	require.NoError(t, ctrl.WaitRunning(time.Second))
	require.NoError(t, ctrl.WaitExports(time.Second))

	exports := ctrl.Exports().(prometheus.Exports)

	// Use the exported Appendable to send metrics to the receiver in the
	// background.
	go func() {
		l := labels.Labels{
			{Name: model.MetricNameLabel, Value: "testMetric"},
			{Name: model.JobLabel, Value: "testJob"},
			{Name: model.InstanceLabel, Value: "otelcol.receiver.prometheus"},
			{Name: "foo", Value: "bar"},
		}
		ts := time.Now().Unix()
		v := 100.

		ctx := context.Background()
		ctx = scrape.ContextWithMetricMetadataStore(ctx, flowprometheus.SimpleMetadataStore{})
		ctx = scrape.ContextWithTarget(ctx, &scrape.Target{})
		app := exports.Receiver.Appender(ctx)
		_, err := app.Append(0, l, ts, v)
		require.NoError(t, err)
		require.NoError(t, app.Commit())
	}()

	// Wait for our client to get the metric.
	select {
	case <-time.After(time.Second):
		require.FailNow(t, "failed waiting for metrics")
	case m := <-metricCh:
		require.Equal(t, 1, m.MetricCount())
		require.Equal(t, "testMetric", m.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0).Name())
	}
}

// makeMetricsOutput returns a ConsumerArguments which will forward metrics to
// the provided channel.
func makeMetricsOutput(ch chan pmetric.Metrics) *otelcol.ConsumerArguments {
	metricsConsumer := fakeconsumer.Consumer{
		ConsumeMetricsFunc: func(ctx context.Context, m pmetric.Metrics) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case ch <- m:
				return nil
			}
		},
	}

	return &otelcol.ConsumerArguments{
		Metrics: []otelcol.Consumer{&metricsConsumer},
	}
}
