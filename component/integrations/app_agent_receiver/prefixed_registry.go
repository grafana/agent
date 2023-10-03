package app_agent_receiver

import (
	"errors"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

// prefixedRegistry is a prometheus.Registerer where the prefix can change at
// runtime.
type prefixedRegistry struct {
	mut             sync.RWMutex
	cs              map[prometheus.Collector]struct{}
	innerCollector  prometheus.Collector
	innerRegisterer prometheus.Registerer
}

var (
	_ prometheus.Registerer = (*prefixedRegistry)(nil)
	_ prometheus.Collector  = (*prefixedRegistry)(nil)
)

// newPrefixedRegistry creates a new prefixedRegistry.
func newPrefixedRegistry(prefix string) *prefixedRegistry {
	pr := &prefixedRegistry{
		cs: make(map[prometheus.Collector]struct{}),
	}

	// Call UpdatePrefix to initialize the inner collector and reigsterer.
	if err := pr.UpdatePrefix(prefix); err != nil {
		// The first call to UpdatePrefix should always succeed since there's no
		// collectors registered yet. If it fails, we should panic.
		panic(err)
	}

	return pr
}

// UpdatePrefix changes the prefix of the prefixedRegistry.
func (pr *prefixedRegistry) UpdatePrefix(prefix string) error {
	pr.mut.Lock()
	defer pr.mut.Unlock()

	var (
		innerCollector                        = prometheus.NewRegistry()
		innerRegisterer prometheus.Registerer = prometheus.WrapRegistererWithPrefix(prefix, innerCollector)
	)

	var errs []error

	// Register all known collectors to the new registerer.
	for collector := range pr.cs {
		errs = append(errs, innerRegisterer.Register(collector))
	}

	// Swap out the inner collector and register.
	pr.innerCollector = innerCollector
	pr.innerRegisterer = innerRegisterer

	return errors.Join(errs...)
}

func (pr *prefixedRegistry) Register(c prometheus.Collector) error {
	pr.mut.Lock()
	defer pr.mut.Unlock()

	if err := pr.innerRegisterer.Register(c); err != nil {
		return err
	}

	pr.cs[c] = struct{}{}
	return nil
}

func (pr *prefixedRegistry) MustRegister(cs ...prometheus.Collector) {
	pr.mut.Lock()
	defer pr.mut.Unlock()

	pr.innerRegisterer.MustRegister(cs...)

	for _, c := range cs {
		pr.cs[c] = struct{}{}
	}
}

func (pr *prefixedRegistry) Unregister(c prometheus.Collector) bool {
	pr.mut.Lock()
	defer pr.mut.Unlock()

	if !pr.innerRegisterer.Unregister(c) {
		return false
	}

	delete(pr.cs, c)
	return true
}

func (pr *prefixedRegistry) Describe(ch chan<- *prometheus.Desc) {
	// prefixedRegistry needs to be an unchecked collectors since the metric
	// names can change at runtime.
	//
	// Unchecked collectors don't implement Describe, so this is a no-op.
}

func (pr *prefixedRegistry) Collect(ch chan<- prometheus.Metric) {
	pr.mut.RLock()
	defer pr.mut.RUnlock()

	pr.innerCollector.Collect(ch)
}
