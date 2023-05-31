package scrape

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/prometheus/scrape"
	"github.com/grafana/agent/component/pyroscope"
	"github.com/grafana/agent/pkg/cluster"
	"github.com/grafana/agent/pkg/river"
	"github.com/grafana/agent/pkg/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
	"go.uber.org/goleak"
)

func TestComponent(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start"))
	reloadInterval = 100 * time.Millisecond
	arg := NewDefaultArguments()
	arg.JobName = "test"
	c, err := New(component.Options{
		Logger:        util.TestFlowLogger(t),
		Registerer:    prometheus.NewRegistry(),
		OnStateChange: func(e component.Exports) {},
		Clusterer:     &cluster.Clusterer{Node: cluster.NewLocalNode("")},
	}, arg)
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		err := c.Run(ctx)
		require.NoError(t, err)
	}()

	// trigger an update
	require.Empty(t, c.appendable.Children())
	require.Empty(t, c.DebugInfo().(scrape.ScraperStatus).TargetStatus)

	arg.ForwardTo = []pyroscope.Appendable{pyroscope.NoopAppendable}
	arg.Targets = []discovery.Target{
		{
			model.AddressLabel: "foo",
		},
		{
			model.AddressLabel: "bar",
		},
	}
	c.Update(arg)

	require.Eventually(t, func() bool {
		fmt.Println(c.DebugInfo().(scrape.ScraperStatus).TargetStatus)
		return len(c.appendable.Children()) == 1 && len(c.DebugInfo().(scrape.ScraperStatus).TargetStatus) == 10
	}, 5*time.Second, 100*time.Millisecond)
}

func TestUnmarshalConfig(t *testing.T) {
	for name, tt := range map[string]struct {
		in          string
		expected    func() Arguments
		expectedErr string
	}{
		"default": {
			in: `
			targets    = [
				{"__address__" = "localhost:9090", "foo" = "bar"},
			]
			forward_to = null
		   `,
			expected: func() Arguments {
				r := NewDefaultArguments()
				r.Targets = []discovery.Target{
					{
						"__address__": "localhost:9090",
						"foo":         "bar",
					},
				}
				return r
			},
		},
		"custom": {
			in: `
			targets    = [
				{"__address__" = "localhost:9090", "foo" = "bar"},
				{"__address__" = "localhost:8080", "foo" = "buzz"},
			]
			forward_to = null
			profiling_config {
				path_prefix = "v1/"

				profile.block {
					enabled = false
				}

				profile.custom "something" {
					enabled = true
					path    = "/debug/fgprof"
					delta   = true
				}
		   }
		   `,
			expected: func() Arguments {
				r := NewDefaultArguments()
				r.Targets = []discovery.Target{
					{
						"__address__": "localhost:9090",
						"foo":         "bar",
					},
					{
						"__address__": "localhost:8080",
						"foo":         "buzz",
					},
				}
				r.ProfilingConfig.Block.Enabled = false
				r.ProfilingConfig.Custom = append(r.ProfilingConfig.Custom, CustomProfilingTarget{
					Enabled: true,
					Path:    "/debug/fgprof",
					Delta:   true,
					Name:    "something",
				})
				r.ProfilingConfig.PprofPrefix = "v1/"
				return r
			},
		},
		"invalid cpu timeout": {
			in: `
			targets    = []
			forward_to = null
			scrape_timeout = "1s"
			scrape_interval = "0.5s"
			`,
			expectedErr: "process_cpu scrape_timeout must be at least 2 seconds",
		},
		"invalid timeout/interval": {
			in: `
			targets    = []
			forward_to = null
			scrape_timeout = "4s"
			scrape_interval = "5s"
			`,
			expectedErr: "scrape_timeout must be greater than scrape_interval",
		},
		"invalid HTTPClientConfig": {
			in: `
			targets    = []
			forward_to = null
			scrape_timeout = "5s"
			scrape_interval = "1s"
			bearer_token = "token"
			bearer_token_file = "/path/to/file.token"
			`,
			expectedErr: "at most one of bearer_token & bearer_token_file must be configured",
		},
	} {
		tt := tt
		name := name
		t.Run(name, func(t *testing.T) {
			arg := Arguments{}
			if tt.expectedErr != "" {
				err := river.Unmarshal([]byte(tt.in), &arg)
				require.Error(t, err)
				require.Equal(t, tt.expectedErr, err.Error())
				return
			}
			require.NoError(t, river.Unmarshal([]byte(tt.in), &arg))
			require.Equal(t, tt.expected(), arg)
		})
	}
}

func TestUpdateWhileScraping(t *testing.T) {
	args := NewDefaultArguments()
	// speed up reload interval for this tests
	old := reloadInterval
	reloadInterval = 1 * time.Microsecond
	defer func() {
		reloadInterval = old
	}()
	args.ScrapeInterval = 1 * time.Second

	c, err := New(component.Options{
		Logger:        util.TestFlowLogger(t),
		Registerer:    prometheus.NewRegistry(),
		OnStateChange: func(e component.Exports) {},
		Clusterer:     &cluster.Clusterer{Node: cluster.NewLocalNode("")},
	}, args)
	require.NoError(t, err)
	scraping := atomic.NewBool(false)
	ctx, cancel := context.WithCancel(context.Background())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scraping.Store(true)
		select {
		case <-ctx.Done():
			return
		case <-time.After(15 * time.Second):
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	address := strings.TrimPrefix(server.URL, "http://")

	defer cancel()

	go c.Run(ctx)

	args.Targets = []discovery.Target{
		{
			model.AddressLabel: address,
			"foo":              "bar",
		},
		{
			model.AddressLabel: address,
			"foo":              "buz",
		},
	}

	c.Update(args)
	c.scraper.reload()
	// Wait for the targets to be scraping.
	require.Eventually(t, func() bool {
		return scraping.Load()
	}, 10*time.Second, 1*time.Second)

	// Send updates to the targets.
	done := make(chan struct{})
	go func() {
		for i := 0; i < 100; i++ {
			args.Targets = []discovery.Target{
				{
					model.AddressLabel: address,
					"foo":              fmt.Sprintf("%d", i),
				},
			}
			require.NoError(t, c.Update(args))
			c.scraper.reload()
		}
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("timed out waiting for updates to finish")
	}
}
