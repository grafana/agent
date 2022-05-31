// Package grafanacloud provides an interface to the Grafana Cloud API.
package grafanacloud

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"gopkg.in/yaml.v2"
)

const defaultAPIURL = "https://integrations-api.grafana.net"

// Client is a grafanacloud API client.
type Client struct {
	c      *http.Client
	apiKey string
	apiURL string
}

// NewClient creates a new Grafana Cloud client. All requests made will be
// performed using the provided http.Client c. If c is nil, the default
// http client will be used instead.
//
// apiKey will be used to authenticate against the apiURL.
func NewClient(c *http.Client, apiKey, apiURL string) *Client {
	if c == nil {
		c = http.DefaultClient
	}
	if apiURL == "" {
		apiURL = defaultAPIURL
	}
	return &Client{c: c, apiKey: apiKey, apiURL: apiURL}
}

// AgentConfig generates a Grafana Agent config from the given stack.
// The config is returned as a string in YAML form.
func (c *Client) AgentConfig(ctx context.Context, stackID, platforms string) (string, error) {
	url := fmt.Sprintf("%s/stacks/%s/agent_config", c.apiURL, stackID)
	if platforms != "" {
		url = fmt.Sprintf("%s?platforms=%s", url, platforms)
	}
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)

	if err != nil {
		return "", fmt.Errorf("failed to generate request: %w", err)
	}
	req.Header.Add("Authorization", "Bearer "+c.apiKey)

	resp, err := c.c.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Even though the API returns json, we'll parse it as YAML here so we can
	// re-encode it with the same order it was decoded in.
	payload := struct {
		Status string        `yaml:"status"`
		Data   yaml.MapSlice `yaml:"data"`
		Error  string        `yaml:"error"`
	}{}

	dec := yaml.NewDecoder(resp.Body)
	dec.SetStrict(true)
	if err := dec.Decode(&payload); err != nil {
		if resp.StatusCode != 200 {
			return "", fmt.Errorf("unexpected status code %d", resp.StatusCode)
		}

		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if payload.Status != "success" {
		return "", fmt.Errorf("request was not successful: %s", payload.Error)
	}

	// Convert the data to YAML
	var sb strings.Builder
	if err := yaml.NewEncoder(&sb).Encode(payload.Data); err != nil {
		return "", fmt.Errorf("failed to generate YAML config: %w", err)
	}

	return sb.String(), nil
}
