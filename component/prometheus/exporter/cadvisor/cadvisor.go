package cadvisor

import (
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/prometheus/exporter"
	"github.com/grafana/agent/pkg/integrations"
	integration "github.com/grafana/agent/pkg/integrations/cadvisor"
)

func init() {
	component.Register(component.Registration{
		Name:    "prometheus.exporter.cadvisor",
		Args:    Arguments{},
		Exports: exporter.Exports{},
		Build:   exporter.New(createExporter, "cadvisor"),
	})
}

func createExporter(opts component.Options, args component.Arguments) (integrations.Integration, error) {
	a := args.(Arguments)
	return a.Convert().NewIntegration(opts.Logger)
}

// DefaultArguments holds non-zero default options for Arguments when it is
// unmarshaled from river.
var DefaultArguments = Arguments{
	// Common cadvisor config defaults
	StoreContainerLabels: true,
	ResctrlInterval:      0,

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

type Arguments struct {
	// Common cadvisor config options
	// StoreContainerLabels converts container labels and environment variables into labels on prometheus metrics for each container. If false, then only metrics exported are container name, first alias, and image name.
	StoreContainerLabels bool `river:"store_container_labels,attr,optional"`

	// AllowlistedContainerLabels list of container labels to be converted to labels on prometheus metrics for each container. store_container_labels must be set to false for this to take effect.
	AllowlistedContainerLabels []string `river:"allowlisted_container_labels,attr,optional"`

	// EnvMetadataAllowlist list of environment variable keys matched with specified prefix that needs to be collected for containers, only support containerd and docker runtime for now.
	EnvMetadataAllowlist []string `river:"env_metadata_allowlist,attr,optional"`

	// RawCgroupPrefixAllowlist list of cgroup path prefix that needs to be collected even when -docker_only is specified.
	RawCgroupPrefixAllowlist []string `river:"raw_cgroup_prefix_allowlist,attr,optional"`

	// PerfEventsConfig path to a JSON file containing configuration of perf events to measure. Empty value disabled perf events measuring.
	PerfEventsConfig string `river:"perf_events_config,attr,optional"`

	// ResctrlInterval resctrl mon groups updating interval. Zero value disables updating mon groups.
	ResctrlInterval int `river:"resctrl_interval,attr,optional"`

	// DisableMetrics list of `metrics` to be disabled.
	DisabledMetrics []string `river:"disabled_metrics,attr,optional"`

	// EnableMetrics list of `metrics` to be enabled. If set, overrides 'disable_metrics'.
	EnabledMetrics []string `river:"enabled_metrics,attr,optional"`

	// StorageDuration length of time to keep data stored in memory (Default: 2m)
	StorageDuration time.Duration `river:"storage_duration,attr,optional"`

	// Containerd config options
	// Containerd containerd endpoint
	Containerd string `river:"containerd,attr,optional"`

	// ContainerdNamespace containerd namespace
	ContainerdNamespace string `river:"containerd_namespace,attr,optional"`

	// Docker config options
	// Docker docker endpoint
	Docker string `river:"docker,attr,optional"`

	// DockerTLS use TLS to connect to docker
	DockerTLS bool `river:"docker_tls,attr,optional"`

	// DockerTLSCert path to client certificate
	DockerTLSCert string `river:"docker_tls_cert,attr,optional"`

	// DockerTLSKey path to private key
	DockerTLSKey string `river:"docker_tls_key,attr,optional"`

	// DockerTLSCA path to trusted CA
	DockerTLSCA string `river:"docker_tls_ca,attr,optional"`

	// Raw config options
	// DockerOnly only report docker containers in addition to root stats
	DockerOnly bool `river:"docker_only,attr,optional"`
}

// UnmarshalRiver implements River unmarshalling for Arguments.
func (a *Arguments) UnmarshalRiver(f func(interface{}) error) error {
	*a = DefaultArguments

	type args Arguments
	if err := f((*args)(a)); err != nil {
		return err
	}

	// In the cadvisor cmd, these are passed as CSVs, and turned into slices using strings.split. As a result the
	// default values are always a slice with 1 or more elements.
	// See: https://github.com/google/cadvisor/blob/v0.43.0/cmd/cadvisor.go#L136
	if len(a.AllowlistedContainerLabels) == 0 {
		a.AllowlistedContainerLabels = []string{""}
	}
	if len(a.RawCgroupPrefixAllowlist) == 0 {
		a.RawCgroupPrefixAllowlist = []string{""}
	}
	if len(a.EnvMetadataAllowlist) == 0 {
		a.EnvMetadataAllowlist = []string{""}
	}
}

func (a *Arguments) Convert() *integration.Config {
	return &integration.Config{
		StoreContainerLabels:       a.StoreContainerLabels,
		AllowlistedContainerLabels: a.AllowlistedContainerLabels,
		EnvMetadataAllowlist:       a.EnvMetadataAllowlist,
		RawCgroupPrefixAllowlist:   a.RawCgroupPrefixAllowlist,
		PerfEventsConfig:           a.PerfEventsConfig,
		ResctrlInterval:            a.ResctrlInterval,
		DisabledMetrics:            a.DisabledMetrics,
		EnabledMetrics:             a.EnabledMetrics,
		StorageDuration:            a.StorageDuration,
		Containerd:                 a.Containerd,
		ContainerdNamespace:        a.ContainerdNamespace,
		Docker:                     a.Docker,
		DockerTLS:                  a.DockerTLS,
		DockerTLSCert:              a.DockerTLSCert,
		DockerTLSKey:               a.DockerTLSKey,
		DockerTLSCA:                a.DockerTLSCA,
		DockerOnly:                 a.DockerOnly,
	}
}
