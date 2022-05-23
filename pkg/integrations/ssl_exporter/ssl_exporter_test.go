// Package ssl_exporter embeds https://github.com/ribbybibby/ssl_exporter/v2
package ssl_exporter

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-kit/log"
	"github.com/gorilla/mux"
	"github.com/prometheus/prometheus/model/textparse"
	"github.com/stretchr/testify/require"
)

func TestSSLCases(t *testing.T) {
	tests := []struct {
		name                   string
		cfg                    Config
		expectConstructorError bool
		expectedMetrics        []string
	}{
		// Test that default config results in some metrics that can be parsed by
		// prometheus.
		{
			name: "Default config",
			cfg: (func() Config {
				c := DefaultConfig
				return c
			})(),
			expectedMetrics: []string{},
		},
		// Test that some invalid pre-constructor config logic causes an error.
		{
			name: "Config file read fail",
			cfg: (func() Config {
				c := DefaultConfig
				c.ConfigFile = "/does/not/exist"
				return c
			})(),
			expectConstructorError: true,
		},
		// Test that a custom config file can be parsed and used.
		{
			name: "Config file read success",
			cfg: (func() Config {
				c := DefaultConfig
				c.ConfigFile = "test-config.yaml"
				return c
			})(),
		},
	}

	logger := log.NewNopLogger()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			integration, err := New(logger, &test.cfg)

			if test.expectConstructorError {
				require.Error(t, err, "expected failure when setting up ssl_exporter")
				return
			}
			require.NoError(t, err, "failed to setup ssl_exporter")

			r := mux.NewRouter()
			handler, err := integration.MetricsHandler()
			require.NoError(t, err)
			r.Handle("/metrics", handler)
			require.NoError(t, err)

			srv := httptest.NewServer(r)
			defer srv.Close()
			res, err := http.Get(srv.URL + "/metrics?target=example.com:443")
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
