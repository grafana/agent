package agentstate

import "bytes"

type Client interface {
	// SetAgentState sets the current agent state for the client. This must be
	// called each time the agent state changes.
	SetAgentState(agentState AgentState)

	// SetComponents sets the current components state for the client. This must
	// be called each time the component state changes.
	SetComponents(components []Component)

	// Send encodes and sends the agent state to the configured destination.
	Send() error

	// Write writes the agent state to the buffer.
	Write() (bytes.Buffer, error)

	// WriteToFile writes the agent state to the specified file.
	WriteToFile(filepath string) error
}
