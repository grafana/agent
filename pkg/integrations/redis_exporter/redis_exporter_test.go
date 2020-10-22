package redis_exporter //nolint:golint

import (
	"bytes"
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

const addr string = "localhost:6379"

func TestRedisCases(t *testing.T) {
	tt := []struct {
		name                   string
		cfg                    Config
		expectedMetrics        []string
		expectConstructorError bool
	}{
		// Test that default config results in some metrics that can be parsed by
		// prometheus.
		{
			name: "Default config",
			cfg: (func() Config {
				c := DefaultConfig
				c.RedisAddr = addr
				return c
			})(),
			expectedMetrics: []string{},
		},
		// Test that exporter metrics are included when configured to do so.
		{
			name: "Include exporter metrics",
			cfg: (func() Config {
				c := DefaultConfig
				c.RedisAddr = addr
				c.IncludeExporterMetrics = true
				return c
			})(),
			expectedMetrics: []string{
				"promhttp_metric_handler_requests_total",
				"promhttp_metric_handler_requests_in_flight",
			},
		},
		// Test that some valid pre-constructor config logic doesn't cause errors.
		{
			name: "Lua script read OK",
			cfg: (func() Config {
				c := DefaultConfig
				c.RedisAddr = addr
				c.ScriptPath = "./config.go" // file content is irrelevant
				return c
			})(),
		},
		// Test that some invalid pre-constructor config logic causes an error.
		{
			name: "Lua script read fail",
			cfg: (func() Config {
				c := DefaultConfig
				c.RedisAddr = addr
				c.ScriptPath = "/does/not/exist"
				return c
			})(),
			expectConstructorError: true,
		},
		// Test exporter complains when no address given via env or config.
		{
			name:                   "no address given",
			cfg:                    Config{}, // no address in here
			expectConstructorError: true,
		},
		// Test exporter constructs ok when password file is defined and exists
		{
			name: "valid password file",
			cfg: (func() Config {
				c := DefaultConfig
				c.RedisAddr = addr
				c.RedisPasswordFile = "./config.go" // contents not important
				return c
			})(),
		},
		// Test exporter construction fails when password file is defined and doesnt
		// exist
		{
			name: "invalid password file",
			cfg: (func() Config {
				c := DefaultConfig
				c.RedisAddr = addr
				c.RedisPasswordFile = "/does/not/exist"
				return c
			})(),
			expectConstructorError: true,
		},
	}

	logger := log.NewNopLogger()

	for _, test := range tt {

		t.Run(test.name, func(t *testing.T) {
			integration, err := New(logger, test.cfg)

			if test.expectConstructorError {
				require.Error(t, err, "expected failure when setting up redis_exporter")
				return
			}
			require.NoError(t, err, "failed to setup redis_exporter")

			r := mux.NewRouter()
			err = integration.RegisterRoutes(r)
			require.NoError(t, err)

			srv := httptest.NewServer(r)
			defer srv.Close()

			res, err := http.Get(srv.URL + "/metrics")
			require.NoError(t, err)

			body, err := ioutil.ReadAll(res.Body)
			require.NoError(t, err)

			foundMetricNames := map[string]bool{}
			for _, name := range test.expectedMetrics {
				foundMetricNames[name] = false
			}

			p := textparse.NewPromParser(body)
			for {
				entry, err := p.Next()
				if err == io.EOF {
					break
				}
				require.NoError(t, err)

				if entry == textparse.EntryHelp {
					matchMetricNames(foundMetricNames, p)
				}
			}

			for metric, exists := range foundMetricNames {
				require.True(t, exists, "could not find metric %s", metric)
			}
		})

	}
}

func matchMetricNames(names map[string]bool, p textparse.Parser) {
	for name := range names {
		metricName, _ := p.Help()
		if bytes.Equal([]byte(name), metricName) {
			names[name] = true
		}
	}
}
