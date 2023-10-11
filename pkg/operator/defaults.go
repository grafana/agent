package operator

// Supported versions of the Grafana Agent.
var (
	DefaultAgentVersion   = "v0.37.1"
	DefaultAgentBaseImage = "grafana/agent"
	DefaultAgentImage     = DefaultAgentBaseImage + ":" + DefaultAgentVersion
)

// Defaults for Prometheus Config Reloader.
var (
	DefaultConfigReloaderVersion   = "v0.67.1"
	DefaultConfigReloaderBaseImage = "quay.io/prometheus-operator/prometheus-config-reloader"
	DefaultConfigReloaderImage     = DefaultConfigReloaderBaseImage + ":" + DefaultConfigReloaderVersion
)
