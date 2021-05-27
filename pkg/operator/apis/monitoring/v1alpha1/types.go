package v1alpha1

import (
	prom_v1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	v1 "k8s.io/api/core/v1"
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
	// LogLevel controls the log level of the generated pods. Defaults to "info" if not set.
	LogLevel string `json:"logLevel,omitempty"`
	// LogFormat controls the logging format of the generated pods. Defaults to "logfmt" if not set.
	LogFormat string `json:"logFormat,omitempty"`
	// APIServerConfig allows specifying a host and auth methods to access the
	// Kubernetes API server. If left empty, the Agent will assume that it is
	// running inside of the cluster and will discover API servers automatically
	// and use the pod's CA certificate and bearer token file at
	// /var/run/secrets/kubernetes.io/serviceaccount.
	APIServerConfig *prom_v1.APIServerConfig `json:"apiServer,omitempty"`
	// PodMetadata configures Labels and Annotations which are propagated to
	// created Grafana Agent pods.
	PodMetadata *EmbeddedObjectMetadata `json:"podMetadata,omitempty"`
	// Version of Grafana Agent to be deployed.
	Version string `json:"version,omitempty"`
	// Paused prevents actions except for deletion to be performed on the
	// underlying managed objects.
	Paused bool `json:"paused,omitempty"`
	// Image, when specified, overrides the image used to run the Agent. It
	// should be specified along with a tag. Version must still be set to ensure
	// the Grafana Agent Operator knows which version of Grafana Agent is being
	// configured.
	Image *string `json:"image,omitempty"`
	// ImagePullSecrets holds an optional list of references to secrets within
	// the same namespace to use for pulling the Grafana Agent image from
	// registries.
	// More info: https://kubernetes.io/docs/user-guide/images#specifying-imagepullsecrets-on-a-pod
	ImagePullSecrets []v1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
	// Storage spec to specify how storage will be used.
	Storage *prom_v1.StorageSpec `json:"storage,omitempty"`
	// Volumes allows configuration of additional volumes on the output
	// StatefulSet definition. Volumes specified will be appended to other
	// volumes that are generated as a result of StorageSpec objects.
	Volumes []v1.Volume `json:"volumes,omitempty"`
	// VolumeMounts allows configuration of additional VolumeMounts on the output
	// StatefulSet definition. VolumEMounts specified will be appended to other
	// VolumeMounts in the Grafana Agent container that are generated as a result
	// of StorageSpec objects.
	VolumeMounts []v1.VolumeMount `json:"volumeMounts,omitempty"`
	// Resources holds requests and limits for individual pods.
	Resources v1.ResourceRequirements `json:"resources,omitempty"`
	// NodeSelector defines which nodes pods should be scheduling on.
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// ServiceAccountName is the name of the ServiceAccount to use for running Grafana Agent pods.
	ServiceAccountName string `json:"serviceAccountName,omitempty"`
	// Secrets is a list of secrets in the same namespace as the GrafanaAgent
	// object which will be mounted into each running Grafana Agent pod.
	// The secrets are mounted into /etc/grafana-agent/secrets/<secret-name>.
	Secrets []string `json:"secrets,omitempty"`
	// ConfigMaps is a liset of config maps in the same namespace as the
	// GrafanaAgent object which will be mounted into each running Grafana Agent
	// pod.
	// The secrets are mounted into /etc/grafana-agent/configmaps/<configmap-name>.
	ConfigMaps []string `json:"configMaps,omitempty"`
	// Affinity, if specified, controls pod scheduling constraints.
	Affinity *v1.Affinity `json:"affinity,omitempty"`
	// Tolerations, if specified, controls the pod's tolerations.
	Tolerations []v1.Toleration `json:"tolerations,omitempty"`
	// TopologySpreadConstraints, if specified, controls the pod's topology spread constraints.
	TopologySpreadConstraints []v1.TopologySpreadConstraint `json:"topologySpreadConstraints,omitempty"`
	// SecurityContext holds pod-level security attributes and common container
	// settings. When unspecified, defaults to the default PodSecurityContext.
	SecurityContext *v1.PodSecurityContext `json:"securityContext,omitempty"`
	// Containers allows injecting additional containers or modifying operator
	// generated containers. This can be used to allow adding an authentication
	// proxy to a Grafana Agent pod or to change the behavior of an
	// operator-generated container. Containers described here modify an operator
	// generated container if they share the same name and modifications are done
	// via a strategic merge patch. The current container names are:
	// `grafana-agent` and `config-reloader`. Overriding containers is entirely
	// outside the scope of what the Grafana Agent team will support and by doing
	// so, you accept that this behavior may break at any time without notice.
	Containers []v1.Container `json:"containers,omitempty"`
	// InitContainers allows adding initContainers to the pod definition. These
	// can be used to, for example, fetch secrets for injection into the Grafana
	// Agent configuration from external sources. Any errors during the execution
	// of an initContainer will lead to a restart of the pod.
	// More info: https://kubernetes.io/docs/concepts/workloads/pods/init-containers/
	// Using initContainers for any use case other than secret fetching is
	// entirely outside the scope of what the Grafana Agent maintainers will
	// support and by doing so, you accept that this behavior may break at any
	// time without notice.
	InitContainers []v1.Container `json:"initContainers,omitempty"`
	// PriorityClassName is the priority class assigned to pods.
	PriorityClassName string `json:"priorityClassName,omitempty"`
	// Port name used for the pods and governing service. This defaults to agent-metrics.
	PortName string `json:"portName,omitempty"`
	// Prometheus controls the Prometheus subsystem of the Agent and settings
	// unique to Prometheus-specific pods that are deployed.
	Prometheus PrometheusSubsystemSpec `json:"prometheus,omitempty"`
}

