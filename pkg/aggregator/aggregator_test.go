package aggregator

import (
	"context"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/storage"
	"github.com/stretchr/testify/require"
)

func TestAggregator(t *testing.T) {
	m := mockAppendable{}
	a := New(log.NewNopLogger(), &m, []Rule{
		{
			Name:   "node_cpu_seconds_total",
			Expr:   "sum without(cpu) (node_cpu_seconds_total)",
			Labels: map[string]string{},
		},
	})

	appender := a.Appender(context.Background())
	samples := []sample{
		{labels.FromStrings(labels.MetricName, "node_cpu_seconds_total", "cpu", "0"), 0, 123},
		{labels.FromStrings(labels.MetricName, "node_cpu_seconds_total", "cpu", "1"), 0, 234},
		{labels.FromStrings(labels.MetricName, "node_cpu_seconds_total", "cpu", "2"), 0, 345},
		{labels.FromStrings(labels.MetricName, "node_cpu_seconds_total", "cpu", "3"), 0, 456},
	}
	for _, s := range samples {
		_, err := appender.Add(s.Labels(), s.t, s.v)
		require.NoError(t, err)
	}
	err := appender.Commit()
	require.NoError(t, err)

	expected := append(samples, sample{labels.FromStrings(labels.MetricName, "node_cpu_seconds_total"), 0, 1158})
	require.Equal(t, expected, m.samples)
}

type mockAppendable struct {
	samples []sample
}

func (m *mockAppendable) Appender(context.Context) storage.Appender {
	return m
}

func (m *mockAppendable) Add(l labels.Labels, t int64, v float64) (uint64, error) {
	m.samples = append(m.samples, sample{l, t, v})
	return 0, nil
}

func (m *mockAppendable) AddFast(ref uint64, t int64, v float64) error {
	return nil
}

func (m *mockAppendable) Commit() error {
	return nil
}

func (m *mockAppendable) Rollback() error {
	return nil
}
