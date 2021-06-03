package operator

// Supported versions of the Grafana Agent.
var (
	AgentCompatibilityMatrix = []string{
		"v0.14.0",
	}

	DefaultAgentVersion   = AgentCompatibilityMatrix[len(AgentCompatibilityMatrix)-1]
	DefaultAgentBaseImage = "grafana/agent"
	DefaultAgentImage     = DefaultAgentBaseImage + ":" + DefaultAgentVersion
)
