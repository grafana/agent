package scrape

import (
	"net/url"
	"sort"
	"testing"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/discovery/targetgroup"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/stretchr/testify/require"
)

func Test_targetsFromGroup(t *testing.T) {
	args := NewDefaultArguments()
	args.ProfilingConfig.Block.Enabled = false
	args.ProfilingConfig.Goroutine.Enabled = false
	args.ProfilingConfig.Mutex.Enabled = false

	active, dropped, err := targetsFromGroup(&targetgroup.Group{
		Targets: []model.LabelSet{
			{model.AddressLabel: "localhost:9090"},
		},
		Labels: model.LabelSet{"foo": "bar"},
	}, args)
	expected := []*Target{
		NewTarget(
			labels.FromMap(map[string]string{
				model.AddressLabel:    "localhost:9090",
				model.MetricNameLabel: pprofMemory,
				ProfilePath:           "/debug/pprof/allocs",
				model.SchemeLabel:     "http",
				"foo":                 "bar",
				"instance":            "localhost:9090",
			}),
			labels.FromMap(map[string]string{
				model.AddressLabel:    "localhost:9090",
				model.MetricNameLabel: pprofMemory,
				ProfilePath:           "/debug/pprof/allocs",
				model.SchemeLabel:     "http",
				"foo":                 "bar",
			}),
			url.Values{}),
		NewTarget(
			labels.FromMap(map[string]string{
				model.AddressLabel:    "localhost:9090",
				model.MetricNameLabel: pprofProcessCPU,
				ProfilePath:           "/debug/pprof/profile",
				model.SchemeLabel:     "http",
				"foo":                 "bar",
				"instance":            "localhost:9090",
			}),
			labels.FromMap(map[string]string{
				model.AddressLabel:    "localhost:9090",
				model.MetricNameLabel: pprofProcessCPU,
				ProfilePath:           "/debug/pprof/profile",
				model.SchemeLabel:     "http",
				"foo":                 "bar",
			}),
			url.Values{"seconds": []string{"14"}}),
	}
	require.NoError(t, err)
	sort.Sort(Targets(active))
	sort.Sort(Targets(expected))
	require.Equal(t, expected, active)
	require.Empty(t, dropped)
}