// EmbeddedObjectMetadata contains a subset of fields included in
// k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta. Only fields which are
// relevant to embedded resources are included.
type EmbeddedObjectMetadata struct {
	// Name must be unique within a namespace. Required when creating resources,
	// although some resources may allow a client to request the generation of an
	// appropriate name automatically. Name is primarily intended for creation
	// idempotence and configuration definition.
	// Cannot be updated.
	// More info: https://kubernetes.io/docs/user-guide/identifiers#names
	Name string `json:"name,omitempty"`

	// Labels holds a map of string keys and values that can be used to organize
	// and categorize (scope and select) objects. May match selectors of
	// replication controllers and services.
	// More info: https://kubernetes.io/docs/user-guide/labels
	Labels map[string]string `json:"labels,omitempty"`

	// Annotations is an unstructured key value map stored with a resource that
	// may be set by external tools to store and retrieve arbitrary metadata.
	// They are not queryable and should be preserved when modifying objects.
	// More info: https://kubernetes.io/docs/user-guide/annotations
	Annotations map[string]string `json:"annotations,omitempty"`
}

// StorageSpec defines the configured storage for a group of Grfana Agents.
// If neither `emptyDir` nor `volumeClaimTemplate` is specified, then by
// default an
// [EmptyDir](https://kubernetes.io/docs/concepts/storage/volumes/#emptydir)
// will be used.
type StorageSpec struct {
	// EmptyDir to be used by the Grafana Agent StatefulSets. If specified, used in place of any volumeClaimTemplate.
	// More info: https://kubernetes.io/docs/concepts/storage/volumes/#emptydir
	EmptyDir *v1.EmptyDirVolumeSource `json:"emptyDir,omitempty"`
	// VolumeClaimTemplate is a PVC spec to be used by the GrafanaAgent StatefulSets.
	VolumeClaimTemplate EmbeddedPersistentVolumeClaim `json:"volumeClaimTemplate,omitempty"`
}

