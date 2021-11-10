package cadvisor //nolint:golint

import (
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/google/cadvisor/container"

	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/integrations/config"
)

// DefaultConfig holds the default settings for the cadvisor integration
var DefaultConfig Config = Config{
	// Common cadvisor config defaults
	StoreContainerLabels: true,
	ResctrlInterval:      0,
	// Matching the default disabled set from cadvisor - https://github.com/google/cadvisor/blob/3c6e3093c5ca65c57368845ddaea2b4ca6bc0da8/cmd/cadvisor.go#L78-L93
	disabledMetricsSet: container.MetricSet{
		container.MemoryNumaMetrics:              struct{}{},
		container.NetworkTcpUsageMetrics:         struct{}{},
		container.NetworkUdpUsageMetrics:         struct{}{},
		container.NetworkAdvancedTcpUsageMetrics: struct{}{},
		container.ProcessSchedulerMetrics:        struct{}{},
		container.ProcessMetrics:                 struct{}{},
		container.HugetlbUsageMetrics:            struct{}{},
		container.ReferencedMemoryMetrics:        struct{}{},
		container.CPUTopologyMetrics:             struct{}{},
		container.ResctrlMetrics:                 struct{}{},
		container.CPUSetMetrics:                  struct{}{},
	},
	enabledMetricsSet: container.MetricSet{},

	// Containerd config defaults
	Containerd:          "/run/containerd/containerd.sock",
	ContainerdNamespace: "k8s.io",
}

// Config controls cadvisor
type Config struct {
	Common config.Common `yaml:",inline"`

	// Common cadvisor config options
	// StoreContainerLabels converts container labels and environment variables into labels on prometheus metrics for each container. If false, then only metrics exported are container name, first alias, and image name.
	StoreContainerLabels bool `yaml:"store_container_labels,omitempty"`

	// WhitelistedContainerLabels list of container labels to be converted to labels on prometheus metrics for each container. store_container_labels must be set to false for this to take effect.
	WhitelistedContainerLabels []string `yaml:"whitelisted_container_labels,omitempty"`

	// EnvMetadataWhitelist list of environment variable keys matched with specified prefix that needs to be collected for containers, only support containerd and docker runtime for now.
	EnvMetadataWhitelist []string `yaml:"env_metadata_whitelist,omitempty"`

	// RawCgroupPrefixWhitelist list of cgroup path prefix that needs to be collected even when -docker_only is specified.
	RawCgroupPrefixWhitelist []string `yaml:"raw_cgroup_prefix_whitelist,omitempty"`

	// PerfEventsConfig path to a JSON file containing configuration of perf events to measure. Empty value disabled perf events measuring.
	PerfEventsConfig string `yaml:"perf_events_config,omitempty"`

	// ResctrlInterval resctrl mon groups updating interval. Zero value disables updating mon groups.
	ResctrlInterval int `yaml:"resctrl_interval,omitempty"`

	// DisableMetrics list of `metrics` to be disabled.
	DisabledMetrics []string `yaml:"disabled_metrics,omitempty"`

	// disabledMetricsSet list of `metrics` to be disabled, in the form required by the cadvisor collector(s)
	disabledMetricsSet container.MetricSet

	// EnableMetrics list of `metrics` to be enabled. If set, overrides 'disable_metrics'.
	EnabledMetrics []string `yaml:"enabled_metrics,omitempty"`

	// enabledMetricsSet list of `metrics` to be enabled, in the form required by the cadvisor collector(s)
	enabledMetricsSet container.MetricSet

	// Containerd config options
	// Containerd containerd endpoint
	Containerd string `yaml:"containerd,omitempty"`

	// ContainerdNamespace containerd namespace
	ContainerdNamespace string `yaml:"containerd_namespace,omitempty"`
}

// UnmarshalYAML implements yaml.Unmarshaler for Config
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultConfig

	type plain Config
	err := unmarshal((*plain)(c))
	// Clear default disabled metrics if explicit disabled metrics are configured
	if len(c.DisabledMetrics) > 0 {
		c.disabledMetricsSet = container.MetricSet{}
	}
	for _, d := range c.DisabledMetrics {
		if err := c.disabledMetricsSet.Set(d); err != nil {
			return fmt.Errorf("failed to set disabled metric: %w", err)
		}
	}

	for _, e := range c.EnabledMetrics {
		if err := c.enabledMetricsSet.Set(e); err != nil {
			return fmt.Errorf("failed to set enabled metric: %w", err)
		}
	}
	return err
}

// Name returns the name of the integration that this config represents.
func (c *Config) Name() string {
	return "cadvisor"
}

// CommonConfig returns the common settings shared across all configs for
// integrations.
func (c *Config) CommonConfig() config.Common {
	return c.Common
}

// InstanceKey returns the hostname:port of the GitHub API server.
func (c *Config) InstanceKey(agentKey string) (string, error) {
	return agentKey, nil
}

// NewIntegration creates a new github_exporter
func (c *Config) NewIntegration(logger log.Logger) (integrations.Integration, error) {
	return New(logger, c)
}

func init() {
	integrations.RegisterIntegration(&Config{})
}

// New creates a new cadvisor integration
func New(logger log.Logger, c *Config) (integrations.Integration, error) {
	return nil, nil
}
