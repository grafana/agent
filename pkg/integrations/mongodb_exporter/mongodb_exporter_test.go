package mongodb_exporter //nolint:golint

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
	"github.com/prometheus/prometheus/pkg/textparse"
	"github.com/stretchr/testify/require"
)

func TestMongoDbExporter(t *testing.T) {
	cfg := DefaultConfig

	cfg.URI = "mongodb://192.168.99.102:27017/"

	// Check that the flags convert and the integration initiailizes
	logger := log.NewNopLogger()
	integration, err := New(logger, &cfg)
	require.NoError(t, err, "failed to setup mongodb_exporter")

	r := mux.NewRouter()
	handler, err := integration.MetricsHandler()
	require.NoError(t, err)
	r.Handle("/metrics", handler)

	// Invoke /metrics and parse the response
	srv := httptest.NewServer(r)
	defer srv.Close()

	res, err := http.Get(srv.URL + "/metrics")
	require.NoError(t, err)

	body, err := ioutil.ReadAll(res.Body)
	require.NoError(t, err)

	p := textparse.NewPromParser(body)
	for {
		_, err := p.Next()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
	}

}
