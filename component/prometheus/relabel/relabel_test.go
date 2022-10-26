package relabel

import (
	"math"
	"os"
	"testing"

	"github.com/go-kit/log"
	"github.com/grafana/agent/component"
	flow_relabel "github.com/grafana/agent/component/common/relabel"
	"github.com/grafana/agent/component/prometheus"
	"github.com/grafana/agent/pkg/util"
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/relabel"
	"github.com/prometheus/prometheus/model/value"
	"github.com/prometheus/prometheus/storage"
	"github.com/stretchr/testify/require"
)

func TestCache(t *testing.T) {
	relabeller := generateRelabel(t)
	lbls := labels.FromStrings("__address__", "localhost")
	relabeller.relabel(0, lbls)
	require.Len(t, relabeller.cache, 1)
	entry, found := relabeller.getFromCache(prometheus.GlobalRefMapping.GetOrAddGlobalRefID(lbls))
	require.True(t, found)
	require.NotNil(t, entry)
	require.True(
		t,
		prometheus.GlobalRefMapping.GetOrAddGlobalRefID(entry.labels) != prometheus.GlobalRefMapping.GetOrAddGlobalRefID(lbls),
	)
}

func TestEviction(t *testing.T) {
	relabeller := generateRelabel(t)
	lbls := labels.FromStrings("__address__", "localhost")
	relabeller.relabel(0, lbls)
	require.Len(t, relabeller.cache, 1)
	relabeller.relabel(math.Float64frombits(value.StaleNaN), lbls)
	require.Len(t, relabeller.cache, 0)
}

func TestUpdateReset(t *testing.T) {
	relabeller := generateRelabel(t)
	lbls := labels.FromStrings("__address__", "localhost")
	relabeller.relabel(0, lbls)
	require.Len(t, relabeller.cache, 1)
	_ = relabeller.Update(Arguments{
		MetricRelabelConfigs: []*flow_relabel.Config{},
	})
	require.Len(t, relabeller.cache, 0)
}

func TestNil(t *testing.T) {
	fanout := &prometheus.Fanout{
		Intercept: func(ref storage.SeriesRef, l labels.Labels, tt int64, v float64) (storage.SeriesRef, labels.Labels, int64, float64, error) {
			require.True(t, false)
			return ref, l, tt, v, nil
		},
	}
	relabeller, err := New(component.Options{
		ID:     "1",
		Logger: util.TestLogger(t),
		OnStateChange: func(e component.Exports) {
		},
		Registerer: prom.NewRegistry(),
	}, Arguments{
		ForwardTo: []storage.Appendable{fanout},
		MetricRelabelConfigs: []*flow_relabel.Config{
			{
				SourceLabels: []string{"__address__"},
				Regex:        flow_relabel.Regexp(relabel.MustNewRegexp("(.+)")),
				Action:       "drop",
			},
		},
	})
	require.NotNil(t, relabeller)
	require.NoError(t, err)

	lbls := labels.FromStrings("__address__", "localhost")
	relabeller.relabel(0, lbls)
}

func BenchmarkCache(b *testing.B) {
	l := log.NewSyncLogger(log.NewLogfmtLogger(os.Stderr))

	fanout := &prometheus.Fanout{
		Intercept: func(ref storage.SeriesRef, l labels.Labels, t int64, v float64) (storage.SeriesRef, labels.Labels, int64, float64, error) {
			if !l.Has("new_label") {
				panic("must have new label")
			}
			return ref, l, t, v, nil
		},
	}
	relabeller, _ := New(component.Options{
		ID:     "1",
		Logger: l,
		OnStateChange: func(e component.Exports) {
		},
		Registerer: prom.NewRegistry(),
	}, Arguments{
		ForwardTo: []storage.Appendable{fanout},
		MetricRelabelConfigs: []*flow_relabel.Config{
			{
				SourceLabels: []string{"__address__"},
				Regex:        flow_relabel.Regexp(relabel.MustNewRegexp("(.+)")),
				TargetLabel:  "new_label",
				Replacement:  "new_value",
				Action:       "replace",
			},
		},
	})

	for i := 0; i < b.N; i++ {
		lbls := labels.FromStrings("__address__", "localhost")
		relabeller.relabel(0, lbls)
	}
}

func generateRelabel(t *testing.T) *Component {
	fanout := &prometheus.Fanout{
		Intercept: func(ref storage.SeriesRef, l labels.Labels, tt int64, v float64) (storage.SeriesRef, labels.Labels, int64, float64, error) {
			require.True(t, l.Has("new_label"))
			return ref, l, tt, v, nil
		},
	}
	relabeller, err := New(component.Options{
		ID:     "1",
		Logger: util.TestLogger(t),
		OnStateChange: func(e component.Exports) {
		},
		Registerer: prom.NewRegistry(),
	}, Arguments{
		ForwardTo: []storage.Appendable{fanout},
		MetricRelabelConfigs: []*flow_relabel.Config{
			{
				SourceLabels: []string{"__address__"},
				Regex:        flow_relabel.Regexp(relabel.MustNewRegexp("(.+)")),
				TargetLabel:  "new_label",
				Replacement:  "new_value",
				Action:       "replace",
			},
		},
	})
	require.NotNil(t, relabeller)
	require.NoError(t, err)
	return relabeller
}
