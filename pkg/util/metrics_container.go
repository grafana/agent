package util

import "github.com/prometheus/client_golang/prometheus"

// MetricsContainer is a helper type useful for implementing
// prometheus.Collector on a struct which exposes a set of metrics.
type MetricsContainer struct {
	cs []prometheus.Collector
}

// Add adds a Collector into the container. Add should only be called when the
// container is being constructed.
func (mc *MetricsContainer) Add(cs ...prometheus.Collector) {
	mc.cs = append(mc.cs, cs...)
}

// Describe implements prometheus.Collector.
func (mc *MetricsContainer) Describe(ch chan<- *prometheus.Desc) {
	for _, c := range mc.cs {
		c.Describe(ch)
	}
}

// Collect implements prometheus.Collector.
func (mc *MetricsContainer) Collect(ch chan<- prometheus.Metric) {
	for _, c := range mc.cs {
		c.Collect(ch)
	}
}
