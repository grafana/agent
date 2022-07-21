package metrics

import (
	"fmt"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

// CollectorRegistry is both a prometheus.Registerer and prometheus.Collector:
// it encapsulates a set of collectors.
//
// Unlike a real Prometheus Registry, CollectorRegistry is unchecked and does
// not validate that metrics are unique at registration time.
type CollectorRegistry struct {
	mut sync.RWMutex
	cs  []prometheus.Collector
}

var _ prometheus.Registerer = (*CollectorRegistry)(nil)
var _ prometheus.Collector = (*CollectorRegistry)(nil)

// NewCollectorRegistry returns a new CollectorRegistry.
func NewCollectorRegistry() *CollectorRegistry {
	return &CollectorRegistry{}
}

// Register implements prometheus.Collector. Unlike a real Prometheus registry,
// Register does not ensure that c provides unique metrics.
func (cr *CollectorRegistry) Register(c prometheus.Collector) error {
	cr.mut.Lock()
	defer cr.mut.Unlock()

	for _, exist := range cr.cs {
		if exist == c {
			return fmt.Errorf("collector already registered")
		}
	}

	cr.cs = append(cr.cs, c)
	return nil
}

// MustRegister implements prometheus.Collector.
func (cr *CollectorRegistry) MustRegister(cs ...prometheus.Collector) {
	for _, c := range cs {
		if err := cr.Register(c); err != nil {
			panic(err)
		}
	}
}

// Unregister implements prometheus.Collector.
func (cr *CollectorRegistry) Unregister(c prometheus.Collector) bool {
	cr.mut.Lock()
	defer cr.mut.Unlock()

	rem := make([]prometheus.Collector, 0, len(cr.cs))

	var removed bool
	for _, exist := range cr.cs {
		if c == exist {
			removed = true
			continue
		}
		rem = append(rem, exist)
	}

	cr.cs = rem
	return removed
}

// Describe implements prometheus.Collector.
func (cr *CollectorRegistry) Describe(ch chan<- *prometheus.Desc) {
	cr.mut.RLock()
	defer cr.mut.RUnlock()

	for _, c := range cr.cs {
		c.Describe(ch)
	}
}

// Collect implements prometheus.Collector.
func (cr *CollectorRegistry) Collect(ch chan<- prometheus.Metric) {
	cr.mut.RLock()
	defer cr.mut.RUnlock()

	for _, c := range cr.cs {
		c.Collect(ch)
	}
}
