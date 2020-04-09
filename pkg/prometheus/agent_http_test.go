package prometheus

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"testing"

	"github.com/cortexproject/cortex/pkg/ring/kv"
	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
	"github.com/grafana/agent/pkg/prometheus/configapi"
	"github.com/prometheus/prometheus/config"
	"github.com/stretchr/testify/require"
)

// Going to do HTTP tests here including API paths.
func TestAgent_ListConfigurations(t *testing.T) {
	env := newApiTestEnvironment(t)

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

type apiTestEnvironment struct {
	agent  *Agent
	srv    *httptest.Server
	router *mux.Router
}

func newApiTestEnvironment(t *testing.T) apiTestEnvironment {
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

type testResponse struct {
	Status string          `json:"status"`
	Data   json.RawMessage `json:"data"`
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