// EmbeddedPersistentVolumeClaim is an embedded version of k8s.io/api/core/v1.PersistentVolumeClaim.
// It contains TypeMeta and a reduced ObjectMeta.
type EmbeddedPersistentVolumeClaim struct {
	metav1.TypeMeta `json:",inline"`
	// EmbeddedObjectMetadata contains metadata relevant to an embedded resource.
	EmbeddedObjectMetadata `json:"metadata,omitempty"`
	// Spec defines the desired characteristics of a volume requested by a pod author.
	// More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#persistentvolumeclaims
	Spec v1.PersistentVolumeClaimSpec `json:"spec,omitempty"`
	// Status represents the current information/status of a persistent volume claim.
	// Read only.
	// More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#persistentvolumeclaims
	Status v1.PersistentVolumeClaimStatus `json:"status,omitempty"`
}

// RemoteWriteSpec defines the remote_write configuration for Prometheus.
type RemoteWriteSpec struct {
	// Name of the remote_write queue. Must be unique if specified. The name is
	// used in metrics and logging in order to differentiate queues.
	Name string `json:"name,omitempty"`
	// URL of the endpoint to send samples to.
	URL string `json:"url"`
	// RemoteTimeout is the timeout for requests to the remote_write endpoint.
	RemoteTimeout string `json:"remoteTimeout,omitempty"`
	// Headers is a set of custom HTTP headers to be sent along with each
	// remote_write request. Be aware that any headers set by Grafana Agent
	// itself can't be overwritten.
	Headers map[string]string `json:"headers,omitempty"`
	// WriteRelabelConfigs holds relabel_configs to relabel samples before they are
	// sent to the remote_write endpoint.
	WriteRelabelConfigs []prom_v1.RelabelConfig `json:"writeRelabelConfigs,omitempty"`
	// BasicAuth for the URL.
	BasicAuth *BasicAuth `json:"basicAuth,omitempty"`
	// BearerToken used for remote_write.
	BearerToken string `json:"bearerToken,omitempty"`
	// BearerTokenFile used to read bearer token.
	BearerTokenFile string `json:"bearerTokenFile,omitempty"`
	// SigV4 configures SigV4-based authentication to the remote_write endpoint.
	// Will be used if SigV4 is defined, even with an empty object.
	SigV4 *SigV4Config `json:"sigv4,omitempty"`
	// TLSConfig to use for remote_write.
	TLSConfig *prom_v1.TLSConfig `json:"tlsConfig,omitempty"`
	// ProxyURL to proxy requests through. Optional.
	ProxyURL string `json:"proxyUrl,omitempty"`
	// QueueConfig allows tuning of the remote_write queue parameters.
	QueueConfig *QueueConfig `json:"queueConfig,omitempty"`
	// MetadataConfig configures the sending of series metadata to remote storage.
	MetadataConfig *MetadataConfig `json:"metadataConfig,omitempty"`
}

// BasicAuth allows an endpoint to authenticate over basic authentication.
// More info: https://prometheus.io/docs/operating/configuration/#endpoints.
type BasicAuth struct {
	// Username is the secret in the ServiceMonitor namespace that contains the
	// username for authentication.
	Username v1.SecretKeySelector `json:"username,omitempty"`
	// Password is the secret in the ServiceMonitor namespace that contains the
	// password for authentication.
	Password v1.SecretKeySelector `json:"password,omitempty"`
}

// SigV4Config specifies configuration to perform SigV4 authentication.
type SigV4Config struct {
	// Region of the AWS endpoint. If blank, the region from the default
	// credentials chain is used.
	Region string `json:"region,omitempty"`
	// AccessKey holds the secret of the AWS API access key to use for signing.
	// If not provided, The environment variable AWS_ACCESS_KEY_ID is used.
	AccessKey *v1.SecretKeySelector `json:"accessKey,omitempty"`
	// SecretKey of the AWS API to use for signing. If blank, the environment
	// variable AWS_SECRET_ACCESS_KEY is used.
	SecretKey *v1.SecretKeySelector `json:"secretKey,omitempty"`
	// Profile is the named AWS profile to use for authentication.
	Profile string `json:"profile,omitempty"`
	// RoleARN is the AWS Role ARN to use for authentication, as an alternative
	// for using the AWS API keys.
	RoleARN string `json:"roleARN,omitempty"`
}

