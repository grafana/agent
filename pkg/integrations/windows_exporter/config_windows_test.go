//go:build windows

package windows_exporter

import (
	"testing"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/collector"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestConfig(t *testing.T) {
	built, total := testConfig(t, "")
	// Default which is windows_exporter defaults minus textfile
	require.Len(t, built, 7)
	// Total should be 50
	require.Len(t, total, 50)
}

func TestMultipleConfig(t *testing.T) {
	cfg1 := `
enabled_collectors: "mssql,os"
mssql:
  enabled_classes: "accessmethods,availreplica"
`
	cfg2 := `
enabled_collectors: "mssql,os,cpu"
mssql:
  enabled_classes: "accessmethods,availreplica,bufman"
`
	cfg3 := `
enabled_collectors: "mssql,os"
mssql: {}  
`
	built1, total1 := testConfig(t, cfg1)
	built2, total2 := testConfig(t, cfg2)
	built3, total3 := testConfig(t, cfg3)
	require.Len(t, built1, 2)
	require.Len(t, built2, 3)
	require.Len(t, built3, 2)
	total1mssql := "accessmethods,availreplica"
	require.True(t, *total1["mssql"].Settings.(*collector.MSSqlSettings).ClassesEnabled == total1mssql)
	total2mssql := "accessmethods,availreplica,bufman"
	require.True(t, *total2["mssql"].Settings.(*collector.MSSqlSettings).ClassesEnabled == total2mssql)
	total3mssql := "accessmethods,availreplica,bufman,databases,dbreplica,genstats,locks,memmgr,sqlstats,sqlerrors,transactions,waitstats"
	require.True(t, *total3["mssql"].Settings.(*collector.MSSqlSettings).ClassesEnabled == total3mssql)
}

func testConfig(t *testing.T, cfg string) (map[string]collector.Collector, map[string]*collector.Initializer) {
	c := DefaultConfig
	err := yaml.Unmarshal([]byte(cfg), &c)
	require.NoError(t, err)
	collectors := collector.CreateInitializers()
	windowsExporter := kingpin.New("", "")
	// We only need this to fill in the appropriate settings structs so we can override them.
	collector.RegisterCollectorsFlags(collectors, windowsExporter)
	// Override the settings structs with our own
	err = c.toExporterConfig(collectors)
	require.NoError(t, err)
	// Register the performance monitors
	collector.RegisterCollectors(collectors)
	// Filter down to the enabled collectors
	enabledCollectorNames := enabledCollectors(c.EnabledCollectors)
	// Finally build the collectors that we need to run.
	builtCollectors, err := buildCollectors(collectors, enabledCollectorNames)
	require.NoError(t, err)
	return builtCollectors, collectors
}
