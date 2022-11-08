package v1alpha1

import (
	prom_v1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MetricsSubsystemSpec defines global settings to apply across the
// Metrics subsystem.
type MetricsSubsystemSpec struct {
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
	// ReplicaExternalLabelName is the name of the metrics external label used
	// to denote replica name. Defaults to __replica__. External label will _not_
	// be added when value is set to the empty string.
	ReplicaExternalLabelName *string `json:"replicaExternalLabelName,omitempty"`
	// MetricsExternalLabelName is the name of the external label used to
	// denote Grafana Agent cluster. Defaults to "cluster." External label will
	// _not_ be added when value is set to the empty string.
	MetricsExternalLabelName *string `json:"metricsExternalLabelName,omitempty"`
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
	ArbitraryFSAccessThroughSMs prom_v1.ArbitraryFSAccessThroughSMsConfig `json:"arbitraryFSAccessThroughSMs,omitempty"`
	// OverrideHonorLabels, if true, overrides all configured honor_labels read
	// from ServiceMonitor or PodMonitor to false.
	OverrideHonorLabels bool `json:"overrideHonorLabels,omitempty"`
	// OverrideHonorTimestamps allows to globally enforce honoring timestamps in all scrape configs.
	OverrideHonorTimestamps bool `json:"overrideHonorTimestamps,omitempty"`
	// IgnoreNamespaceSelectors, if true, will ignore NamespaceSelector settings
	// from the PodMonitor and ServiceMonitor configs, and they will only
	// discover endpoints within their current namespace.
	IgnoreNamespaceSelectors bool `json:"ignoreNamespaceSelectors,omitempty"`
	// EnforcedNamespaceLabel enforces adding a namespace label of origin for
	// each metric that is user-created. The label value will always be the
	// namespace of the object that is being created.
	EnforcedNamespaceLabel string `json:"enforcedNamespaceLabel,omitempty"`
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

	// InstanceSelector determines which MetricsInstances should be selected
	// for running. Each instance runs its own set of Metrics components,
	// including service discovery, scraping, and remote_write.
	InstanceSelector *metav1.LabelSelector `json:"instanceSelector,omitempty"`
	// InstanceNamespaceSelector are the set of labels to determine which
	// namespaces to watch for MetricsInstances. If not provided, only checks own namespace.
	InstanceNamespaceSelector *metav1.LabelSelector `json:"instanceNamespaceSelector,omitempty"`
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
	BasicAuth *prom_v1.BasicAuth `json:"basicAuth,omitempty"`
	// Oauth2 for URL
	OAuth2 *prom_v1.OAuth2 `json:"oauth2,omitempty"`
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

// +kubebuilder:object:root=true
// +kubebuilder:resource:path="metricsinstances"
// +kubebuilder:resource:singular="metricsinstance"
// +kubebuilder:resource:categories="agent-operator"

// MetricsInstance controls an individual Metrics instance within a
// Grafana Agent deployment.
type MetricsInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec holds the specification of the desired behavior for the Metrics
	// instance.
	Spec MetricsInstanceSpec `json:"spec,omitempty"`
}

// ServiceMonitorSelector returns a selector to find ServiceMonitors.
func (p *MetricsInstance) ServiceMonitorSelector() ObjectSelector {
	return ObjectSelector{
		ObjectType:        &prom_v1.ServiceMonitor{},
		ParentNamespace:   p.Namespace,
		NamespaceSelector: p.Spec.ServiceMonitorNamespaceSelector,
		Labels:            p.Spec.ServiceMonitorSelector,
	}
}

// PodMonitorSelector returns a selector to find PodMonitors.
func (p *MetricsInstance) PodMonitorSelector() ObjectSelector {
	return ObjectSelector{
		ObjectType:        &prom_v1.PodMonitor{},
		ParentNamespace:   p.Namespace,
		NamespaceSelector: p.Spec.PodMonitorNamespaceSelector,
		Labels:            p.Spec.PodMonitorSelector,
	}
}

// ProbeSelector returns a selector to find Probes.
func (p *MetricsInstance) ProbeSelector() ObjectSelector {
	return ObjectSelector{
		ObjectType:        &prom_v1.Probe{},
		ParentNamespace:   p.Namespace,
		NamespaceSelector: p.Spec.ProbeNamespaceSelector,
		Labels:            p.Spec.ProbeSelector,
	}
}

// MetricsInstanceSpec controls how an individual instance will be used to discover PodMonitors.
type MetricsInstanceSpec struct {
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

// MetricsInstanceList is a list of MetricsInstance.
type MetricsInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	// Items is the list of MetricsInstance.
	Items []*MetricsInstance `json:"items"`
}
