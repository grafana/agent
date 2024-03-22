package kubelet

import (
	"github.com/prometheus/prometheus/discovery"
)

var _ discovery.DiscovererMetrics = (*kubeletMetrics)(nil)

type kubeletMetrics struct {
	refreshMetrics discovery.RefreshMetricsInstantiator
}

func newDiscovererMetrics(rmi discovery.RefreshMetricsInstantiator) discovery.DiscovererMetrics {
	m := &kubeletMetrics{
		refreshMetrics: rmi,
	}
	return m
}

// Register implements discovery.DiscovererMetrics.
func (m *kubeletMetrics) Register() error {
	return nil
}

// Unregister implements discovery.DiscovererMetrics.
func (m *kubeletMetrics) Unregister() {}
