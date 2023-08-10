package observer

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/grafana/agent/component/common/config"
	prom_config "github.com/prometheus/common/config"
)

// httpAgentStateWriter sends the Agent state via HTTP
type httpAgentStateWriter struct {
	HttpClient     *http.Client
	AgentID        string
	RemoteEndpoint string
	Headers        map[string]string
	Context        context.Context
}

var _ agentStateWriter = (*httpAgentStateWriter)(nil)

func newHttpAgentStateWriter(httpConfig config.HTTPClientConfig, agentID string, remoteEndpoint string, headers map[string]string) (*httpAgentStateWriter, error) {
	httpClient, err := prom_config.NewClientFromConfig(
		*httpConfig.Convert(),
		"agent_observer",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create http client for HttpAgentStateWriter: %w", err)
	}

	return &httpAgentStateWriter{
		HttpClient:     httpClient,
		AgentID:        agentID,
		RemoteEndpoint: remoteEndpoint,
		Headers:        headers,
		Context:        context.Background(),
	}, nil
}

func (w *httpAgentStateWriter) Write(p []byte) (n int, err error) {
	fullEndpoint := fmt.Sprintf("%s/agents/%s", w.RemoteEndpoint, w.AgentID)
	agentStateReader := bytes.NewReader(p)

	req, err := http.NewRequest(http.MethodPost, fullEndpoint, agentStateReader)
	if err != nil {
		return 0, err
	}

	req.Header.Set("Content-Type", "application/parquet")
	for key, value := range w.Headers {
		req.Header.Set(key, value)
	}

	resp, err := w.HttpClient.Do(req.WithContext(w.Context))
	if err != nil {
		return 0, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return 0, err
		}
		return 0, fmt.Errorf("failed to send agent state: %s  body: %s", resp.Status, string(data))
	}

	return len(p), nil
}

func (w *httpAgentStateWriter) SetContext(ctx context.Context) {
	w.Context = ctx
}
