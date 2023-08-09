package observer

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/grafana/agent/pkg/river/encoding/riveragentstate"
	"github.com/parquet-go/parquet-go"
	prom_config "github.com/prometheus/common/config"
)

// ParquetClient manages and writes agent state to a parquet file.
type ParquetClient struct {
	agentState riveragentstate.AgentState
	components []riveragentstate.Component

	httpClient *http.Client
}

var _ Client = (*ParquetClient)(nil)

// NewParquetClient creates a new client for managing and writing agent state.
func NewParquetClient(agentState riveragentstate.AgentState, components []riveragentstate.Component, args *Arguments) *ParquetClient {
	if args != nil {
		cli, err := prom_config.NewClientFromConfig(
			*args.HTTPClientConfig.Convert(),
			"agent_observer",
		)
		if err != nil {
			panic(fmt.Sprintf("failed to create http client: %s", err))
		}

		return &ParquetClient{
			agentState: agentState,
			components: components,
			httpClient: cli,
		}
	}

	return &ParquetClient{
		agentState: agentState,
		components: components,
		httpClient: nil,
	}
}

// Implements agentstate.Client
func (pc *ParquetClient) SetAgentState(agentState riveragentstate.AgentState) {
	pc.agentState = agentState
}

// Implements agentstate.Client
func (pc *ParquetClient) SetComponents(components []riveragentstate.Component) {
	pc.components = components
}

// Implements agentstate.Client
func (pc *ParquetClient) Send(ctx context.Context, agentID string, args Arguments) error {
	buf, err := pc.Write()
	if err != nil {
		return err
	}

	fullEndpoint := fmt.Sprintf("%s/agents/%s", args.RemoteEndpoint, agentID)
	req, err := http.NewRequest(http.MethodPost, fullEndpoint, &buf)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/parquet")
	for key, value := range args.Headers {
		req.Header.Set(key, value)
	}

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
	writer := parquet.NewGenericWriter[riveragentstate.Component](&buf)

	// Write the component data to the buffer.
	rowGroup := parquet.NewGenericBuffer[riveragentstate.Component]()
	_, err := rowGroup.Write(c.components)
	if err != nil {
		return buf, err
	}

	_, err = writer.WriteRowGroup(rowGroup)
	if err != nil {
		return buf, err
	}

	// Write the metadata to the buffer.
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
