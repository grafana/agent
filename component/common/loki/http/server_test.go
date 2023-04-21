package http

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/go-kit/log"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"github.com/weaveworks/common/server"
)

const (
	localhost = "127.0.0.1"
)

func TestTargetServer(t *testing.T) {
	// dependencies
	w := log.NewSyncWriter(os.Stderr)
	logger := log.NewLogfmtLogger(w)
	reg := prometheus.NewRegistry()
	cfg, port, err := getServerConfigWithAvailablePort()
	require.NoError(t, err)

	ts, err := NewTargetServer(logger, "test_namespace", reg, ServerConfig{
		Server: cfg,
	})
	require.NoError(t, err)

	ts.MountAndRun(func(router *mux.Router) {
		router.Methods("GET").Path("/hello").Handler(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
	})
	defer ts.StopAndShutdown()

	// assert over endpoint getter
	require.Equal(t, fmt.Sprintf("127.0.0.1:%d", port), ts.HTTPListenAddr())

	// test mounted endpoint
	req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:%d/hello", port), nil)
	require.NoError(t, err)
	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, 200, res.StatusCode)

	// assert all metrics have the prefix applied
	metrics, err := reg.Gather()
	require.NoError(t, err)
	for _, m := range metrics {
		require.True(t, strings.HasPrefix(m.GetName(), "test_namespace"))
	}
}

func getServerConfigWithAvailablePort() (cfg server.Config, port int, err error) {
	// Get a randomly available port by open and closing a TCP socket
	addr, err := net.ResolveTCPAddr("tcp", localhost+":0")
	if err != nil {
		return
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return
	}
	port = l.Addr().(*net.TCPAddr).Port
	err = l.Close()
	if err != nil {
		return
	}

	// Adjust some of the defaults
	cfg.RegisterFlags(flag.NewFlagSet("empty", flag.ContinueOnError))
	cfg.HTTPListenAddress = localhost
	cfg.HTTPListenPort = port
	cfg.GRPCListenAddress = localhost
	cfg.GRPCListenPort = 0 // Not testing GRPC, a random port will be assigned

	return
}
