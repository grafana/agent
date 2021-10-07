package v1alpha1

import (
	prom_v1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IntegrationsSubsystemSpec defines global settings to apply across the logging
// subsystem.
type IntegrationsSubsystemSpec struct {
	// InstanceSelector determines which LogInstances should be selected
	// for running. Each instance runs its own set of Prometheus components,
	// including service discovery, scraping, and remote_write.
	InstanceSelector *metav1.LabelSelector `json:"instanceSelector,omitempty"`
	// InstanceNamespaceSelector are the set of labels to determine which
	// namespaces to watch for LogInstances. If not provided, only checks own
	// namespace.
	InstanceNamespaceSelector *metav1.LabelSelector `json:"instanceNamespaceSelector,omitempty"`
}

// +kubebuilder:validation:Enum=per-node;single

// AgentIntegrationMode defines how an integration should be run.
type AgentIntegrationMode string

const (
	// AgentIntegrationModePerNode is used for integrations that should be run on
	// every node (e.g., like a DaemonSet).
	AgentIntegrationModePerNode AgentIntegrationMode = "per-node"

	// AgentIntegrationModeSingle is used for integrations that only need one
	// instance for the whole cluster.
	AgentIntegrationModeSingle AgentIntegrationMode = "single"

	// AgentIntegrationModeDefault holds the default integration mode.
	AgentIntegrationModeDefault = AgentIntegrationModeSingle
)

// +kubebuilder:object:root=true
// +kubebuilder:resource:path="integrationinstances"
// +kubebuilder:resource:singular="integrationinstance"
// +kubebuilder:resource:categories="agent-operator"

// IntegrationInstance defines an integration for the Grafana Agent. The
// integration, when discovered by a GrafanaAgent, will be run as one or more
// pods on the cluster.
type IntegrationInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec holds the specification of the desired behavior for the integration.
	Spec IntegrationInstanceSpec `json:"spec,omitempty"`
}

// IntegrationInstanceSpec is a specification of the desired behavior for the
// integration.
type IntegrationInstanceSpec struct {
	// +kubebuilder:validation:Required

	// Name is the name of the integration this spec configures.
	// Must be an integration available in the GrafanaAgent version used.
	// Examples: node_exporter, agent, mysqld_exporter.
	Name string `json:"name"`

	// +kubebuilder:validation:Required
	// +kubebuilder:default:=single

	// Mode configures how this integration is run on the cluster. Integrations that
	// are related to metrics from machines should use "per-node" so that the integration
	// runs on every single machine. Otherwise, the default mode of "single" is
	// appropriate.
	Mode AgentIntegrationMode `json:"mode"`

	// +kubebuilder:validation:Required

	// Config holds the contents of the config for the integration.
	Config string `json:"config"`
}

// +kubebuilder:object:root=true

// IntegrationInstanceList is a list of IntegrationInstance.
type IntegrationInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	// Items holds the list of IntegrationInstance.
	Items []*IntegrationInstance `json:"items"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path="integrationmonitors"
// +kubebuilder:resource:singular="integrationmonitor"
// +kubebuilder:resource:categories="agent-operator"

// IntegrationMonitor defines monitoring for a set of IntegrationInstances.
type IntegrationMonitor struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Spec controls how IntegrationInstances are discovered by Prometheus for
	// metrics collection.
	Spec IntegrationMonitorSpec `json:"spec"`
}

// IntegrationMonitorSpec contains specification parameters for a ServiceMonitor.
// The job label is not configurable, as metrics must conform to having a specific
// job label value of "integrations/<integration name>".
type IntegrationMonitorSpec struct {
	// PodTargetLabels transfers labels on the Kubernetes Pod onto the target.
	PodTargetLabels []string `json:"podTargetLabels,omitempty"`
	// Selector to select Endpoints objects.
	Selector metav1.LabelSelector `json:"selector"`
	// Selector to select which namespaces the Endpoints objects are discovered from.
	NamespaceSelector prom_v1.NamespaceSelector `json:"namespaceSelector,omitempty"`
	// SampleLimit defines per-scrape limit on number of scraped samples that will be accepted.
	SampleLimit uint64 `json:"sampleLimit,omitempty"`
	// TargetLimit defines a limit on the number of scraped targets that will be accepted.
	TargetLimit uint64 `json:"targetLimit,omitempty"`
}

// +kubebuilder:object:root=true

// IntegrationMonitorList is a list of IntegrationMonitor.
type IntegrationMonitorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	// Items holds the list of IntegrationMonitor.
	Items []*IntegrationMonitor `json:"items"`
}
