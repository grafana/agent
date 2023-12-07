package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/common/config"
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/river"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
	"gotest.tools/assert"
)

func TestRiverConfig(t *testing.T) {
	var exampleRiverConfig = `
	url = "https://www.example.com:12345/foo"
	refresh_interval = "14s"
	basic_auth {
		username = "123"
		password = "456"
	}
`
	var args Arguments
	err := river.Unmarshal([]byte(exampleRiverConfig), &args)
	require.NoError(t, err)
	assert.Equal(t, args.HTTPClientConfig.BasicAuth.Username, "123")
}

func TestConvert(t *testing.T) {
	args := DefaultArguments
	u, err := url.Parse("https://www.example.com:12345/foo")
	require.NoError(t, err)
	args.URL = config.URL{URL: u}

	sd := args.Convert()
	assert.Equal(t, "https://www.example.com:12345/foo", sd.URL)
	assert.Equal(t, model.Duration(60*time.Second), sd.RefreshInterval)
	assert.Equal(t, true, sd.HTTPClientConfig.EnableHTTP2)
}

func TestComponent(t *testing.T) {
	discovery.MaxUpdateFrequency = time.Second / 2
	endpointCalled := false
	var stateChanged atomic.Bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		endpointCalled = true
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(200)
		// from https://prometheus.io/docs/prometheus/latest/http_sd/
		w.Write([]byte(`[
			{
				"targets": ["10.0.10.2:9100", "10.0.10.3:9100", "10.0.10.4:9100", "10.0.10.5:9100"],
				"labels": {
					"__meta_datacenter": "london",
					"__meta_prometheus_job": "node"
				}
			},
			{
				"targets": ["10.0.40.2:9100", "10.0.40.3:9100"],
				"labels": {
					"__meta_datacenter": "london",
					"__meta_prometheus_job": "alertmanager"
				}
			},
			{
				"targets": ["10.0.40.2:9093", "10.0.40.3:9093"],
				"labels": {
					"__meta_datacenter": "newyork",
					"__meta_prometheus_job": "alertmanager"
				}
			}
		]`))
	}))
	defer srv.Close()
	u, _ := url.Parse(srv.URL)
	var cancel func()
	component, err := New(
		component.Options{
			OnStateChange: func(e component.Exports) {
				stateChanged.Store(true)
				args, ok := e.(discovery.Exports)
				assert.Equal(t, true, ok)
				assert.Equal(t, 8, len(args.Targets))
				cancel()
			},
		},
		Arguments{
			RefreshInterval:  time.Second,
			HTTPClientConfig: config.DefaultHTTPClientConfig,
			URL: config.URL{
				URL: u,
			},
		})
	assert.NilError(t, err)
	wg := sync.WaitGroup{}
	var ctx context.Context
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	wg.Add(1)
	go func() {
		err := component.Run(ctx)
		assert.NilError(t, err)
		wg.Done()
	}()
	wg.Wait()
	cancel()
	assert.Equal(t, true, endpointCalled)
	assert.Equal(t, true, stateChanged.Load())
}
