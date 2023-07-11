package scrape

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/pyroscope"
	"github.com/grafana/agent/pkg/util"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/discovery/targetgroup"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
	"go.uber.org/goleak"
)

func TestScrapePool(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start"))

	args := NewDefaultArguments()
	args.Targets = []discovery.Target{
		{"instance": "foo"},
	}
	args.ProfilingConfig.Block.Enabled = false
	args.ProfilingConfig.Goroutine.Enabled = false
	args.ProfilingConfig.Memory.Enabled = false

	p, err := newScrapePool(args, pyroscope.AppendableFunc(
		func(ctx context.Context, labels labels.Labels, samples []*pyroscope.RawSample) error {
			return nil
		}),
		util.TestLogger(t))
	require.NoError(t, err)

	defer p.stop()

	for _, tt := range []struct {
		name     string
		groups   []*targetgroup.Group
		expected []*Target
	}{
		{
			name:     "no targets",
			groups:   []*targetgroup.Group{},
			expected: []*Target{},
		},
		{
			name: "targets",
			groups: []*targetgroup.Group{
				{
					Targets: []model.LabelSet{
						{model.AddressLabel: "localhost:9090", serviceNameLabel: "s"},
						{model.AddressLabel: "localhost:8080", serviceNameK8SLabel: "k"},
					},
					Labels: model.LabelSet{"foo": "bar"},
				},
			},
			expected: []*Target{
				NewTarget(
					labels.FromStrings("instance", "localhost:8080", "foo", "bar", model.AddressLabel, "localhost:8080", model.MetricNameLabel, pprofMutex, model.SchemeLabel, "http", ProfilePath, "/debug/pprof/mutex", serviceNameLabel, "k"),
					labels.FromStrings("foo", "bar", model.AddressLabel, "localhost:8080", model.MetricNameLabel, pprofMutex, model.SchemeLabel, "http", ProfilePath, "/debug/pprof/mutex", serviceNameK8SLabel, "k"),
					url.Values{},
				),
				NewTarget(
					labels.FromStrings("instance", "localhost:8080", "foo", "bar", model.AddressLabel, "localhost:8080", model.MetricNameLabel, pprofProcessCPU, model.SchemeLabel, "http", ProfilePath, "/debug/pprof/profile", serviceNameLabel, "k"),
					labels.FromStrings("foo", "bar", model.AddressLabel, "localhost:8080", model.MetricNameLabel, pprofProcessCPU, model.SchemeLabel, "http", ProfilePath, "/debug/pprof/profile", serviceNameK8SLabel, "k"),
					url.Values{"seconds": []string{"14"}},
				),
				NewTarget(
					labels.FromStrings("instance", "localhost:9090", "foo", "bar", model.AddressLabel, "localhost:9090", model.MetricNameLabel, pprofMutex, model.SchemeLabel, "http", ProfilePath, "/debug/pprof/mutex", serviceNameLabel, "s"),
					labels.FromStrings("foo", "bar", model.AddressLabel, "localhost:9090", model.MetricNameLabel, pprofMutex, model.SchemeLabel, "http", ProfilePath, "/debug/pprof/mutex", serviceNameLabel, "s"),
					url.Values{},
				),
				NewTarget(
					labels.FromStrings("instance", "localhost:9090", "foo", "bar", model.AddressLabel, "localhost:9090", model.MetricNameLabel, pprofProcessCPU, model.SchemeLabel, "http", ProfilePath, "/debug/pprof/profile", serviceNameLabel, "s"),
					labels.FromStrings("foo", "bar", model.AddressLabel, "localhost:9090", model.MetricNameLabel, pprofProcessCPU, model.SchemeLabel, "http", ProfilePath, "/debug/pprof/profile", serviceNameLabel, "s"),
					url.Values{"seconds": []string{"14"}},
				),
			},
		},
		{
			name: "Remove targets",
			groups: []*targetgroup.Group{
				{
					Targets: []model.LabelSet{
						{model.AddressLabel: "localhost:9090", serviceNameLabel: "s"},
					},
				},
			},
			expected: []*Target{
				NewTarget(
					labels.FromStrings("instance", "localhost:9090", model.AddressLabel, "localhost:9090", model.MetricNameLabel, pprofMutex, model.SchemeLabel, "http", ProfilePath, "/debug/pprof/mutex", serviceNameLabel, "s"),
					labels.FromStrings(model.AddressLabel, "localhost:9090", model.MetricNameLabel, pprofMutex, model.SchemeLabel, "http", ProfilePath, "/debug/pprof/mutex", serviceNameLabel, "s"),
					url.Values{},
				),
				NewTarget(
					labels.FromStrings("instance", "localhost:9090", model.AddressLabel, "localhost:9090", model.MetricNameLabel, pprofProcessCPU, model.SchemeLabel, "http", ProfilePath, "/debug/pprof/profile", serviceNameLabel, "s"),
					labels.FromStrings(model.AddressLabel, "localhost:9090", model.MetricNameLabel, pprofProcessCPU, model.SchemeLabel, "http", ProfilePath, "/debug/pprof/profile", serviceNameLabel, "s"),
					url.Values{"seconds": []string{"14"}},
				),
			},
		},
		{
			name: "Sync targets",
			groups: []*targetgroup.Group{
				{
					Targets: []model.LabelSet{
						{model.AddressLabel: "localhost:9090", "__type__": "foo", serviceNameLabel: "s"},
					},
				},
			},
			expected: []*Target{
				NewTarget(
					labels.FromStrings("instance", "localhost:9090", model.AddressLabel, "localhost:9090", model.MetricNameLabel, pprofMutex, model.SchemeLabel, "http", ProfilePath, "/debug/pprof/mutex", serviceNameLabel, "s"),
					labels.FromStrings("__type__", "foo", model.AddressLabel, "localhost:9090", model.MetricNameLabel, pprofMutex, model.SchemeLabel, "http", ProfilePath, "/debug/pprof/mutex", serviceNameLabel, "s"),
					url.Values{},
				),
				NewTarget(
					labels.FromStrings("instance", "localhost:9090", model.AddressLabel, "localhost:9090", model.MetricNameLabel, pprofProcessCPU, model.SchemeLabel, "http", ProfilePath, "/debug/pprof/profile", serviceNameLabel, "s"),
					labels.FromStrings("__type__", "foo", model.AddressLabel, "localhost:9090", model.MetricNameLabel, pprofProcessCPU, model.SchemeLabel, "http", ProfilePath, "/debug/pprof/profile", serviceNameLabel, "s"),
					url.Values{"seconds": []string{"14"}},
				),
			},
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			p.sync(tt.groups)
			actual := p.ActiveTargets()
			sort.Sort(Targets(actual))
			sort.Sort(Targets(tt.expected))
			require.Equal(t, tt.expected, actual)
			require.Empty(t, p.DroppedTargets())
		})
	}

	// reload the cfg
	args.ScrapeTimeout = 1 * time.Second
	args.ScrapeInterval = 2 * time.Second
	p.reload(args)
	for _, ta := range p.activeTargets {
		require.Equal(t, 1*time.Second, ta.timeout)
		require.Equal(t, 2*time.Second, ta.interval)
	}
}

