package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

type WrappedRegisterer struct {
	mut                sync.RWMutex
	internalCollectors map[prometheus.Collector]struct{}
}

// NewWrappedRegisterer creates a wrapped register
func NewWrappedRegisterer() *WrappedRegisterer {
	return &WrappedRegisterer{internalCollectors: make(map[prometheus.Collector]struct{})}
}

// Describe implements the interface
func (w *WrappedRegisterer) Describe(descs chan<- *prometheus.Desc) {
	w.mut.RLock()
	defer w.mut.RUnlock()

	for c := range w.internalCollectors {
		c.Describe(descs)
	}
}

// Collect implements the interface
func (w *WrappedRegisterer) Collect(metrics chan<- prometheus.Metric) {
	w.mut.RLock()
	defer w.mut.RUnlock()

	for c := range w.internalCollectors {
		c.Collect(metrics)
	}
}

// Register implements the interface
func (w *WrappedRegisterer) Register(collector prometheus.Collector) error {
	w.mut.Lock()
	defer w.mut.Unlock()

	w.internalCollectors[collector] = struct{}{}
	return nil
}

// MustRegister implements the interface
func (w *WrappedRegisterer) MustRegister(collector ...prometheus.Collector) {
	w.mut.Lock()
	defer w.mut.Unlock()

	for _, c := range collector {
		w.internalCollectors[c] = struct{}{}
	}
}

// Unregister implements the interface
func (w *WrappedRegisterer) Unregister(collector prometheus.Collector) bool {
	w.mut.Lock()
	defer w.mut.Unlock()

	delete(w.internalCollectors, collector)
	return true
}
