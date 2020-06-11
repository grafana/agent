// +build !race

package node_exporter //nolint:golint

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
	"github.com/prometheus/prometheus/pkg/textparse"
	"github.com/stretchr/testify/require"
)

// TestNodeExporter runs an integration test for node_exporter, doing the
// following:
//
// 1. Enabling all collectors (minus some that cause issues in cross-platform testing)
// 2. Creating the integration
// 3. Scrape the integration once
// 4. Parse the result of the scrape
//
// This ensures that the flag parsing is correct and that the handler is
// set up properly. We do not test the contents of the scrape, just that it
// was parsable by Prometheus.
func TestNodeExporter(t *testing.T) {
	cfg := DefaultConfig

	// Enable all collectors except perf
	cfg.SetCollectors = make([]string, 0, len(Collectors))
	for c := range Collectors {
		cfg.SetCollectors = append(cfg.SetCollectors, c)
	}
	cfg.DisableCollectors = []string{CollectorPerf, CollectorBuddyInfo}

	// Check that the flags convert and the integration initiailizes
	logger := log.NewNopLogger()
	integration, err := New(logger, cfg)
	require.NoError(t, err, "failed to setup node_exporter")

	r := mux.NewRouter()
	err = integration.RegisterRoutes(r)
	require.NoError(t, err)

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

// TestFTestNodeExporter_IgnoredFlags ensures that flags don't get ignored for
// misspellings.
func TestNodeExporter_IgnoredFlags(t *testing.T) {
	cfg := DefaultConfig

	// Enable all collectors except perf
	cfg.SetCollectors = make([]string, 0, len(Collectors))
	for c := range Collectors {
		cfg.SetCollectors = append(cfg.SetCollectors, c)
	}
	cfg.DisableCollectors = []string{CollectorPerf}

	_, ignored := MapConfigToNodeExporterFlags(&cfg)
	var expect []string

	switch runtime.GOOS {
	case "darwin":
		expect = []string{"collector.cpu.info", "collector.diskstats.ignored-devices"}
	}

	require.Equal(t, expect, ignored)
}

// TestFlags makes sure that boolean flags and some known non-boolean flags
// work as expected
func TestFlags(t *testing.T) {
	var f flags
	f.add("--path.rootfs", "/")
	require.Equal(t, []string{"--path.rootfs", "/"}, f.accepted)

	// Set up booleans to use as pointers
	var (
		truth = true

		// You know, the opposite of truth?
		falth = false
	)

	f = flags{}
	f.addBools(map[*bool]string{&truth: "collector.textfile"})
	require.Equal(t, []string{"--collector.textfile"}, f.accepted)

	f = flags{}
	f.addBools(map[*bool]string{&falth: "collector.textfile"})
	require.Equal(t, []string{"--no-collector.textfile"}, f.accepted)
}
