package operator

// Supported versions of the Grafana Agent.
var (
	DefaultAgentVersion   = "v0.34.1"
	DefaultAgentBaseImage = "grafana/agent"
	DefaultAgentImage     = DefaultAgentBaseImage + ":" + DefaultAgentVersion
)
