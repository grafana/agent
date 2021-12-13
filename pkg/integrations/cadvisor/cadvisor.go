//go:build linux
// +build linux

package cadvisor //nolint:golint

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/google/cadvisor/cache/memory"
	"github.com/google/cadvisor/container"
	v2 "github.com/google/cadvisor/info/v2"
	"github.com/google/cadvisor/manager"
	"github.com/google/cadvisor/metrics"
	"github.com/google/cadvisor/storage"
	"github.com/google/cadvisor/utils/sysfs"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"k8s.io/klog/v2"
	"k8s.io/utils/clock"

	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/integrations/config"

	// Register container providers
	"github.com/google/cadvisor/container/containerd"
	_ "github.com/google/cadvisor/container/containerd/install" // register containerd container plugin
	_ "github.com/google/cadvisor/container/crio/install"       // register crio container plugin
	"github.com/google/cadvisor/container/docker"
	_ "github.com/google/cadvisor/container/docker/install" // register docker container plugin
	"github.com/google/cadvisor/container/raw"
	_ "github.com/google/cadvisor/container/systemd/install" // register systemd container plugin
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

	StorageDuration: 2 * time.Minute,

	// Containerd config defaults
	Containerd:          "/run/containerd/containerd.sock",
	ContainerdNamespace: "k8s.io",

	// Docker config defaults
	Docker:        "unix:///var/run/docker.sock",
	DockerTLS:     false,
	DockerTLSCert: "cert.pem",
	DockerTLSKey:  "key.pem",
	DockerTLSCA:   "ca.pem",

	// Raw config defaults
	DockerOnly: false,
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

	// includedMetrics is the final calculated set of metrics which will be scraped for containers
	includedMetrics container.MetricSet

	// StorageDuration length of time to keep data stored in memory (Default: 2m)
	StorageDuration time.Duration `yaml:"storage_duration,omitempty"`

	// Containerd config options
	// Containerd containerd endpoint
	Containerd string `yaml:"containerd,omitempty"`

	// ContainerdNamespace containerd namespace
	ContainerdNamespace string `yaml:"containerd_namespace,omitempty"`

	// Docker config options
	// Docker docker endpoint
	Docker string `yaml:"docker,omitempty"`

	// DockerTLS use TLS to connect to docker
	DockerTLS bool `yaml:"docker_tls,omitempty"`

	// DockerTLSCert path to client certificate
	DockerTLSCert string `yaml:"docker_tls_cert,omitempty"`

	// DockerTLSKey path to private key
	DockerTLSKey string `yaml:"docker_tls_key,omitempty"`

	// DockerTLSCA path to trusted CA
	DockerTLSCA string `yaml:"docker_tls_ca,omitempty"`

	// Raw config options
	// DockerOnly only report docker containers in addition to root stats
	DockerOnly bool `yaml:"docker_only,omitempty"`
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

	if len(c.enabledMetricsSet) > 0 {
		c.includedMetrics = c.enabledMetricsSet
	} else {
		c.includedMetrics = container.AllMetrics.Difference(c.disabledMetricsSet)
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

// InstanceKey returns the hostname:port of the cadvisor API server.
func (c *Config) InstanceKey(agentKey string) (string, error) {
	return agentKey, nil
}

// NewIntegration creates a new cadvisor integration
func (c *Config) NewIntegration(logger log.Logger) (integrations.Integration, error) {
	return New(logger, c)
}

func init() {
	integrations.RegisterIntegration(&Config{})
}

// New creates a new cadvisor integration
func New(logger log.Logger, c *Config) (integrations.Integration, error) {
	klog.SetLogger(logger)

	// Do gross global configs. This works, so long as there is only one instance of the cAdvisor integration
	// per host.
	// Containerd
	containerd.ArgContainerdEndpoint = &c.Containerd
	containerd.ArgContainerdNamespace = &c.ContainerdNamespace

	// Docker
	docker.ArgDockerEndpoint = &c.Docker
	docker.ArgDockerTLS = &c.DockerTLS
	docker.ArgDockerCert = &c.DockerTLSCert
	docker.ArgDockerKey = &c.DockerTLSKey
	docker.ArgDockerCA = &c.DockerTLSCA

	// Raw
	raw.DockerOnly = &c.DockerOnly

	// Only using in-memory storage, with no backup storage for cadvisor stats
	memoryStorage := memory.New(c.StorageDuration, []storage.StorageDriver{})

	sysFs := sysfs.NewRealSysFs()

	collectorHTTPClient := http.Client{}

	rm, err := manager.New(memoryStorage, sysFs, manager.HousekeepingConfigFlags, c.includedMetrics, &collectorHTTPClient, c.RawCgroupPrefixWhitelist, c.EnvMetadataWhitelist, c.PerfEventsConfig, time.Duration(c.ResctrlInterval))
	if err != nil {
		return nil, fmt.Errorf("failed to create a manager: %w", err)
	}

	if err := rm.Start(); err != nil {
		return nil, fmt.Errorf("failed to start manager: %w", err)
	}

	containerLabelFunc := metrics.DefaultContainerLabels
	if !c.StoreContainerLabels {
		containerLabelFunc = metrics.BaseContainerLabels(c.WhitelistedContainerLabels)
	}

	goCol := collectors.NewGoCollector()                                         // This is already emitted by the agent, but not with the integration job name.
	procCol := collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}) // Same as above
	machCol := metrics.NewPrometheusMachineCollector(rm, c.includedMetrics)
	// This is really just a concatenation of the defaults found at;
	// https://github.com/google/cadvisor/tree/f89291a53b80b2c3659fff8954c11f1fc3de8a3b/cmd/internal/api/versions.go#L536-L540
	// https://github.com/google/cadvisor/tree/f89291a53b80b2c3659fff8954c11f1fc3de8a3b/cmd/internal/http/handlers.go#L109-L110
	// AFAIK all we are ever doing is the "default" metrics request, and we don't need to support the "docker" request type.
	reqOpts := v2.RequestOptions{
		IdType:    v2.TypeName,
		Count:     1,
		Recursive: true,
	}
	contCol := metrics.NewPrometheusCollector(rm, containerLabelFunc, c.includedMetrics, clock.RealClock{}, reqOpts)

	integration := integrations.NewCollectorIntegration(c.Name(), integrations.WithCollectors(goCol, procCol, machCol, contCol))
	return integration, nil
}
