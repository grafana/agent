package scrape

import (
	"testing"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/pkg/river"
	"github.com/stretchr/testify/require"
)

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
				pprof_config = {
				   fgprof = {
						path = "/debug/fgprof",
						delta = true,
					   enabled = true,
				   },
				   block = { enabled = false },
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
				r.ProfilingConfig.PprofConfig["fgprof"] = &PprofProfilingConfig{
					Enabled: trueValue(),
					Path:    "/debug/fgprof",
					Delta:   true,
				}
				r.ProfilingConfig.PprofConfig["block"].Enabled = falseValue()
				r.ProfilingConfig.PprofPrefix = "v1/"
				return r
			},
		},
		"invalid cpu timeout": {
			in: `
			targets    = []
			forward_to = null
			scrape_timeout = "1s"
			`,
			expectedErr: "process_cpu scrape_timeout must be at least 2 seconds",
		},
		"invalid timeout/interval": {
			in: `
			targets    = []
			forward_to = null
			scrape_timeout = "4s"
			scrape_interval = "2s"
			`,
			expectedErr: "scrape timeout must be larger or equal to inverval",
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

func falseValue() *bool {
	a := false
	return &a
}
