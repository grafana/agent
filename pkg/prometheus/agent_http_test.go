package prometheus

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/cortexproject/cortex/pkg/ring/kv"
	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
	"github.com/grafana/agent/pkg/prometheus/configapi"
	"github.com/prometheus/prometheus/config"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestAgent_ListConfigurations(t *testing.T) {
	env := newAPITestEnvironment(t)

	// Store some configs
	cfgs := []*InstanceConfig{
		{Name: "a"},
		{Name: "b"},
		{Name: "c"},
	}
	for _, cfg := range cfgs {
		err := env.agent.kv.CAS(context.Background(), cfg.Name, func(in interface{}) (out interface{}, retry bool, err error) {
			return cfg, false, nil
		})
		require.NoError(t, err)
	}

	resp, err := http.Get(env.srv.URL + "/agent/api/v1/configs")
	require.NoError(t, err)

	var apiResp configapi.ListConfigurationsResponse
	unmarshalTestResponse(t, resp.Body, &apiResp)
	sort.Strings(apiResp.Configs)

	expect := configapi.ListConfigurationsResponse{Configs: []string{"a", "b", "c"}}
	require.Equal(t, expect, apiResp)
}

// TestAgent_GetConfiguration_Invalid makes sure that requesting an invalid
// config does not panic.
func TestAgent_GetConfiguration_Invalid(t *testing.T) {
	env := newAPITestEnvironment(t)

	resp, err := http.Get(env.srv.URL + "/agent/api/v1/configs/does-not-exist")
	require.NoError(t, err)

	var apiResp configapi.ErrorResponse
	unmarshalTestResponse(t, resp.Body, &apiResp)
	require.Equal(t, "configuration does-not-exist does not exist", apiResp.Error)
}

func TestAgent_GetConfiguration_YAML(t *testing.T) {
	env := newAPITestEnvironment(t)

	cfg := DefaultInstanceConfig
	cfg.Name = "a"
	cfg.HostFilter = true
	cfg.RemoteFlushDeadline = 10 * time.Minute
	err := env.agent.kv.CAS(context.Background(), cfg.Name, func(in interface{}) (out interface{}, retry bool, err error) {
		return &cfg, false, nil
	})
	require.NoError(t, err)

	resp, err := http.Get(env.srv.URL + "/agent/api/v1/configs/a")
	require.NoError(t, err)

	var apiResp configapi.GetConfigurationResponse
	unmarshalTestResponse(t, resp.Body, &apiResp)

	var actual InstanceConfig
	err = yaml.Unmarshal([]byte(apiResp.Value), &actual)
	require.NoError(t, err)

	require.Equal(t, cfg, actual, "unmarshaled stored configuration did not match input")
}

func TestAgent_PutConfiguration(t *testing.T) {
	env := newAPITestEnvironment(t)

	cfg := DefaultInstanceConfig
	cfg.Name = "newconfig"
	cfg.HostFilter = true
	cfg.RemoteFlushDeadline = 10 * time.Minute

	bb, err := yaml.Marshal(cfg)
	require.NoError(t, err)

	resp, err := http.Post(env.srv.URL+"/agent/api/v1/config/newconfig", "", bytes.NewReader(bb))
	require.NoError(t, err)
	unmarshalTestResponse(t, resp.Body, nil)

	// Get the stored config back
	resp, err = http.Get(env.srv.URL + "/agent/api/v1/configs/newconfig")
	require.NoError(t, err)
	var apiResp configapi.GetConfigurationResponse
	unmarshalTestResponse(t, resp.Body, &apiResp)

	var actual InstanceConfig
	err = yaml.Unmarshal([]byte(apiResp.Value), &actual)
	require.NoError(t, err)

	require.Equal(t, cfg, actual, "unmarshaled stored configuration did not match input")
}

func TestAgent_DeleteConfiguration(t *testing.T) {
	env := newAPITestEnvironment(t)

	// Store some configs
	cfgs := []*InstanceConfig{
		{Name: "a"},
		{Name: "b"},
		{Name: "c"},
	}
	for _, cfg := range cfgs {
		err := env.agent.kv.CAS(context.Background(), cfg.Name, func(in interface{}) (out interface{}, retry bool, err error) {
			return cfg, false, nil
		})
		require.NoError(t, err)
	}

	// Delete the configs
	for _, cfg := range cfgs {
		req, err := http.NewRequest("DELETE", env.srv.URL+"/agent/api/v1/config/"+cfg.Name, nil)
		require.NoError(t, err)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		unmarshalTestResponse(t, resp.Body, nil)
	}

	// Do a list, nothing we stored should be there anymore.
	resp, err := http.Get(env.srv.URL + "/agent/api/v1/configs")
	require.NoError(t, err)

	var apiResp configapi.ListConfigurationsResponse
	unmarshalTestResponse(t, resp.Body, &apiResp)
	for _, cfg := range cfgs {
		require.NotContains(t, apiResp.Configs, cfg.Name)
	}
}

type apiTestEnvironment struct {
	agent  *Agent
	srv    *httptest.Server
	router *mux.Router
}

func newAPITestEnvironment(t *testing.T) apiTestEnvironment {
	t.Helper()

	dir, err := ioutil.TempDir(os.TempDir(), "etcd_backend_test")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(dir) })

	router := mux.NewRouter()
	srv := httptest.NewServer(router)
	t.Cleanup(srv.Close)

	// Create a new agent with an HTTP store
	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	a, err := New(Config{
		WALDir: dir,
		Global: config.DefaultGlobalConfig,
		ServiceConfig: ServiceConfig{
			Enabled: true,
			KVStore: kv.Config{
				Store:  "inmemory",
				Prefix: "configs/",
			},
		},
	}, logger)
	require.NoError(t, err)
	t.Cleanup(a.Stop)

	// Wire the API
	a.WireAPI(router)

	return apiTestEnvironment{
		agent:  a,
		srv:    srv,
		router: router,
	}
}

// unmarshalTestResponse will unmarshal a test response's data to v. If v is
// nil, unmarshalTestResponse expects that the test response's data should be
// empty.
func unmarshalTestResponse(t *testing.T, r io.ReadCloser, v interface{}) {
	defer r.Close()
	t.Helper()

	resp := struct {
		Status string          `json:"status"`
		Data   json.RawMessage `json:"data"`
	}{}

	err := json.NewDecoder(r).Decode(&resp)
	require.NoError(t, err)

	if v == nil {
		require.True(t, len(resp.Data) == 0, "data in response was not empty as expected")
		return
	}

	err = json.Unmarshal(resp.Data, v)
	require.NoError(t, err)
}