func TestScrapeLoop(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start"))

	down := atomic.NewBool(false)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The test was failing on Windows, as the scrape loop was too fast for
		// the Windows timer resolution.
		// This used to lead the `t.lastScrapeDuration = time.Since(start)` to
		// be recorded as zero. The small delay here allows the timer to record
		// the time since the last scrape properly.
		time.Sleep(2 * time.Millisecond)
		if down.Load() {
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.Write([]byte("ok"))
	}))
	defer server.Close()
	appendTotal := atomic.NewInt64(0)

	loop := newScrapeLoop(
		NewTarget(
			labels.FromStrings(
				model.SchemeLabel, "http",
				model.AddressLabel, strings.TrimPrefix(server.URL, "http://"),
				ProfilePath, "/debug/pprof/profile",
			), labels.FromStrings(), url.Values{
				"seconds": []string{"1"},
			}),
		server.Client(),
		pyroscope.AppendableFunc(func(_ context.Context, labels labels.Labels, samples []*pyroscope.RawSample) error {
			appendTotal.Inc()
			require.Equal(t, []byte("ok"), samples[0].RawProfile)
			return nil
		}),
		200*time.Millisecond, 30*time.Second, util.TestLogger(t))
	defer loop.stop(true)

	require.Equal(t, HealthUnknown, loop.Health())
	loop.start()
	require.Eventually(t, func() bool { return appendTotal.Load() > 3 }, 5000*time.Millisecond, 100*time.Millisecond)
	require.Equal(t, HealthGood, loop.Health())

	down.Store(true)
	require.Eventually(t, func() bool {
		return HealthBad == loop.Health()
	}, time.Second, 100*time.Millisecond)

	require.Error(t, loop.LastError())
	require.WithinDuration(t, time.Now(), loop.LastScrape(), 1*time.Second)
	require.NotEmpty(t, loop.LastScrapeDuration())
}
