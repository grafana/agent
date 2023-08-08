package agentstate

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/parquet-go/parquet-go"
)

// ParquetClient manages and writes agent state to a parquet file.
type ParquetClient struct {
	agentState AgentState
	components []Component
	endpoint   string
	tenant     string

	httpClient http.Client
}

var _ Client = (*ParquetClient)(nil)

// NewParquetClient creates a new client for managing and writing agent state.
func NewParquetClient(agentState AgentState, components []Component, endpoint string, tenant string) *ParquetClient {
	return &ParquetClient{
		agentState: agentState,
		components: components,
		endpoint:   endpoint,
		tenant:     tenant,
		httpClient: http.Client{Timeout: 5 * time.Second},
	}
}

// Implements agentstate.Client
func (pc *ParquetClient) SetAgentState(agentState AgentState) {
	pc.agentState = agentState
}

// Implements agentstate.Client
func (pc *ParquetClient) SetComponents(components []Component) {
	pc.components = components
}

// Implements agentstate.Client
func (pc *ParquetClient) Send(ctx context.Context) error {
	buf, err := pc.Write()
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, pc.endpoint, &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/parquet")
	req.Header.Set("X-TenantID", pc.tenant)

	resp, err := pc.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("failed to send agent state: %s  body: %s", resp.Status, string(data))
	}

	return nil
}

// Implements agentstate.Client
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

// Implements agentstate.Client
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
