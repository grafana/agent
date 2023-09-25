package http

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/pkg/flow"
	"github.com/grafana/agent/pkg/flow/componenttest"
	"github.com/grafana/agent/pkg/util"
	"github.com/grafana/agent/service"
	"github.com/grafana/river"
	"github.com/phayes/freeport"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/config"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
)

func TestHTTP(t *testing.T) {
	ctx := componenttest.TestContext(t)

	env, err := newTestEnvironment(t)
	require.NoError(t, err)
	require.NoError(t, env.ApplyConfig(`/* empty */`))

	go func() {
		require.NoError(t, env.Run(ctx))
	}()

	util.Eventually(t, func(t require.TestingT) {
		cli, err := config.NewClientFromConfig(config.HTTPClientConfig{}, "test")
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s/-/ready", env.ListenAddr()), nil)
		require.NoError(t, err)

		resp, err := cli.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestTLS(t *testing.T) {
	ctx := componenttest.TestContext(t)

	env, err := newTestEnvironment(t)
	require.NoError(t, err)
	require.NoError(t, env.ApplyConfig(`
		tls {
			cert_file = "testdata/test-cert.crt"
			key_file = "testdata/test-key.key"
		}
	`))

	go func() {
		require.NoError(t, env.Run(ctx))
	}()

	util.Eventually(t, func(t require.TestingT) {
		cli, err := config.NewClientFromConfig(config.HTTPClientConfig{
			TLSConfig: config.TLSConfig{
				CAFile: "testdata/test-cert.crt",
			},
		}, "test")
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://%s/-/ready", env.ListenAddr()), nil)
		require.NoError(t, err)

		resp, err := cli.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func Test_Toggle_TLS(t *testing.T) {
	ctx := componenttest.TestContext(t)

	env, err := newTestEnvironment(t)
	require.NoError(t, err)

	go func() {
		require.NoError(t, env.Run(ctx))
	}()

	{
		// Start with plain HTTP.
		require.NoError(t, env.ApplyConfig(`/* empty */`))
		util.Eventually(t, func(t require.TestingT) {
			cli, err := config.NewClientFromConfig(config.HTTPClientConfig{}, "test")
			require.NoError(t, err)

			req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s/-/ready", env.ListenAddr()), nil)
			require.NoError(t, err)

			resp, err := cli.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			require.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}

	{
		// Toggle TLS.
		require.NoError(t, env.ApplyConfig(`
			tls {
				cert_file = "testdata/test-cert.crt"
				key_file = "testdata/test-key.key"
			}
		`))

		util.Eventually(t, func(t require.TestingT) {
			cli, err := config.NewClientFromConfig(config.HTTPClientConfig{
				TLSConfig: config.TLSConfig{
					CAFile: "testdata/test-cert.crt",
				},
			}, "test")
			require.NoError(t, err)

			req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://%s/-/ready", env.ListenAddr()), nil)
			require.NoError(t, err)

			resp, err := cli.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			require.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}

	{
		// Disable TLS.
		require.NoError(t, env.ApplyConfig(`/* empty */`))
		util.Eventually(t, func(t require.TestingT) {
			cli, err := config.NewClientFromConfig(config.HTTPClientConfig{}, "test")
			require.NoError(t, err)

			req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s/-/ready", env.ListenAddr()), nil)
			require.NoError(t, err)

			resp, err := cli.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			require.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

type testEnvironment struct {
	svc  *Service
	addr string
}

func newTestEnvironment(t *testing.T) (*testEnvironment, error) {
	port, err := freeport.GetFreePort()
	if err != nil {
		return nil, err
	}

	svc := New(Options{
		Logger:   util.TestLogger(t),
		Tracer:   trace.NewNoopTracerProvider(),
		Gatherer: prometheus.NewRegistry(),

		ReadyFunc:  func() bool { return true },
		ReloadFunc: func() (*flow.Source, error) { return nil, nil },

		HTTPListenAddr:   fmt.Sprintf("127.0.0.1:%d", port),
		MemoryListenAddr: "agent.internal:12345",
		EnablePProf:      true,
	})

	return &testEnvironment{
		svc:  svc,
		addr: fmt.Sprintf("127.0.0.1:%d", port),
	}, nil
}

func (env *testEnvironment) ApplyConfig(config string) error {
	var args Arguments
	if err := river.Unmarshal([]byte(config), &args); err != nil {
		return err
	}
	return env.svc.Update(args)
}

func (env *testEnvironment) Run(ctx context.Context) error {
	return env.svc.Run(ctx, fakeHost{})
}

func (env *testEnvironment) ListenAddr() string { return env.addr }

type fakeHost struct{}

var _ service.Host = (fakeHost{})

func (fakeHost) GetComponent(id component.ID, opts component.InfoOptions) (*component.Info, error) {
	return nil, fmt.Errorf("no such component %s", id)
}

func (fakeHost) ListComponents(moduleID string, opts component.InfoOptions) ([]*component.Info, error) {
	if moduleID == "" {
		return nil, nil
	}
	return nil, fmt.Errorf("no such module %q", moduleID)
}

func (fakeHost) GetServiceConsumers(serviceName string) []service.Consumer { return nil }
