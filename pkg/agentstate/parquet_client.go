package agentstate

import (
	"bytes"
	"os"

	"github.com/parquet-go/parquet-go"
)

// ParquetClient manages and writes agent state to a parquet file.
type ParquetClient struct {
	agentState AgentState
	components []Component
}

var _ Client = (*ParquetClient)(nil)

// NewParquetClient creates a new client for managing and writing agent state.
func NewParquetClient(agentState AgentState, components []Component) *ParquetClient {
	return &ParquetClient{
		agentState: agentState,
		components: components,
	}
}

// SetAgentState sets the current agent state for the client.
func (pc *ParquetClient) SetAgentState(agentState AgentState) {
	pc.agentState = agentState
}

// SetComponents sets the current components state for the client.
func (pc *ParquetClient) SetComponents(components []Component) {
	pc.components = components
}

func (pc *ParquetClient) Send() error {
	// TODO

	return nil
}

// Write writes the agent state to the buffer.
func (c *ParquetClient) Write() (bytes.Buffer, error) {
	var buf bytes.Buffer
	writer := parquet.NewGenericWriter[Component](&buf)

	// Write the component data to the buffer.
	rowGroup := parquet.NewGenericBuffer[Component]()
	_, err := rowGroup.Write(c.components)
	if err != nil {
		return buf, err
	}

	_, err = writer.WriteRowGroup(rowGroup)
	if err != nil {
		return buf, err
	}

	// Write the metadata to the buffer.
	writer.SetKeyValueMetadata("ID", c.agentState.ID)
	for key, label := range c.agentState.Labels {
		writer.SetKeyValueMetadata(key, label)
	}

	err = writer.Close()
	return buf, err
}

// WriteToFile writes the agent state to a file at the given filepath. This
// will overwrite the file if it already exists.
func (c *ParquetClient) WriteToFile(filepath string) error {
	buf, err := c.Write()
	if err != nil {
		return err
	}

	f, err := os.Create(filepath)
	if err != nil {
		return err
	}
	_, err = f.Write(buf.Bytes())
	return err
}
