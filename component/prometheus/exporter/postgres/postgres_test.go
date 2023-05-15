package postgres

import (
	"testing"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/pkg/integrations/postgres_exporter"
	"github.com/grafana/agent/pkg/river"
	"github.com/grafana/agent/pkg/river/rivertypes"
	config_util "github.com/prometheus/common/config"
	"github.com/stretchr/testify/require"
)

func TestRiverConfigUnmarshal(t *testing.T) {
	var exampleRiverConfig = `
	data_source_names = ["postgresql://username:password@localhost:5432/database?sslmode=disable"]
	disable_settings_metrics = true
	disable_default_metrics = true
	custom_queries_config_path = "/tmp/queries.yaml"
	
	autodiscovery {
		enabled = false
		database_allowlist = ["include1"]
		database_denylist  = ["exclude1", "exclude2"]
	}`

	var args Arguments
	err := river.Unmarshal([]byte(exampleRiverConfig), &args)
	require.NoError(t, err)

	expected := Arguments{
		DataSourceNames:        []rivertypes.Secret{rivertypes.Secret("postgresql://username:password@localhost:5432/database?sslmode=disable")},
		DisableSettingsMetrics: true,
		AutoDiscovery: AutoDiscovery{
			Enabled:           false,
			DatabaseDenylist:  []string{"exclude1", "exclude2"},
			DatabaseAllowlist: []string{"include1"},
		},
		DisableDefaultMetrics:   true,
		CustomQueriesConfigPath: "/tmp/queries.yaml",
	}

	require.Equal(t, expected, args)
}

func TestRiverConfigConvert(t *testing.T) {
	var exampleRiverConfig = `
	data_source_names = ["postgresql://username:password@localhost:5432/database?sslmode=disable"]
	disable_settings_metrics = true
	disable_default_metrics = true
	custom_queries_config_path = "/tmp/queries.yaml"
	
	autodiscovery {
		enabled = false
		database_allowlist = ["include1"]
		database_denylist  = ["exclude1", "exclude2"]
	}`

	var args Arguments
	err := river.Unmarshal([]byte(exampleRiverConfig), &args)
	require.NoError(t, err)

	c := args.Convert()

	expected := postgres_exporter.Config{
		DataSourceNames:        []config_util.Secret{config_util.Secret("postgresql://username:password@localhost:5432/database?sslmode=disable")},
		DisableSettingsMetrics: true,
		AutodiscoverDatabases:  false,
		ExcludeDatabases:       []string{"exclude1", "exclude2"},
		IncludeDatabases:       []string{"include1"},
		DisableDefaultMetrics:  true,
		QueryPath:              "/tmp/queries.yaml",
	}
	require.Equal(t, expected, *c)
}

func TestParsePostgresURL(t *testing.T) {
	dsn := "postgresql://linus:42secret@localhost:5432/postgres?sslmode=disable"
	expected := map[string]string{
		"dbname":   "postgres",
		"host":     "localhost",
		"password": "42secret",
		"port":     "5432",
		"sslmode":  "disable",
		"user":     "linus",
	}

	actual, err := parsePostgresURL(dsn)
	require.NoError(t, err)
	require.Equal(t, actual, expected)
}

func TestCustomizeTargetValid(t *testing.T) {
	args := Arguments{
		DataSourceNames: []rivertypes.Secret{rivertypes.Secret("postgresql://username:password@localhost:5432/database?sslmode=disable")},
	}

	baseTarget := discovery.Target{}
	newTargets := customizeTarget(baseTarget, args)
	require.Equal(t, 1, len(newTargets))
	require.Equal(t, "postgresql://localhost:5432/database", newTargets[0]["instance"])
}

func TestCustomizeTargetInvalid(t *testing.T) {
	args := Arguments{
		DataSourceNames: []rivertypes.Secret{rivertypes.Secret("invalid_ds@localhost:5432/database?sslmode=disable")},
	}

	baseTarget := discovery.Target{
		"instance": "default value",
	}
	newTargets := customizeTarget(baseTarget, args)
	require.Equal(t, 1, len(newTargets))
	require.Equal(t, "default value", newTargets[0]["instance"])
}
