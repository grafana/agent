package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IntegrationsSubsystemSpec defines global settings to apply across the
// integrations subsystem.
type IntegrationsSubsystemSpec struct {
	// Label selector to find integration resources (such as MetricsIntegration)
	// to run. When nil, no integration resources will be defined.
	Selector *metav1.LabelSelector `json:"selector,omitempty"`

	// Label selector for namespaces to search when discovering integration
	// resources. If nil, integration resources are only discovered in the
	// namespace of the GrafanaAgent resource.
	//
	// Set to `{}` to search all namespaces.
	NamespaceSelector *metav1.LabelSelector `json:"namespaceSelector,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path="metricsintegrations"
// +kubebuilder:resource:singular="metricsintegration"
// +kubebuilder:resource:categories="agent-operator"

// MetricsIntegration runs a single Grafana Agent integration which exposes
// metrics. MetricsIntegration can be used for any metrics-based integration.
type MetricsIntegration struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specifies the desired behavior of the metrics integration.
	Spec MetricsIntegrationSpec `json:"spec,omitempty"`
}

// MetricsIntegrationSpec specifies the desired behavior of a metrics
// integration.
type MetricsIntegrationSpec struct {
	// Name of the metrics-based integration to run (e.g., "node_exporter",
	// "mysqld_exporter").
	Name string `json:"name"`

	// +kubebuilder:default:=normal

	// Type informs the Grafana Agent Operator what the type of integration from
	// the "name" field is.
	//
	// Each integration exposed by Grafana Agent is one of three types:
	//
	// normal:    An integration that may exist any number of times per
	//            GrafanaAgent deployment. e.g., mysqld_exporter.
	//
	// unique:    An integration which must be unique per GrafanaAgent
	//            deployment. e.g., statsd_exporter.
	//
	// daemonset: An integration which must be unique per GrafanaAgent deployment
	//            and run on every Node in the Kubernetes cluster. e.g.,
	//            node_exporter.
	//
	// Grafana Agent Operator does not know the list of available integrations or
	// their types from the agent image being deployed. MetricsIntegrations must
	// be configured with the proper combination of name and type, otherwise the
	// Grafana Agent pod may fail to start.
	Type IntegrationType `json:"type"`

	// +kubebuilder:validation:Type=object

	// The configuration for the named integration.
	Config apiextv1.JSON `json:"config"`

	// An extra list of Volumes to be associated with the Grafana Agent pods
	// running this integration. Volume names will be mutated to be unique for an
	// agent deployment. Note that the specified volumes should be able to
	// tolerate existing on multiple pods at once when type is daemonset.
	Volumes []corev1.Volume `json:"volumes,omitempty"`

	// An extra list of VolumeMounts to be associated with the Grafana Agent pods
	// running this integration. VolumeMount names will be mutated to be unique
	// for an agent deployment.
	VolumeMounts []corev1.VolumeMount `json:"volumeMounts,omitempty"`

	// An extra list of Secret names in the same namespace as the
	// MetricsIntegration which will be mounted into the Grafana Agent pod
	// running this integration.
	//
	// Secrets will be mounted at
	// /etc/grafana-agent/integration-secrets/<secret_namespace>/<secret_name>.
	Secrets []string `json:"secrets,omitempty"`

	// An extra list of ConfigMaps in the same namespace as the
	// MetricsIntegration which will be mounted into the Grafana Agent pod
	// running this integration.
	//
	// ConfigMaps will be mounted at
	// /etc/grafana-agent/integration-secrets/<secret_namespace>/<secret_name>.
	ConfigMaps []string `json:"configMaps,omitempty"`
}

// +kubebuilder:validation:Enum=daemonset;singleton;normal

// IntegrationType specifies a type of an integration, influencing how it will
// run on a Kubernetes cluster.
type IntegrationType string

const (
	// IntegrationTypeDaemonset specifies that the named integration should run
	// on every Node in the Kubernetes cluster.
	IntegrationTypeDaemonset IntegrationType = "daemonset"

	// IntegrationTypeUnique specifies that the named integration exists exactly
	// once within a GrafanaAgent deployment.
	IntegrationTypeUnique IntegrationType = "unique"

	// IntegrationTypeNormal specifies that the named integration may exist any
	// number of times within a GrafanaAgent deployment.
	IntegrationTypeNormal IntegrationType = "normal"
)

// +kubebuilder:object:root=true

// MetricsIntegrationList is a list of MetricsIntegration.
type MetricsIntegrationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	// Items is the list of MetricsIntegration.
	Items []*MetricsIntegration `json:"items"`
}
