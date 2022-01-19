package v1alpha1

import (
	"github.com/grafana/agent/pkg/operator/assets"
	prom_v1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

// Hierarchy is a hierarchy of resources discovered by a root
// GrafanaAgent resource.
type Hierarchy struct {
	// Root resource in the hierarchy.
	Agent *GrafanaAgent
	// Metrics resources within the hierarchy.
	Metrics []MetricsHierarchy
	// Logs resources within the hierarchy.
	Logs []LogsHierarchy
	// Integrations within the hierarchy.
	Integrations []*MetricsIntegration
	// Secrets loaded into the hierarchy.
	Secrets assets.SecretStore
}

// MetricsHierarchy is a hierarchy of resources discovered by a root
// MetricsInstance resource.
type MetricsHierarchy struct {
	Instance            *MetricsInstance
	ServiceMonitors     []*prom_v1.ServiceMonitor
	PodMonitors         []*prom_v1.PodMonitor
	Probes              []*prom_v1.Probe
	IntegrationMonitors []*IntegrationMonitor
}

// LogsHierarchy is a hierarchy of resources discovered by a root LogsInstance
// resource.
type LogsHierarchy struct {
	Instance *LogsInstance
	PodLogs  []*PodLogs
}
