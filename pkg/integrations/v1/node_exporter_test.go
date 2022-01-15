//go:build !race
// +build !race

package v1 //nolint:golint

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"

	"github.com/grafana/agent/pkg/integrations/v1/node_exporter"

	"github.com/go-kit/log/level"
	"github.com/grafana/agent/pkg/util"
	"github.com/stretchr/testify/assert"
	"gopkg.in/alecthomas/kingpin.v2"
)

// TestFTestNodeExporter_IgnoredFlags ensures that flags don't get ignored for
// misspellings.
func TestNodeExporter_IgnoredFlags(t *testing.T) {
	l := util.TestLogger(t)
	cfg := node_exporter.DefaultConfig

	// Enable all collectors except perf
	cfg.SetCollectors = make([]string, 0, len(node_exporter.Collectors))
	for c := range node_exporter.Collectors {
		cfg.SetCollectors = append(cfg.SetCollectors, c)
	}
	cfg.DisableCollectors = []string{node_exporter.CollectorPerf}

	_, ignored := node_exporter.MapConfigToNodeExporterFlags(&cfg)
	var expect []string

	switch runtime.GOOS {
	case "darwin":
		expect = []string{
			"collector.cpu.info",
			"collector.cpu.guest",
			"collector.cpu.info.flags-include",
			"collector.cpu.info.bugs-include",
			"collector.diskstats.ignored-devices",
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

func TestNodeExporter_Config(t *testing.T) {
	var c NodeExporter

	err := yaml.Unmarshal([]byte("{}"), &c)
	require.NoError(t, err)
	require.Equal(t, node_exporter.DefaultConfig, c.Config)
}
