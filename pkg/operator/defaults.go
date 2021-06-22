package operator

// Supported versions of the Grafana Agent.
var (
	AgentCompatibilityMatrix = []string{
		"v0.14.0",
		"v0.15.0",
		// "v0.16.0", // Pulled due to critical bug fixed in v0.16.1.
		"v0.16.1",

		// NOTE(rfratto): when performing an upgrade, add the newest version above instead of changing the existing reference.
	}

	DefaultAgentVersion   = AgentCompatibilityMatrix[len(AgentCompatibilityMatrix)-1]
	DefaultAgentBaseImage = "grafana/agent"
	DefaultAgentImage     = DefaultAgentBaseImage + ":" + DefaultAgentVersion
)