// QueueConfig allows the tuning of remote_write queue_config parameters.
type QueueConfig struct {
	// Capacity is the number of samples to buffer per shard before we start dropping them.
	Capacity int `json:"capacity,omitempty"`
	// MinShards is the minimum number of shards, i.e. amount of concurrency.
	MinShards int `json:"minShards,omitempty"`
	// MaxShards is the maximum number of shards, i.e. amount of concurrency.
	MaxShards int `json:"maxShards,omitempty"`
	// MaxSamplesPerSend is the maximum number of samples per send.
	MaxSamplesPerSend int `json:"maxSamplesPerSend,omitempty"`
	// BatchSendDeadline is the maximum time a sample will wait in buffer.
	BatchSendDeadline string `json:"batchSendDeadline,omitempty"`
	// MaxRetries is the maximum number of times to retry a batch on recoverable errors.
	MaxRetries int `json:"maxRetries,omitempty"`
	// MinBackoff is the initial retry delay. Gets doubled for every retry.
	MinBackoff string `json:"minBackoff,omitempty"`
	// MaxBackoff is the maximum retry delay.
	MaxBackoff string `json:"maxBackoff,omitempty"`
	// RetryOnRateLimit retries requests when encountering rate limits.
	RetryOnRateLimit bool `json:"retryOnRateLimit,omitempty"`
}

// MetadataConfig configures the sending of series metadata to remote storage.
type MetadataConfig struct {
	// Send enables metric metadata to be sent to remote storage.
	Send bool `json:"send,omitempty"`
	// SendInterval controls how frequently metric metadata is sent to remote storage.
	SendInterval string `json:"sendInterval,omitempty"`
}

// ArbitraryFSAccessThroughSMsConfig enables users to configure whether a
// ServiceMonitor selected by the Grafana Agent instance is allowed to use
// arbitrary files on the file system of the Grafana Agent container. For
// example, this is the case when a service monitor specifies a BearerTokenFile
// in an endpoint. A malicious user could create a ServiceMonitor selecting
// arbitrary secret files in the Grafana Agent container. Those secrets would
// then be sent with a scrape request by Grafana Agent to the malicious target.
// Denying the above would prevent the attack. As an alternative, users can use
// the BearerTokenSecret field.
type ArbitraryFSAccessThroughSMsConfig struct {
	Deny bool `json:"deny,omitempty"`
}

