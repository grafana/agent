package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:resource:path="grafana-agents"
// +kubebuilder:resource:singular="grafana-agent"
// +kubebuilder:resource:categories="agent-operator"

// GrafanaAgent defines a Grafana Agent deployment.
type GrafanaAgent struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec holds the specification of the desired behavior for the Grafana Agent
	// cluster.
	Spec GrafanaAgentSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

// GrafanaAgentList is a list of GrafanaAgents.
type GrafanaAgentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	// Items is the list of GrafanaAgents.
	Items []*GrafanaAgent `json:"items"`
}

// GrafanaAgentSpec is a specification of the desired behavior of the Grafana
// Agent cluster.
type GrafanaAgentSpec struct {
	// TODO(rfratto): fields
}
