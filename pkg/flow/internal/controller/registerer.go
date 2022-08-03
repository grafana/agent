package controller

import (
	"fmt"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

// register handles a single components collector wrapping a registerer
type register struct {
	mut                sync.Mutex
	internal           prometheus.Registerer
	internalCollectors map[prometheus.Collector]struct{}
}

func newRegister(reg prometheus.Registerer) *register {
	return &register{
		internal:           reg,
		internalCollectors: make(map[prometheus.Collector]struct{}, 0),
	}
}

// RegisterComponent registers a set of collectors
func (r *register) RegisterComponent(collectors ...prometheus.Collector) error {
	r.mut.Lock()
	defer r.mut.Unlock()

	if r.internal == nil {
		return fmt.Errorf("internal registerer not set")
	}
	for _, c := range collectors {
		err := r.internal.Register(c)
		if err != nil {
			return err
		}
		r.internalCollectors[c] = struct{}{}
	}
	return nil
}

// UnregisterComponent unregisters all collectors from EITHER Register or RegisterComponent
func (r *register) UnregisterComponent() bool {
	r.mut.Lock()
	defer r.mut.Unlock()

	if r.internal == nil {
		return false
	}

	for coll := range r.internalCollectors {
		r.internal.Unregister(coll)
		delete(r.internalCollectors, coll)
	}
	return true
}

// Register registers a single collector
func (r *register) Register(collector prometheus.Collector) error {
	r.mut.Lock()
	defer r.mut.Unlock()

	if r.internal == nil {
		return fmt.Errorf("internal registerer not set")
	}
	err := r.internal.Register(collector)
	if err != nil {
		return err
	}
	r.internalCollectors[collector] = struct{}{}
	return nil
}

// MustRegister calls the internal register and adds the collectors to the internal collection
func (r *register) MustRegister(collector ...prometheus.Collector) {
	r.mut.Lock()
	defer r.mut.Unlock()

	r.internal.MustRegister(collector...)
	for _, c := range collector {
		r.internalCollectors[c] = struct{}{}
	}
}

// Unregister calls the internal unregister and removes the collector from the internal collection
func (r *register) Unregister(collector prometheus.Collector) bool {
	r.mut.Lock()
	defer r.mut.Unlock()

	if r.internal == nil {
		return false
	}
	delete(r.internalCollectors, collector)
	return r.internal.Unregister(collector)
}
