package operator

// Supported versions of the Grafana Agent.
var (
	DefaultAgentVersion   = "v0.33.0-rc.2"
	DefaultAgentBaseImage = "grafana/agent"
	DefaultAgentImage     = DefaultAgentBaseImage + ":" + DefaultAgentVersion
)
