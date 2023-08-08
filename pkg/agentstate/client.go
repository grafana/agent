package agentstate

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"

	"github.com/parquet-go/parquet-go"
)

// Client manages and writes agent state to a parquet file.
type Client struct {
	agentState AgentState
	components []Component

	writer *parquet.GenericWriter[Component]
	buf    bytes.Buffer

	agentSeedController *AgentSeedController
}

// NewClient creates a new client for managing and writing agent state.
func NewClient(agentState AgentState, components []Component) *Client {
	var buf bytes.Buffer
	agentSeedController := NewAgentSeedController(agentSeedFileName())

	return &Client{
		agentState:          agentState,
		components:          components,
		writer:              parquet.NewGenericWriter[Component](&buf),
		buf:                 buf,
		agentSeedController: agentSeedController,
	}
}

func agentSeedFileName() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("APPDATA"), "grafana_agent_seed.json")
	}
	// linux/mac
	return "/tmp/grafana_agent_seed.json"
}

// SetAgentState sets the current agent state for the client.
func (c *Client) SetAgentState(agentState AgentState) {
	c.agentState = agentState
}

// SetComponents sets the current components state for the client.
func (c *Client) SetComponents(components []Component) {
	c.components = components
}

// Buf returns the buffer containing the agent state.
func (c *Client) Buf() bytes.Buffer {
	return c.buf
}

// Write writes the agent state to the buffer.
func (c *Client) Write() error {
	c.buf.Reset()
	c.writer.Reset(&c.buf)
	if err := c.writeRowGroups(); err != nil {
		return err
	}

	c.writeMetadata()
	return c.writer.Close()
}

// WriteToFile writes the agent state to a file at the given filepath. This
// will overwrite the file if it already exists.
func (c *Client) WriteToFile(filepath string) error {
	if err := c.Write(); err != nil {
		return err
	}

	f, err := os.Create(filepath)
	if err != nil {
		return err
	}
	_, err = f.Write(c.buf.Bytes())
	return err
}

// writeRowGroups writes the agent state components to the parquet file.
func (c *Client) writeRowGroups() error {
	rowGroup := parquet.NewGenericBuffer[Component]()
	_, err := rowGroup.Write(c.components)
	if err != nil {
		return err
	}

	_, err = c.writer.WriteRowGroup(rowGroup)
	return err
}

// writeMetadata writes the agent state metadata to the parquet file.
func (c *Client) writeMetadata() {
	// SetKeyValueMetadata will overwrite metadata on matching keys rather than panic.
	c.writer.SetKeyValueMetadata("ID", c.agentState.ID)
	for key, label := range c.agentState.Labels {
		c.writer.SetKeyValueMetadata(key, label)
	}
}
