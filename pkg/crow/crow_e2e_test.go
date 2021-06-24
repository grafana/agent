// +build has_docker,has_network

package crow

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/grafana/agent/pkg/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
)

func TestCrow_validate(t *testing.T) {
	var (
		validateReg = prometheus.NewRegistry()
		httpAddr    = exposeValidate(t, validateReg)
	)

	// Create a sample generator
	sendCh := make(chan []*sample, 5)
	gen := &sampleGenerator{
		numSamples: 1,
		sendCh:     sendCh,
		r:          rand.New(rand.NewSource(0)),
	}
	validateReg.MustRegister(gen)

	promAddr, promContainerAddr := launchPrometheusWithConfig(t, ``)

	launchAgentWithConfig(t, fmt.Sprintf(
		util.Untab(`
prometheus:
  global:
		scrape_interval: 5s
		external_labels:
			cluster: test
	configs:
	- name: default
		scrape_configs:
		- job_name: crow
			metrics_path: /validate
			static_configs:
				- targets: ['host.docker.internal:%d']
		remote_write:
		- url: http://%s/api/v1/write
`),
		httpAddr.(*net.TCPAddr).Port,
		promContainerAddr,
	))

	// Wait for a scrape to happen
	var s *sample
	select {
	case <-time.After(30 * time.Second):
		require.FailNow(t, "timed out waiting for a scrape")
	case samples := <-sendCh:
		s = samples[0]
	}

	crowCfg := DefaultConfig
	crowCfg.PrometheusAddr = "http://" + promAddr
	crowCfg.GenerateSamples = 1
	crowCfg.ExtraSelectors = `cluster="test"`
	crowCfg.UserID = "user"

	c, err := newCrow(crowCfg)
	require.NoError(t, err)

	for i := 0; i < 10; i++ {
		time.Sleep(time.Second)
		err = c.validate(s)
		if err == nil {
			break
		}
	}
	require.NoError(t, err)
}

func exposeValidate(t *testing.T, g prometheus.Gatherer) net.Addr {
	t.Helper()

	lis, err := net.Listen("tcp", "0.0.0.0:0")
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = lis.Close()
	})

	var (
		opts   promhttp.HandlerOpts
		router = mux.NewRouter()
	)
	router.Handle("/validate", promhttp.HandlerFor(g, opts))
	go func() {
		_ = http.Serve(lis, router)
	}()

	return lis.Addr()
}

func launchAgentWithConfig(t *testing.T, config string) (hostAddr, containerAddr string) {
	t.Helper()

	// Create config for Agent to use
	f, err := ioutil.TempFile(os.TempDir(), "agent-conf-*.yml")
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = os.RemoveAll(f.Name())
	})
	defer f.Close()
	_, err = io.Copy(f, strings.NewReader(config))
	require.NoError(t, err)

	req := testcontainers.GenericContainerRequest{
		Started: true,
		ContainerRequest: testcontainers.ContainerRequest{
			Image: "grafana/agent:v0.16.1",
			BindMounts: map[string]string{
				f.Name(): "/etc/agent.yaml",
			},
			Entrypoint: []string{
				"/bin/agent",
				"-log.level=debug",
				"-config.file=/etc/agent.yaml",
				"-prometheus.wal-directory=/tmp/agent",
			},
			ExposedPorts: []string{"80/tcp"},
		},
	}

	ctx := context.Background()
	c, err := testcontainers.GenericContainer(ctx, req)
	require.NoError(t, err)
	t.Cleanup(func() {
		c.Terminate(context.Background())
	})

	// If the test failed, print logs on exit
	t.Cleanup(func() {
		if !t.Failed() {
			return
		}

		r, err := c.Logs(ctx)
		if err != nil {
			return
		}

		var sb strings.Builder
		fmt.Fprintf(&sb, "=== start agent container logs ===\n")
		io.Copy(&sb, r)
		fmt.Fprintf(&sb, "=== end agent container logs ===\n")
		fmt.Println(sb.String())
	})

	containerIP, err := c.ContainerIP(ctx)
	require.NoError(t, err)
	hostIP, err := c.Host(ctx)
	require.NoError(t, err)
	port, err := c.MappedPort(ctx, "80")
	require.NoError(t, err)
	return fmt.Sprintf("%s:%d", hostIP, port.Int()), fmt.Sprintf("%s:80", containerIP)
}

func launchPrometheusWithConfig(t *testing.T, config string) (hostAddr, containerAddr string) {
	t.Helper()

	// Create config for Agent to use
	f, err := ioutil.TempFile(os.TempDir(), "prometheus-conf-*.yml")
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = os.RemoveAll(f.Name())
	})
	defer f.Close()
	_, err = io.Copy(f, strings.NewReader(config))
	require.NoError(t, err)

	req := testcontainers.GenericContainerRequest{
		Started: true,
		ContainerRequest: testcontainers.ContainerRequest{
			Image: "prom/prometheus:v2.26.0",
			BindMounts: map[string]string{
				f.Name(): "/etc/prometheus.yml",
			},
			Cmd: []string{
				"--config.file=/etc/prometheus.yml",
				"--enable-feature=remote-write-receiver",
			},
			ExposedPorts: []string{"9090/tcp"},
		},
	}

	ctx := context.Background()
	c, err := testcontainers.GenericContainer(ctx, req)
	require.NoError(t, err)
	t.Cleanup(func() {
		c.Terminate(context.Background())
	})

	containerIP, err := c.ContainerIP(ctx)
	require.NoError(t, err)
	hostIP, err := c.Host(ctx)
	require.NoError(t, err)
	port, err := c.MappedPort(ctx, "9090")
	require.NoError(t, err)
	return fmt.Sprintf("%s:%d", hostIP, port.Int()), fmt.Sprintf("%s:9090", containerIP)
}
