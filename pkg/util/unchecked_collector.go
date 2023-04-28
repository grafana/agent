package util

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

// UncheckedCollector is a prometheus.Collector which stores a set of unchecked
// metrics.
type UncheckedCollector struct {
	mut   sync.RWMutex
	inner prometheus.Collector
}

var _ prometheus.Collector = (*UncheckedCollector)(nil)

// NewUncheckedCollector creates a new UncheckedCollector. If inner is nil,
// UncheckedCollector returns no metrics.
func NewUncheckedCollector(inner prometheus.Collector) *UncheckedCollector {
	return &UncheckedCollector{inner: inner}
}

// SetCollector replaces the inner collector.
func (uc *UncheckedCollector) SetCollector(inner prometheus.Collector) {
	uc.mut.Lock()
	defer uc.mut.Unlock()

	uc.inner = inner
}

// Describe implements [prometheus.Collector]. Because UncheckedCollector is
// unchecked, nothing is written to the provided ch.
func (uc *UncheckedCollector) Describe(ch chan<- *prometheus.Desc) {
	// no-op: do not send any descriptions of metrics to avoid having them be
	// checked.
}

// Collector implements [prometheus.Collector]. If the UncheckedCollector has a
// non-nil inner collector, metrics will be collected from it.
func (uc *UncheckedCollector) Collect(ch chan<- prometheus.Metric) {
	uc.mut.RLock()
	defer uc.mut.RUnlock()

	if uc.inner != nil {
		uc.inner.Collect(ch)
	}
}
