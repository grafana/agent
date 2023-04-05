//go:build !race
// +build !race

package node_exporter //nolint:golint

import (
	"io"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/gorilla/mux"
	"github.com/grafana/agent/pkg/util"
	"github.com/prometheus/prometheus/model/textparse"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/alecthomas/kingpin.v2"
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

	// Check that the flags convert and the integration initializes
	logger := log.NewNopLogger()
	integration, err := New(logger, &cfg)
	require.NoError(t, err, "failed to setup node_exporter")

	r := mux.NewRouter()
	handler, err := integration.MetricsHandler()
	require.NoError(t, err)
	r.Handle("/metrics", handler)

	// Invoke /metrics and parse the response
	srv := httptest.NewServer(r)
	defer srv.Close()

	res, err := http.Get(srv.URL + "/metrics")
	require.NoError(t, err)

	body, err := io.ReadAll(res.Body)
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
	l := util.TestLogger(t)
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
		expect = []string{
			"collector.cpu.info",
			"collector.cpu.guest",
			"collector.cpu.info.flags-include",
			"collector.cpu.info.bugs-include",
			"collector.filesystem.mount-timeout",
		}
	}

	if !assert.ElementsMatch(t, expect, ignored) {
		level.Debug(l).Log("msg", "printing available flags")
		for _, flag := range kingpin.CommandLine.Model().Flags {
			level.Debug(l).Log("flag", flag.Name, "hidden", flag.Hidden)
		}
	}
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
