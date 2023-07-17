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
			{model.AddressLabel: "localhost:9091", serviceNameLabel: "svc"},
			{model.AddressLabel: "localhost:9092", serviceNameK8SLabel: "k8s-svc"},
			{model.AddressLabel: "localhost:9093", "__meta_kubernetes_namespace": "ns", "__meta_kubernetes_pod_container_name": "container"},
			{model.AddressLabel: "localhost:9094", "__meta_docker_container_name": "docker-container"},
		},
		Labels: model.LabelSet{
			"foo": "bar",
		},
	}, args)
	expected := []*Target{
		// unspecified
		NewTarget(
			labels.FromMap(map[string]string{
				model.AddressLabel:    "localhost:9090",
				serviceNameLabel:      "unspecified",
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
				serviceNameLabel:      "unspecified",
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

		//specified
		NewTarget(
			labels.FromMap(map[string]string{
				model.AddressLabel:    "localhost:9091",
				serviceNameLabel:      "svc",
				model.MetricNameLabel: pprofMemory,
				ProfilePath:           "/debug/pprof/allocs",
				model.SchemeLabel:     "http",
				"foo":                 "bar",
				"instance":            "localhost:9091",
			}),
			labels.FromMap(map[string]string{
				model.AddressLabel:    "localhost:9091",
				serviceNameLabel:      "svc",
				model.MetricNameLabel: pprofMemory,
				ProfilePath:           "/debug/pprof/allocs",
				model.SchemeLabel:     "http",
				"foo":                 "bar",
			}),
			url.Values{}),
		NewTarget(
			labels.FromMap(map[string]string{
				model.AddressLabel:    "localhost:9091",
				serviceNameLabel:      "svc",
				model.MetricNameLabel: pprofProcessCPU,
				ProfilePath:           "/debug/pprof/profile",
				model.SchemeLabel:     "http",
				"foo":                 "bar",
				"instance":            "localhost:9091",
			}),
			labels.FromMap(map[string]string{
				model.AddressLabel:    "localhost:9091",
				serviceNameLabel:      "svc",
				model.MetricNameLabel: pprofProcessCPU,
				ProfilePath:           "/debug/pprof/profile",
				model.SchemeLabel:     "http",
				"foo":                 "bar",
			}),
			url.Values{"seconds": []string{"14"}}),

		// k8s annotation specified
		NewTarget(
			labels.FromMap(map[string]string{
				model.AddressLabel:    "localhost:9092",
				serviceNameLabel:      "k8s-svc",
				model.MetricNameLabel: pprofMemory,
				ProfilePath:           "/debug/pprof/allocs",
				model.SchemeLabel:     "http",
				"foo":                 "bar",
				"instance":            "localhost:9092",
			}),
			labels.FromMap(map[string]string{
				model.AddressLabel:    "localhost:9092",
				serviceNameK8SLabel:   "k8s-svc",
				model.MetricNameLabel: pprofMemory,
				ProfilePath:           "/debug/pprof/allocs",
				model.SchemeLabel:     "http",
				"foo":                 "bar",
			}),
			url.Values{}),
		NewTarget(
			labels.FromMap(map[string]string{
				model.AddressLabel:    "localhost:9092",
				serviceNameLabel:      "k8s-svc",
				model.MetricNameLabel: pprofProcessCPU,
				ProfilePath:           "/debug/pprof/profile",
				model.SchemeLabel:     "http",
				"foo":                 "bar",
				"instance":            "localhost:9092",
			}),
			labels.FromMap(map[string]string{
				model.AddressLabel:    "localhost:9092",
				serviceNameK8SLabel:   "k8s-svc",
				model.MetricNameLabel: pprofProcessCPU,
				ProfilePath:           "/debug/pprof/profile",
				model.SchemeLabel:     "http",
				"foo":                 "bar",
			}),
			url.Values{"seconds": []string{"14"}}),

		// unspecified, infer from k8s
		NewTarget(
			labels.FromMap(map[string]string{
				model.AddressLabel:    "localhost:9093",
				serviceNameLabel:      "ns/container",
				model.MetricNameLabel: pprofMemory,
				ProfilePath:           "/debug/pprof/allocs",
				model.SchemeLabel:     "http",
				"foo":                 "bar",
				"instance":            "localhost:9093",
			}),
			labels.FromMap(map[string]string{
				model.AddressLabel:                     "localhost:9093",
				"__meta_kubernetes_namespace":          "ns",
				"__meta_kubernetes_pod_container_name": "container",
				model.MetricNameLabel:                  pprofMemory,
				ProfilePath:                            "/debug/pprof/allocs",
				model.SchemeLabel:                      "http",
				"foo":                                  "bar",
			}),
			url.Values{}),
		NewTarget(
			labels.FromMap(map[string]string{
				model.AddressLabel:    "localhost:9093",
				serviceNameLabel:      "ns/container",
				model.MetricNameLabel: pprofProcessCPU,
				ProfilePath:           "/debug/pprof/profile",
				model.SchemeLabel:     "http",
				"foo":                 "bar",
				"instance":            "localhost:9093",
			}),
			labels.FromMap(map[string]string{
				model.AddressLabel:                     "localhost:9093",
				"__meta_kubernetes_namespace":          "ns",
				"__meta_kubernetes_pod_container_name": "container",
				model.MetricNameLabel:                  pprofProcessCPU,
				ProfilePath:                            "/debug/pprof/profile",
				model.SchemeLabel:                      "http",
				"foo":                                  "bar",
			}),
			url.Values{"seconds": []string{"14"}}),

		// unspecified, infer from docker
		NewTarget(
			labels.FromMap(map[string]string{
				model.AddressLabel:    "localhost:9094",
				serviceNameLabel:      "docker-container",
				model.MetricNameLabel: pprofMemory,
				ProfilePath:           "/debug/pprof/allocs",
				model.SchemeLabel:     "http",
				"foo":                 "bar",
				"instance":            "localhost:9094",
			}),
			labels.FromMap(map[string]string{
				model.AddressLabel:             "localhost:9094",
				"__meta_docker_container_name": "docker-container",
				model.MetricNameLabel:          pprofMemory,
				ProfilePath:                    "/debug/pprof/allocs",
				model.SchemeLabel:              "http",
				"foo":                          "bar",
			}),
			url.Values{}),
		NewTarget(
			labels.FromMap(map[string]string{
				model.AddressLabel:    "localhost:9094",
				serviceNameLabel:      "docker-container",
				model.MetricNameLabel: pprofProcessCPU,
				ProfilePath:           "/debug/pprof/profile",
				model.SchemeLabel:     "http",
				"foo":                 "bar",
				"instance":            "localhost:9094",
			}),
			labels.FromMap(map[string]string{
				model.AddressLabel:             "localhost:9094",
				"__meta_docker_container_name": "docker-container",
				model.MetricNameLabel:          pprofProcessCPU,
				ProfilePath:                    "/debug/pprof/profile",
				model.SchemeLabel:              "http",
				"foo":                          "bar",
			}),
			url.Values{"seconds": []string{"14"}}),
	}
	require.NoError(t, err)
	sort.Sort(Targets(active))
	sort.Sort(Targets(expected))
	require.Equal(t, expected, active)
	require.Empty(t, dropped)
}
