package relabel

import (
	"fmt"
	"math"
	"os"
	"testing"
	"time"

	"golang.org/x/net/context"

	"github.com/go-kit/log"
	"github.com/grafana/agent/component"
	flow_relabel "github.com/grafana/agent/component/common/relabel"
	"github.com/grafana/agent/component/prometheus"
	"github.com/grafana/agent/pkg/flow/componenttest"
	"github.com/grafana/agent/pkg/river"
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
	fanout := prometheus.NewInterceptor(nil, prometheus.WithAppendHook(func(ref storage.SeriesRef, _ labels.Labels, _ int64, _ float64, _ storage.Appender) (storage.SeriesRef, error) {
		require.True(t, false)
		return ref, nil
	}))
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

	fanout := prometheus.NewInterceptor(nil, prometheus.WithAppendHook(func(ref storage.SeriesRef, l labels.Labels, _ int64, _ float64, _ storage.Appender) (storage.SeriesRef, error) {
		require.True(b, l.Has("new_label"))
		return ref, nil
	}))
	var entry storage.Appendable
	_, _ = New(component.Options{
		ID:     "1",
		Logger: l,
		OnStateChange: func(e component.Exports) {
			newE := e.(Exports)
			entry = newE.Receiver
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

	lbls := labels.FromStrings("__address__", "localhost")
	app := entry.Appender(context.Background())

	for i := 0; i < b.N; i++ {
		app.Append(0, lbls, time.Now().UnixMilli(), 0)
	}
	app.Commit()
}

func generateRelabel(t *testing.T) *Component {
	fanout := prometheus.NewInterceptor(nil, prometheus.WithAppendHook(func(ref storage.SeriesRef, l labels.Labels, _ int64, _ float64, _ storage.Appender) (storage.SeriesRef, error) {
		require.True(t, l.Has("new_label"))
		return ref, nil
	}))
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

func TestRuleGetter(t *testing.T) {
	// Set up the component Arguments.
	originalCfg := `rule {
         action       = "keep"
		 source_labels = ["__name__"]
         regex        = "up"
       }
		forward_to = []`
	var args Arguments
	require.NoError(t, river.Unmarshal([]byte(originalCfg), &args))

	// Set up and start the component.
	tc, err := componenttest.NewControllerFromID(nil, "prometheus.relabel")
	require.NoError(t, err)
	go func() {
		err = tc.Run(componenttest.TestContext(t), args)
		require.NoError(t, err)
	}()
	require.NoError(t, tc.WaitExports(time.Second))

	// Use the getter to retrieve the original relabeling rules.
	exports := tc.Exports().(Exports)
	fmt.Println("exports:", exports.Receiver, "exports.rules", exports.Rules)
	gotOriginal := exports.Rules()

	// Update the component with new relabeling rules and retrieve them.
	updatedCfg := `rule {
         action       = "drop"
		 source_labels = ["__name__"]
         regex        = "up"
       }
		forward_to = []`
	require.NoError(t, river.Unmarshal([]byte(updatedCfg), &args))

	require.NoError(t, tc.Update(args))
	gotUpdated := exports.Rules()

	require.NotEqual(t, gotOriginal, gotUpdated)
	require.Len(t, gotOriginal, 1)
	require.Len(t, gotUpdated, 1)

	require.Equal(t, gotOriginal[0].Action, relabel.Keep)
	require.Equal(t, gotUpdated[0].Action, relabel.Drop)
	require.Equal(t, gotUpdated[0].SourceLabels, gotOriginal[0].SourceLabels)
	require.Equal(t, gotUpdated[0].Regex, gotOriginal[0].Regex)
}
