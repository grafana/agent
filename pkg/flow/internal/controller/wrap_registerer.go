package controller

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

type wrappedRegisterer struct {
	mut                sync.RWMutex
	internalCollectors map[prometheus.Collector]struct{}
}

// newWrappedRegisterer creates a wrapped register
func newWrappedRegisterer() *wrappedRegisterer {
	return &wrappedRegisterer{internalCollectors: make(map[prometheus.Collector]struct{})}
}

// Describe implements the interface
func (w *wrappedRegisterer) Describe(descs chan<- *prometheus.Desc) {
	w.mut.RLock()
	defer w.mut.RUnlock()

	for c := range w.internalCollectors {
		c.Describe(descs)
	}
}

// Collect implements the interface
func (w *wrappedRegisterer) Collect(metrics chan<- prometheus.Metric) {
	w.mut.RLock()
	defer w.mut.RUnlock()

	for c := range w.internalCollectors {
		c.Collect(metrics)
	}
}

// Register implements the interface
func (w *wrappedRegisterer) Register(collector prometheus.Collector) error {
	w.mut.Lock()
	defer w.mut.Unlock()

	w.internalCollectors[collector] = struct{}{}
	return nil
}

// MustRegister implements the interface
func (w *wrappedRegisterer) MustRegister(collector ...prometheus.Collector) {
	w.mut.Lock()
	defer w.mut.Unlock()

	for _, c := range collector {
		w.internalCollectors[c] = struct{}{}
	}
}

// Unregister implements the interface
func (w *wrappedRegisterer) Unregister(collector prometheus.Collector) bool {
	w.mut.Lock()
	defer w.mut.Unlock()

	delete(w.internalCollectors, collector)
	return true
}
