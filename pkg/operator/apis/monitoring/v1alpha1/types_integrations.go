package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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
// +kubebuilder:resource:path="agentintegrations"
// +kubebuilder:resource:singular="agentintegration"
// +kubebuilder:resource:categories="agent-operator"

// AgentIntegration defines an integration for the Grafana Agent. The integration
// will have a single pod for the whole cluster.
type AgentIntegration struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec holds the specification of the desired behavior for the integration.
	Spec AgentIntegrationSpec `json:"spec,omitempty"`
}

// AgentIntegrationSpec is a specification of the desired behavior for the
// integration.
type AgentIntegrationSpec struct {
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

// AgentIntegrationList is a list of Integration.
type AgentIntegrationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	// Items holds the list of AgentIntegration.
	Items []*AgentIntegration `json:"items"`
}
