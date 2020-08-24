package instance

import (
	"github.com/prometheus/client_golang/prometheus"
)

// CountingManager counts how many unique configs pass through a Manager.
// It may be distinct from the set of instances depending on available
// Managers.
//
// CountingManager implements Manager.
type CountingManager struct {
	currentActiveConfigs prometheus.Gauge

	cache map[string]struct{}
	inner Manager
}

// NewCountingManager creates a new CountingManager. The Manager provided
// by inner will be wrapped by CountingManager and will handle requests to
// apply configs.
func NewCountingManager(reg prometheus.Registerer, inner Manager) *CountingManager {
	currentActiveConfigs := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "agent_prometheus_active_configs",
		Help: "Current number of active configs being used by the agent.",
	})
	if reg != nil {
		reg.MustRegister(currentActiveConfigs)
	}

	return &CountingManager{
		currentActiveConfigs: currentActiveConfigs,
		cache:                make(map[string]struct{}),
		inner:                inner,
	}
}

// ListInstances implements Manager.
func (cm *CountingManager) ListInstances() map[string]ManagedInstance {
	return cm.inner.ListInstances()
}

// ListConfigs implements Manager.
func (cm *CountingManager) ListConfigs() map[string]Config {
	return cm.inner.ListConfigs()
}

// ApplyConfig implements Manager.
func (cm *CountingManager) ApplyConfig(c Config) error {
	err := cm.inner.ApplyConfig(c)
	if err != nil {
		return err
	}

	// If the config isn't in the cache, add it and increment the counter.
	if _, ok := cm.cache[c.Name]; !ok {
		cm.cache[c.Name] = struct{}{}
		cm.currentActiveConfigs.Inc()
	}

	return nil
}

// DeleteConfig implements Manager.
func (cm *CountingManager) DeleteConfig(name string) error {
	err := cm.inner.DeleteConfig(name)
	if err != nil {
		return err
	}

	// Remove the config from the cache and decrement the counter.
	delete(cm.cache, name)
	cm.currentActiveConfigs.Dec()
	return nil
}

// Stop implements Manager.
func (cm *CountingManager) Stop() {
	cm.inner.Stop()
}
