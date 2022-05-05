package grafanacloud

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testSecret  = "secret-key"
	testStackID = "12345"
)

func TestClient_AgentConfig(t *testing.T) {
	httpClient := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/stacks/"+testStackID+"/agent_config", r.URL.Path)
		assert.Equal(t, "Bearer "+testSecret, r.Header.Get("Authorization"))

		_, err := w.Write([]byte(`{
			"status": "success",
			"data": {
				"server": {
					"log_level": "debug"
				},
				"integrations": {
					"agent": {
						"enabled": true
					}
				}
			}
		}`))
		assert.NoError(t, err)
	}))

	cli := NewClient(httpClient, testSecret, "")
	cfg, err := cli.AgentConfig(context.Background(), testStackID)
	require.NoError(t, err)
	fmt.Println(cfg)

	expect := `
server:
  log_level: debug
integrations:
  agent:
    enabled: true
`

	require.YAMLEq(t, expect, cfg)
}

func TestClient_AgentConfig_Error(t *testing.T) {
	httpClient := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))

	cli := NewClient(httpClient, testSecret, "")
	_, err := cli.AgentConfig(context.Background(), testStackID)
	require.Error(t, err, "unexpected status code 404")
}

func TestClient_AgentConfig_ErrorMessage(t *testing.T) {
	httpClient := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte(`{
			"status": "error",
			"error": "Something went wrong"
		}`))
		assert.NoError(t, err)
	}))

	cli := NewClient(httpClient, testSecret, "")
	_, err := cli.AgentConfig(context.Background(), testStackID)
	require.Error(t, err, "request was not successful: Something went wrong")
}

func testClient(t *testing.T, handler http.HandlerFunc) *http.Client {
	h := httptest.NewTLSServer(handler)
	t.Cleanup(func() {
		h.Close()
	})

	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return net.Dial(network, h.Listener.Addr().String())
			},
		},
	}
}