// PrometheusSubsystemSpec defines global settings to apply across the
// Prometheus subsystem.
type PrometheusSubsystemSpec struct {
	// RemoteWrite controls default remote_write settings for all instances. If
	// an instance does not provide its own remoteWrite settings, these will be
	// used instead.
	RemoteWrite []RemoteWriteSpec `json:"remoteWrite,omitempty"`
	// Replicas of each shard to deploy for metrics pods. Number of replicas
	// multiplied by the number of shards is the total number of pods created.
	Replicas *int32 `json:"replicas,omitempty"`
	// Shards to distribute targets onto. Number of replicas multiplied by the
	// number of shards is the total number of pods created. Note that scaling
	// down shards will not reshard data onto remaining instances, it must be
	// manually moved. Increasing shards will not reshard data either but it will
	// continue to be available from the same instances. Sharding is performed on
	// the content of the __address__ target meta-label.
	Shards *int32 `json:"shards,omitempty"`
	// ReplicaExternalLabelName is the name of the Prometheus external label used
	// to denote replica name. Defaults to __replica__. External label will _not_
	// be added when value is set to the empty string.
	ReplicaExternalLabelName *string `json:"replicaExternalLabelName,omitempty"`
	// PrometheusExternalLabelName is the name of the external label used to
	// denote Grafana Agent cluster. Defaults to "cluster." External label will
	// _not_ be added when value is set to the empty string.
	PrometheusExternalLabelName *string `json:"prometheusExternalLabelName,omitempty"`
	// ScrapeInterval is the time between consecutive scrapes.
	ScrapeInterval string `json:"scrapeInterval,omitempty"`
	// ScrapeTimeout is the time to wait for a target to respond before marking a
	// scrape as failed.
	ScrapeTimeout string `json:"scrapeTimeout,omitempty"`
	// ExternalLabels are labels to add to any time series when sending data over
	// remote_write.
	ExternalLabels map[string]string `json:"externalLabels,omitempty"`
	// ArbitraryFSAccessThroughSMs configures whether configuration based on a
	// ServiceMonitor can access arbitrary files on the file system of the
	// Grafana Agent container e.g. bearer token files.
	ArbitraryFSAccessThroughSMs ArbitraryFSAccessThroughSMsConfig `json:"arbitraryFSAccessThroughSMs,omitempty"`
	// OverrideHonorLabels, if true, overrides all configured honor_labels read
	// from ServiceMonitor or PodMonitor to false.
	OverrideHonorLabels bool `json:"overrideHonorLabels,omitempty"`
	// OverrideHonorTimestamps allows to globally enforce honoring timestamps in all scrape configs.
	OverrideHonorTimestamps bool `json:"overrideHonorTimestamps,omitempty"`
	// IgnoreNamespaceSelectors, if true, will ignore NamespaceSelector settings
	// from the PodMonitor and ServiceMonitor configs, and they will only
	// discover endpoints within their current namespace.
	IgnoreNamespaceSelectors bool `json:"ignoreNamespaceSelectors,omitempty"`
	// EnforcedNamepsaceLabel enforces adding a namespace label of origin for
	// each metric that is user-created. The label value will always be the
	// namespace of the object that is being created.
	EnforcedNamepsaceLabel string `json:"enforcedNamespaceLabel,omitempty"`
	// EnforcedSampleLimit defines global limit on the number of scraped samples
	// that will be accepted. This overrides any SampleLimit set per
	// ServiceMonitor and/or PodMonitor. It is meant to be used by admins to
	// enforce the SampleLimit to keep the overall number of samples and series
	// under the desired limit. Note that if a SampleLimit from a ServiceMonitor
	// or PodMonitor is lower, that value will be used instead.
	EnforcedSampleLimit *uint64 `json:"enforcedSampleLimit,omitempty"`
	// EnforcedTargetLimit defines a global limit on the number of scraped
	// targets. This overrides any TargetLimit set per ServiceMonitor and/or
	// PodMonitor. It is meant to be used by admins to enforce the TargetLimit to
	// keep the overall number of targets under the desired limit. Note that if a
	// TargetLimit from a ServiceMonitor or PodMonitor is higher, that value will
	// be used instead.
	EnforcedTargetLimit *uint64 `json:"enforcedTargetLimit,omitempty"`

	// InstanceSelector determines which PrometheusInstances should be selected
	// for running. Each instance runs its own set of Prometheus components,
	// including service discovery, scraping, and remote_write.
	InstanceSelector *metav1.LabelSelector `json:"instanceSelector,omitempty"`
	// InstanceNamespaceSelector are the set of labels to determine which
	// namespaces to watch for PrometheusInstances. If not provided, only checks own namespace.
	InstanceNamespaceSelector *metav1.LabelSelector `json:"instanceNamespaceSelector,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path="prometheus-instances"
// +kubebuilder:resource:singular="prometheus-instance"
// +kubebuilder:resource:categories="agent-operator"

// PrometheusInstance controls an individual Prometheus instance within a
// Grafana Agent deployment.
type PrometheusInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec holds the specification of the desired behavior for the Prometheus
	// instance.
	Spec PrometheusInstanceSpec `json:"spec,omitempty"`
}

// PrometheusInstanceSpec controls how an individual instance will be used to discover PodMonitors.
type PrometheusInstanceSpec struct {
	// WALTruncateFrequency specifies how frequently the WAL truncation process
	// should run. Higher values causes the WAL to increase and for old series to
	// stay in the WAL for longer, but reduces the chances of data loss when
	// remote_write is failing for longer than the given frequency.
	WALTruncateFrequency string `json:"walTruncateFrequency,omitempty"`
	// MinWALTime is the minimum amount of time series and samples may exist in
	// the WAL before being considered for deletion.
	MinWALTime string `json:"minWALTime,omitempty"`
	// MaxWALTime is the maximum amount of time series and asmples may exist in
	// the WAL before being forcibly deleted.
	MaxWALTime string `json:"maxWALTime,omitempty"`
	// RemoteFlushDeadline is the deadline for flushing data when an instance
	// shuts down.
	RemoteFlushDeadline string `json:"remoteFlushDeadline,omitempty"`
	// WriteStaleOnShutdown writes staleness markers on shutdown for all series.
	WriteStaleOnShutdown *bool `json:"writeStaleOnShutdown,omitempty"`
	// ServiceMonitorSelector determines which ServiceMonitors should be selected
	// for target discovery.
	ServiceMonitorSelector *metav1.LabelSelector `json:"serviceMonitorSelector,omitempty"`
	// ServiceMonitorNamespaceSelector are the set of labels to determine which
	// namespaces to watch for ServiceMonitor discovery. If nil, only checks own
	// namespace.
	ServiceMonitorNamespaceSelector *metav1.LabelSelector `json:"serviceMonitorNamespaceSelector,omitempty"`
	// PodMonitorSelector determines which PodMonitors should be selected for target
	// discovery. Experimental.
	PodMonitorSelector *metav1.LabelSelector `json:"podMonitorSelector,omitempty"`
	// PodMonitorNamespaceSelector are the set of labels to determine which
	// namespaces to watch for PodMonitor discovery. If nil, only checks own
	// namespace.
	PodMonitorNamespaceSelector *metav1.LabelSelector `json:"podMonitorNamespaceSelector,omitempty"`
	// ProbeSelector determines which Probes should be selected for target
	// discovery.
	ProbeSelector *metav1.LabelSelector `json:"probeSelector,omitempty"`
	// ProbeNamespaceSelector are the set of labels to determine which namespaces
	// to watch for Probe discovery. If nil, only checks own namespace.
	ProbeNamespaceSelector *metav1.LabelSelector `json:"probeNamespaceSelector,omitempty"`
	// RemoteWrite controls remote_write settings for this instance.
	RemoteWrite []RemoteWriteSpec `json:"remoteWrite,omitempty"`
	// AdditionalScrapeConfigs allows specifying a key of a Secret containing
	// additional Grafana Agent Prometheus scrape configurations. SCrape
	// configurations specified are appended to the configurations generated by
	// the Grafana Agent Operator. Job configurations specified must have the
	// form as specified in the official Prometheus documentation:
	// https://prometheus.io/docs/prometheus/latest/configuration/configuration/#scrape_config.
	// As scrape configs are appended, the user is responsible to make sure it is
	// valid. Note that using this feature may expose the possibility to break
	// upgrades of Grafana Agent. It is advised to review both Grafana Agent and
	// Prometheus release notes to ensure that no incompatible scrape configs are
	// going to break Grafana Agent after the upgrade.
	AdditionalScrapeConfigs *v1.SecretKeySelector `json:"additionalScrapeConfigs,omitempty"`
}

// +kubebuilder:object:root=true

// PrometheusInstanceList is a list of PrometheusInsatnce.
type PrometheusInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	// Items is the list of PrometheusInstance.
	Items []*PrometheusInstance `json:"items"`
}
