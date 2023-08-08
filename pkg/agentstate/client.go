package agentstate

import (
	"bytes"
	"context"
)

type Client interface {
	// SetAgentState sets the current agent state for the client. This must be
	// called each time the agent state changes.
	SetAgentState(agentState AgentState)

	// SetComponents sets the current components state for the client. This must
	// be called each time the component state changes.
	SetComponents(components []Component)

	// Send encodes and sends the agent state to the configured destination.
	Send(ctx context.Context) error

	// Write writes the agent state to the buffer.
	Write() (bytes.Buffer, error)

	// WriteToFile writes the agent state to a file at the given filepath. This
	// will overwrite the file if it already exists.
	WriteToFile(filepath string) error
}
