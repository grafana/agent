package relabel

import (
	"context"
	"math"
	"strconv"
	"testing"
	"time"

	"github.com/grafana/agent/component"
	flow_relabel "github.com/grafana/agent/component/common/relabel"
	"github.com/grafana/agent/component/prometheus"
	"github.com/grafana/agent/pkg/flow/componenttest"
	"github.com/grafana/agent/pkg/util"
	"github.com/grafana/agent/service/labelstore"
	"github.com/grafana/river"
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/relabel"
	"github.com/prometheus/prometheus/model/value"
	"github.com/prometheus/prometheus/storage"
	"github.com/stretchr/testify/require"
)

func TestCache(t *testing.T) {
	lc := labelstore.New(nil, prom.DefaultRegisterer)
	relabeller := generateRelabel(t)
	lbls := labels.FromStrings("__address__", "localhost")
	s := lc.ConvertToSeries(0, 0, lbls)
	relabeller.relabel(s)
	require.True(t, relabeller.cache.Len() == 1)
	entry, found := relabeller.getFromCache(lc.GetOrAddGlobalRefID(lbls))
	require.True(t, found)
	require.NotNil(t, entry)
	require.True(
		t,
		lc.GetOrAddGlobalRefID(entry.Lbls) != lc.GetOrAddGlobalRefID(lbls),
	)
}

func TestUpdateReset(t *testing.T) {
	ls := labelstore.New(nil, prom.DefaultRegisterer)
	relabeller := generateRelabel(t)
	lbls := labels.FromStrings("__address__", "localhost")
	s := ls.ConvertToSeries(0, 0, lbls)
	relabeller.relabel(s)
	require.True(t, relabeller.cache.Len() == 1)
	_ = relabeller.Update(Arguments{
		CacheSize:            100000,
		MetricRelabelConfigs: []*flow_relabel.Config{},
	})
	require.True(t, relabeller.cache.Len() == 0)
}

func TestValidator(t *testing.T) {
	args := Arguments{CacheSize: 0}
	err := args.Validate()
	require.Error(t, err)

	args.CacheSize = 1
	err = args.Validate()
	require.NoError(t, err)
}

func TestNil(t *testing.T) {
	ls := labelstore.New(nil, prom.DefaultRegisterer)
	fanout := prometheus.NewInterceptor(nil, ls, prometheus.WithAppendHook(func(s *labelstore.Series, _ labelstore.Appender) (storage.SeriesRef, error) {
		require.True(t, false)
		return storage.SeriesRef(s.GlobalID), nil
	}))
	relabeller, err := New(component.Options{
		ID:            "1",
		Logger:        util.TestFlowLogger(t),
		OnStateChange: func(e component.Exports) {},
		Registerer:    prom.NewRegistry(),
		GetServiceData: func(name string) (interface{}, error) {
			return labelstore.New(nil, prom.DefaultRegisterer), nil
		},
	}, Arguments{
		ForwardTo: []labelstore.Appendable{fanout},
		MetricRelabelConfigs: []*flow_relabel.Config{
			{
				SourceLabels: []string{"__address__"},
				Regex:        flow_relabel.Regexp(relabel.MustNewRegexp("(.+)")),
				Action:       "drop",
			},
		},
		CacheSize: 100000,
	})
	require.NotNil(t, relabeller)
	require.NoError(t, err)

	lbls := labels.FromStrings("__address__", "localhost")
	s := ls.ConvertToSeries(0, 0, lbls)
	relabeller.relabel(s)
}

func TestLRU(t *testing.T) {
	relabeller := generateRelabel(t)
	ls := labelstore.New(nil, prom.DefaultRegisterer)
	for i := 0; i < 600_000; i++ {
		lbls := labels.FromStrings("__address__", "localhost", "inc", strconv.Itoa(i))
		s := ls.ConvertToSeries(0, 0, lbls)
		relabeller.relabel(s)
	}
	require.True(t, relabeller.cache.Len() == 100_000)
}

func TestLRUNaN(t *testing.T) {
	ls := labelstore.New(nil, prom.DefaultRegisterer)
	relabeller := generateRelabel(t)
	lbls := labels.FromStrings("__address__", "localhost")
	relabeller.relabel(ls.ConvertToSeries(0, 0, lbls))
	require.True(t, relabeller.cache.Len() == 1)
	nan := ls.ConvertToSeries(0, math.Float64frombits(value.StaleNaN), lbls)
	relabeller.relabel(nan)
	require.True(t, relabeller.cache.Len() == 0)
}

func BenchmarkCache(b *testing.B) {
	ls := labelstore.New(nil, prom.DefaultRegisterer)
	fanout := prometheus.NewInterceptor(nil, ls, prometheus.WithAppendHook(func(s *labelstore.Series, _ labelstore.Appender) (storage.SeriesRef, error) {
		require.True(b, s.Lbls.Has("new_label"))
		return storage.SeriesRef(s.GlobalID), nil
	}))
	var entry labelstore.Appendable
	_, _ = New(component.Options{
		ID:     "1",
		Logger: util.TestFlowLogger(b),
		OnStateChange: func(e component.Exports) {
			newE := e.(Exports)
			entry = newE.Receiver
		},
		Registerer: prom.NewRegistry(),
	}, Arguments{
		ForwardTo: []labelstore.Appendable{fanout},
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
		s := ls.ConvertToSeries(time.Now().UnixMilli(), 0, lbls)
		app.Append(s)
	}
	app.Commit()
}

func generateRelabel(t *testing.T) *Component {
	ls := labelstore.New(nil, prom.DefaultRegisterer)
	fanout := prometheus.NewInterceptor(nil, ls, prometheus.WithAppendHook(func(s *labelstore.Series, _ labelstore.Appender) (storage.SeriesRef, error) {
		require.True(t, s.Lbls.Has("new_label"))
		return storage.SeriesRef(s.GlobalID), nil
	}))
	relabeller, err := New(component.Options{
		ID:            "1",
		Logger:        util.TestFlowLogger(t),
		OnStateChange: func(e component.Exports) {},
		Registerer:    prom.NewRegistry(),
		GetServiceData: func(name string) (interface{}, error) {
			return labelstore.New(nil, prom.DefaultRegisterer), nil
		},
	}, Arguments{
		ForwardTo: []labelstore.Appendable{fanout},
		MetricRelabelConfigs: []*flow_relabel.Config{
			{
				SourceLabels: []string{"__address__"},
				Regex:        flow_relabel.Regexp(relabel.MustNewRegexp("(.+)")),
				TargetLabel:  "new_label",
				Replacement:  "new_value",
				Action:       "replace",
			},
		},
		CacheSize: 100_000,
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
	gotOriginal := exports.Rules

	// Update the component with new relabeling rules and retrieve them.
	updatedCfg := `rule {
         action       = "drop"
		 source_labels = ["__name__"]
         regex        = "up"
       }
		forward_to = []`
	require.NoError(t, river.Unmarshal([]byte(updatedCfg), &args))

	require.NoError(t, tc.Update(args))
	exports = tc.Exports().(Exports)
	gotUpdated := exports.Rules

	require.NotEqual(t, gotOriginal, gotUpdated)
	require.Len(t, gotOriginal, 1)
	require.Len(t, gotUpdated, 1)

	require.Equal(t, gotOriginal[0].Action, flow_relabel.Keep)
	require.Equal(t, gotUpdated[0].Action, flow_relabel.Drop)
	require.Equal(t, gotUpdated[0].SourceLabels, gotOriginal[0].SourceLabels)
	require.Equal(t, gotUpdated[0].Regex, gotOriginal[0].Regex)
}
