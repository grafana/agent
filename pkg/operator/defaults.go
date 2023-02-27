package operator

// Supported versions of the Grafana Agent.
var (
	AgentCompatibilityMatrix = []string{
		"v0.14.0",
		"v0.15.0",
		// "v0.16.0", // Pulled due to critical bug fixed in v0.16.1.
		"v0.16.1",
		"v0.17.0",
		"v0.18.0",
		"v0.18.1",
		"v0.18.2",
		"v0.18.3",
		"v0.18.4",
		"v0.19.0",
		"v0.20.0",
		"v0.20.1",
		"v0.21.0",
		"v0.21.1",
		"v0.21.2",
		"v0.22.0",
		"v0.23.0",
		"v0.24.0",
		"v0.24.1",
		"v0.24.2",
		"v0.25.0",
		"v0.25.1",
		"v0.26.0",
		"v0.26.1",
		"v0.27.0",
		"v0.27.1",
		"v0.28.0",
		"v0.28.1",
		"v0.29.0",
		"v0.30.0",
		"v0.30.1",
		"v0.30.2",
		"v0.31.0",
		"v0.31.1",
		// NOTE(rfratto): when performing an upgrade, add the newest version above instead of changing the existing reference.
	}

	DefaultAgentVersion   = AgentCompatibilityMatrix[len(AgentCompatibilityMatrix)-1]
	DefaultAgentBaseImage = "grafana/agent"
	DefaultAgentImage     = DefaultAgentBaseImage + ":" + DefaultAgentVersion
)
