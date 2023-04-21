package http

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/go-kit/log"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func TestTargetServer(t *testing.T) {
	// dependencies
	w := log.NewSyncWriter(os.Stderr)
	logger := log.NewLogfmtLogger(w)
	reg := prometheus.NewRegistry()

	ts, err := NewTargetServer(logger, "test_namespace", reg, &ServerConfig{})
	require.NoError(t, err)

	ts.MountAndRun(func(router *mux.Router) {
		router.Methods("GET").Path("/hello").Handler(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
	})
	defer ts.StopAndShutdown()

	// test mounted endpoint
	req, err := http.NewRequest("GET", fmt.Sprintf("http://%s/hello", ts.HTTPListenAddr()), nil)
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
