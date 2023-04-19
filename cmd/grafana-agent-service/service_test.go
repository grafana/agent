package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"testing"

	"github.com/go-kit/log"
	"github.com/grafana/agent/pkg/flow/componenttest"
	"github.com/grafana/agent/pkg/util"
	"github.com/phayes/freeport"
	"github.com/stretchr/testify/require"
)

const goosWindows = "windows"

func Test_serviceManager(t *testing.T) {
	l := util.TestLogger(t)

	serviceBinary := buildExampleService(t, l)

	t.Run("can run service binary", func(t *testing.T) {
		listenHost := getListenHost(t)

		mgr := newServiceManager(l, serviceManagerConfig{
			Path: serviceBinary,
			Args: []string{"-listen-addr", listenHost},
		})
		go mgr.Run(componenttest.TestContext(t))

		util.Eventually(t, func(t require.TestingT) {
			resp, err := makeServiceRequest(listenHost, "/echo/response", []byte("Hello, world!"))
			require.NoError(t, err)
			require.Equal(t, []byte("Hello, world!"), resp)
		})
	})

	t.Run("terminates service binary", func(t *testing.T) {
		listenHost := getListenHost(t)

		mgr := newServiceManager(l, serviceManagerConfig{
			Path: serviceBinary,
			Args: []string{"-listen-addr", listenHost},
		})

		ctx, cancel := context.WithCancel(componenttest.TestContext(t))
		defer cancel()
		go mgr.Run(ctx)

		util.Eventually(t, func(t require.TestingT) {
			resp, err := makeServiceRequest(listenHost, "/echo/response", []byte("Hello, world!"))
			require.NoError(t, err)
			require.Equal(t, []byte("Hello, world!"), resp)
		})

		// Cancel the context, which should stop the manager.
		cancel()

		util.Eventually(t, func(t require.TestingT) {
			_, err := makeServiceRequest(listenHost, "/echo/response", []byte("Hello, world!"))

			if runtime.GOOS == goosWindows {
				require.ErrorContains(t, err, "No connection could be made")
			} else {
				require.ErrorContains(t, err, "connection refused")
			}
		})
	})

	t.Run("can forward to stdout", func(t *testing.T) {
		listenHost := getListenHost(t)

		var buf syncBuffer

		mgr := newServiceManager(l, serviceManagerConfig{
			Path:   serviceBinary,
			Args:   []string{"-listen-addr", listenHost},
			Stdout: &buf,
		})

		ctx, cancel := context.WithCancel(componenttest.TestContext(t))
		defer cancel()
		go mgr.Run(ctx)

		// Test making the request and testing the buffer contents separately,
		// otherwise we may log to stdout more than we intend to.

		util.Eventually(t, func(t require.TestingT) {
			_, err := makeServiceRequest(listenHost, "/echo/stdout", []byte("Hello, world!"))
			require.NoError(t, err)
		})

		util.Eventually(t, func(t require.TestingT) {
			require.Equal(t, []byte("Hello, world!"), buf.Bytes())
		})
	})

	t.Run("can forward to stderr", func(t *testing.T) {
		listenHost := getListenHost(t)

		var buf syncBuffer

		mgr := newServiceManager(l, serviceManagerConfig{
			Path:   serviceBinary,
			Args:   []string{"-listen-addr", listenHost},
			Stderr: &buf,
		})

		ctx, cancel := context.WithCancel(componenttest.TestContext(t))
		defer cancel()
		go mgr.Run(ctx)

		// Test making the request and testing the buffer contents separately,
		// otherwise we may log to stderr more than we intend to.

		util.Eventually(t, func(t require.TestingT) {
			_, err := makeServiceRequest(listenHost, "/echo/stderr", []byte("Hello, world!"))
			require.NoError(t, err)
		})

		util.Eventually(t, func(t require.TestingT) {
			require.Equal(t, []byte("Hello, world!"), buf.Bytes())
		})
	})
}

func buildExampleService(t *testing.T, l log.Logger) string {
	t.Helper()

	writer := log.NewStdlibAdapter(l)

	servicePath := filepath.Join(t.TempDir(), "example-service")
	if runtime.GOOS == goosWindows {
		servicePath = servicePath + ".exe"
	}

	cmd := exec.Command(
		"go", "build",
		"-o", servicePath,
		"testdata/example_service.go",
	)
	cmd.Stdout = writer
	cmd.Stderr = writer

	require.NoError(t, cmd.Run())

	return servicePath
}

func getListenHost(t *testing.T) string {
	t.Helper()

	port, err := freeport.GetFreePort()
	require.NoError(t, err)

	return fmt.Sprintf("127.0.0.1:%d", port)
}

func makeServiceRequest(host string, path string, body []byte) ([]byte, error) {
	resp, err := http.Post(
		fmt.Sprintf("http://%s%s", host, path),
		"text/plain",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %s", resp.Status)
	}
	return io.ReadAll(resp.Body)
}

// syncBuffer wraps around a bytes.Buffer and makes it safe to use from
// multiple goroutines.
type syncBuffer struct {
	mut sync.RWMutex
	buf bytes.Buffer
}

func (sb *syncBuffer) Bytes() []byte {
	sb.mut.RLock()
	defer sb.mut.RUnlock()

	return sb.buf.Bytes()
}

func (sb *syncBuffer) Write(p []byte) (n int, err error) {
	sb.mut.Lock()
	defer sb.mut.Unlock()

	return sb.buf.Write(p)
}
