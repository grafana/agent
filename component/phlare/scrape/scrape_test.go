package scrape

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/phlare"
	"github.com/grafana/agent/component/prometheus/scrape"
	"github.com/grafana/agent/pkg/river"
	"github.com/grafana/agent/pkg/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestComponent(t *testing.T) {
	defer goleak.VerifyNone(t)
	reloadInterval = 100 * time.Millisecond
	arg := NewDefaultArguments()
	arg.JobName = "test"
	c, err := New(component.Options{
		Logger:     util.TestLogger(t),
		Registerer: prometheus.NewRegistry(),
	}, arg)
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		err := c.Run(ctx)
		require.NoError(t, err)
	}()

	// triger an update
	require.Empty(t, c.appendable.Children())
	require.Empty(t, c.DebugInfo().(scrape.ScraperStatus).TargetStatus)

	arg.ForwardTo = []phlare.Appendable{phlare.NoopAppendable}
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
