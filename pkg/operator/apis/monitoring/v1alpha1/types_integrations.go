package v1alpha1

import (
	prom_v1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IntegrationsSubsystemSpec defines global settings to apply across the logging
// subsystem.
type IntegrationsSubsystemSpec struct {
	// InstanceSelector determines which integrations should be run based on
	// labels.
	InstanceSelector *metav1.LabelSelector `json:"instanceSelector,omitempty"`
	// InstanceNamespaceSelector are the set of labels to determine which
	// namespaces to watch for integrations. If not provided, only checks own
	// namespace.
	InstanceNamespaceSelector *metav1.LabelSelector `json:"instanceNamespaceSelector,omitempty"`
}

// +kubebuilder:validation:Enum=daemonset;singleton;normal

// AgentIntegrationType defines the type of integration. Supported values:
//
// daemonset: integrations which run on every Kubernetes Node. These are
// integrations which collect machine level metrics, like node_exporter or
// cadvisor.
//
// singleton: integrations which run once per GrafanaAgent CRD. These are
// integrations like statsd_exporter, where you want a single place to send
// statsd metrics.
//
// normal: all other integrations.
//
// The default value is normal.
type AgentIntegrationType string

const (
	// AgentIntegrationModeDaemonSet is used for integrations that should be run
	// on every Kubernetes Node. daemonset integrations must be unique per
	// GrafanaAgent deployment.
	//
	// Example daemonset integrations: node_exporter, cadvisor.
	AgentIntegrationModeDaemonSet AgentIntegrationType = "daemonset"

	// AgentIntegrationModeSingleton is used for integrations that should be run
	// once per GrafanaAgent. singleton integrations must be unique per
	// GrafanaAgent deployment.
	//
	// Example singleton integrations: stats_exporter.
	AgentIntegrationModeSingleton AgentIntegrationType = "singleton"

	// AgentIntegrationModeNormal is used for integrations that can be defined
	// multiple times per GrafanaAgent. Integrations that aren't daemonset or
	// singleton fall under this category.
	AgentIntegrationModeNormal AgentIntegrationType = "normal"

	// AgentIntegrationModeDefault holds the default integration mode.
	AgentIntegrationModeDefault = AgentIntegrationModeNormal
)

// +kubebuilder:object:root=true
// +kubebuilder:resource:path="integrationinstances"
// +kubebuilder:resource:singular="integrationinstance"
// +kubebuilder:resource:categories="agent-operator"

// MetricsIntegrationInstance defines an integration for the Grafana Agent. The
// integration, when discovered by a GrafanaAgent, will be run as one or more
// pods on the cluster.
type MetricsIntegrationInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec holds the specification of the desired behavior for the integration.
	Spec MetricsIntegrationInstanceSpec `json:"spec,omitempty"`
}

// MetricsIntegrationInstanceSpec is a specification of the desired behavior for the
// integration.
type MetricsIntegrationInstanceSpec struct {
	// +kubebuilder:validation:Required

	// Name is the name of the integration this spec configures.
	// Must be an integration available in the GrafanaAgent version used.
	// Examples: node_exporter, agent, mysqld_exporter.
	Name string `json:"name"`

	// +kubebuilder:validation:Required
	// +kubebuilder:default:=normal

	// Type specifies the type of integration being configured.
	Type AgentIntegrationType `json:"type"`

	// +kubebuilder:validation:Required

	// Config holds the contents of the config for the integration.
	Config string `json:"config"`
}

// +kubebuilder:object:root=true

// MetricsIntegrationInstanceList is a list of IntegrationInstance.
type MetricsIntegrationInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	// Items holds the list of IntegrationInstance.
	Items []*MetricsIntegrationInstance `json:"items"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path="metricsintegrationmonitors"
// +kubebuilder:resource:singular="metricsintegrationmonitor"
// +kubebuilder:resource:categories="agent-operator"

// MetricsIntegrationMonitor defines monitoring for a metrics-based
// integration.
type MetricsIntegrationMonitor struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Spec controls how IntegrationInstances are discovered by Prometheus for
	// metrics collection.
	Spec MetricsIntegrationMonitorSpec `json:"spec"`
}

// MetricsIntegrationMonitorSpec contains specification parameters for an
// MetricsIntegrationMonitor. The job label is not configurable, as metrics
// must conform to having a specific job label value of
// "integrations/<integration name>".
type MetricsIntegrationMonitorSpec struct {
	// PodTargetLabels transfers labels on the Kubernetes Pod onto the target.
	PodTargetLabels []string `json:"podTargetLabels,omitempty"`
	// Selector to select IntegrationInstance objects.
	Selector metav1.LabelSelector `json:"selector"`
	// Selector to select which namespaces the IntegrationInstance objects are discovered from.
	NamespaceSelector prom_v1.NamespaceSelector `json:"namespaceSelector,omitempty"`
	// Interval at which metrics should be scraped
	Interval string `json:"interval,omitempty"`
	// Timeout after which the scrape is ended
	ScrapeTimeout string `json:"scrapeTimeout,omitempty"`
	// HonorLabels chooses the metric's labels on collisions with target labels.
	HonorLabels bool `json:"honorLabels,omitempty"`
	// HonorTimestamps controls whether Prometheus respects the timestamps present in scraped data.
	HonorTimestamps *bool `json:"honorTimestamps,omitempty"`
	// MetricRelabelConfigs to apply to samples before ingestion.
	MetricRelabelConfigs []*prom_v1.RelabelConfig `json:"metricRelabelings,omitempty"`
	// RelabelConfigs to apply to samples before scraping.
	// Prometheus Operator automatically adds relabelings for a few standard Kubernetes fields
	// and replaces original scrape job name with __tmp_prometheus_job_name.
	// More info: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#relabel_config
	RelabelConfigs []*prom_v1.RelabelConfig `json:"relabelings,omitempty"`
	// ProxyURL eg http://proxyserver:2195 Directs scrapes to proxy through this endpoint.
	ProxyURL *string `json:"proxyUrl,omitempty"`
	// SampleLimit defines per-scrape limit on number of scraped samples that will be accepted.
	SampleLimit uint64 `json:"sampleLimit,omitempty"`
	// TargetLimit defines a limit on the number of scraped targets that will be accepted.
	TargetLimit uint64 `json:"targetLimit,omitempty"`
}

// +kubebuilder:object:root=true

// MetricsIntegrationMonitorList is a list of MetricsIntegrationMonitor.
type MetricsIntegrationMonitorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	// Items holds the list of IntegrationMonitor.
	Items []*MetricsIntegrationMonitor `json:"items"`
}
