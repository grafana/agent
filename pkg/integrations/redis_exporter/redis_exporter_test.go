package redis_exporter //nolint:golint

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
	"github.com/prometheus/prometheus/pkg/textparse"
	"github.com/stretchr/testify/require"
)

const addr string = "localhost:6379"

type testCase struct {
	name                   string
	cfg                    Config
	expectedMetrics        []string
	expectConstructorError bool
	envVars                map[string]string
}

var testCases = []testCase{
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
	// Test that exporter picks up env var
	{
		name: "address from env OK",
		cfg:  Config{}, // no address in here
		envVars: map[string]string{
			"REDIS_EXPORTER_ADDRESS": "redis:1234",
		},
	},
	// Test exporter complains when no address given via env or config.
	{
		name:                   "no address given",
		cfg:                    Config{}, // no address in here
		expectConstructorError: true,
	},
}

func TestRedisCases(t *testing.T) {
	logger := log.NewNopLogger()

	for _, test := range testCases {

		// Pre-test actions
		if len(test.envVars) > 0 {
			for k, v := range test.envVars {
				os.Setenv(k, v)
			}
		}

		// Test logic
		cfg := test.cfg

		integration, err := New(logger, cfg)
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
			require.True(t, exists, "case: %s - could not find metric %s", test.name, metric)
		}

		// Post-test actions
		if len(test.envVars) > 0 {
			for k := range test.envVars {
				os.Unsetenv(k)
			}
		}

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
