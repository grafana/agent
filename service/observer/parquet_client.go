package observer

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/grafana/agent/component/common/config"
	prom_config "github.com/prometheus/common/config"
)

// HttpAgentStateWriter sends the Agent state via HTTP
type HttpAgentStateWriter struct {
	HttpClient     *http.Client
	AgentID        string
	RemoteEndpoint string
	Headers        map[string]string
}

var _ AgentStateWriter = (*HttpAgentStateWriter)(nil)

func NewHttpAgentStateWriter(httpConfig config.HTTPClientConfig, agentID string, remoteEndpoint string, headers map[string]string) (*HttpAgentStateWriter, error) {
	httpClient, err := prom_config.NewClientFromConfig(
		*httpConfig.Convert(),
		"agent_observer",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create http client for HttpAgentStateWriter: %w", err)
	}

	return &HttpAgentStateWriter{
		HttpClient:     httpClient,
		AgentID:        agentID,
		RemoteEndpoint: remoteEndpoint,
		Headers:        headers,
	}, nil
}

func (w *HttpAgentStateWriter) Write(ctx context.Context, agentState []byte) error {
	fullEndpoint := fmt.Sprintf("%s/agents/%s", w.RemoteEndpoint, w.AgentID)
	agentStateReader := bytes.NewReader(agentState)

	req, err := http.NewRequest(http.MethodPost, fullEndpoint, agentStateReader)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/parquet")
	for key, value := range w.Headers {
		req.Header.Set(key, value)
	}

	resp, err := w.HttpClient.Do(req.WithContext(ctx))
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

// FileAgentStateWriter writes the Agent state to a file
type FileAgentStateWriter struct {
	filepath string
}

var _ AgentStateWriter = (*FileAgentStateWriter)(nil)

func (w *FileAgentStateWriter) Write(_ context.Context, agentState []byte) error {
	f, err := os.Create(w.filepath)
	if err != nil {
		return err
	}
	_, err = f.Write(agentState)
	return err
}
