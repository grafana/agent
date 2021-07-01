package remotewriteexporter

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/grafana/agent/pkg/prom/instance"
	"github.com/prometheus/prometheus/pkg/exemplar"
	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/scrape"
	"github.com/prometheus/prometheus/storage"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/consumer/pdata"
)

const (
	sumMetric    = "tempo_spanmetrics_latency_sum"
	countMetric  = "tempo_spanmetrics_latency_count"
	bucketMetric = "tempo_spanmetrics_latency_bucket"
)

func TestRemoteWriteExporter_handleHistogramIntDataPoints(t *testing.T) {
	var (
		countValue     uint64 = 20
		sumValue       int64  = 100
		bucketCounts          = []uint64{1, 2, 3, 4, 5, 6}
		explicitBounds        = []float64{1, 2.5, 5, 7.5, 10}
		ts                    = time.Date(2020, 1, 2, 3, 4, 5, 6, time.UTC)
	)

	manager := &mockManager{}
	exp := remoteWriteExporter{
		manager:      manager,
		namespace:    "tempo_spanmetrics",
		promInstance: "tempo",
	}
	instance, _ := manager.GetInstance("tempo")
	app := instance.Appender(context.TODO())

	// Build data point
	dp := pdata.NewIntHistogramDataPoint()
	dp.SetTimestamp(pdata.TimestampFromTime(ts.UTC()))
	dp.SetBucketCounts(bucketCounts)
	dp.SetExplicitBounds(explicitBounds)
	dp.SetCount(countValue)
	dp.SetSum(sumValue)
	dps := pdata.NewIntHistogramDataPointSlice()
	dps.Append(dp)

	err := exp.handleHistogramIntDataPoints(app, "latency", dps)
	require.NoError(t, err)

	// Verify _sum
	sum := manager.instance.GetAppended(sumMetric)
	require.Equal(t, len(sum), 1)
	require.Equal(t, sum[0].v, float64(sumValue))
	require.Equal(t, sum[0].l, labels.Labels{{Name: nameLabelKey, Value: "tempo_spanmetrics_latency_" + sumSuffix}})

	// Check _count
	count := manager.instance.GetAppended(countMetric)
	require.Equal(t, len(count), 1)
	require.Equal(t, count[0].v, float64(countValue))
	require.Equal(t, count[0].l, labels.Labels{{Name: nameLabelKey, Value: "tempo_spanmetrics_latency_" + countSuffix}})

	// Check _bucket
	buckets := manager.instance.GetAppended(bucketMetric)
	require.Equal(t, len(buckets), len(bucketCounts))
	var bCount uint64
	for i, b := range buckets {
		bCount += bucketCounts[i]
		require.Equal(t, b.v, float64(bCount))
		eb := infBucket
		if len(explicitBounds) > i {
			eb = fmt.Sprint(explicitBounds[i])
		}
		require.Equal(t, b.l, labels.Labels{
			{Name: nameLabelKey, Value: "tempo_spanmetrics_latency_" + bucketSuffix},
			{Name: leStr, Value: eb},
		})
	}
}

type mockManager struct {
	instance *mockInstance
}

func (m *mockManager) GetInstance(name string) (instance.ManagedInstance, error) {
	if m.instance == nil {
		m.instance = &mockInstance{}
	}
	return m.instance, nil
}

func (m *mockManager) ListInstances() map[string]instance.ManagedInstance { return nil }

func (m *mockManager) ListConfigs() map[string]instance.Config { return nil }

func (m *mockManager) ApplyConfig(_ instance.Config) error { return nil }

func (m *mockManager) DeleteConfig(_ string) error { return nil }

func (m *mockManager) Stop() {}

type mockInstance struct {
	appender *mockAppender
}

func (m *mockInstance) Run(_ context.Context) error { return nil }

func (m *mockInstance) Update(_ instance.Config) error { return nil }

func (m *mockInstance) TargetsActive() map[string][]*scrape.Target { return nil }

func (m *mockInstance) StorageDirectory() string { return "" }

func (m *mockInstance) Appender(_ context.Context) storage.Appender {
	if m.appender == nil {
		m.appender = &mockAppender{}
	}
	return m.appender
}

func (m *mockInstance) GetAppended(n string) []metric {
	return m.appender.GetAppended(n)
}

type metric struct {
	l labels.Labels
	v float64
}

type mockAppender struct {
	appendedMetrics []metric
}

func (a *mockAppender) GetAppended(n string) []metric {
	var ms []metric
	for _, m := range a.appendedMetrics {
		if n == m.l.Get(nameLabelKey) {
			ms = append(ms, m)
		}
	}
	return ms
}

func (a *mockAppender) Append(_ uint64, l labels.Labels, _ int64, v float64) (uint64, error) {
	a.appendedMetrics = append(a.appendedMetrics, metric{l: l, v: v})
	return 0, nil
}

func (a *mockAppender) Commit() error { return nil }

func (a *mockAppender) Rollback() error { return nil }

func (a *mockAppender) AppendExemplar(_ uint64, _ labels.Labels, _ exemplar.Exemplar) (uint64, error) {
	return 0, nil
}
