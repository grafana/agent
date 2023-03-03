package postgres

import (
	"testing"

	"github.com/grafana/agent/pkg/flow/rivertypes"
	"github.com/grafana/agent/pkg/river"
	config_util "github.com/prometheus/common/config"
	"github.com/stretchr/testify/require"
)

func TestRiverConfigUnmarshal(t *testing.T) {
	var exampleRiverConfig = `
	data_source_names = ["postgresql://username:password@localhost:5432/database?sslmode=disable"]

	disable_settings_metrics = true
	autodiscover_databases = false
	exclude_databases = ["exclude1", "exclude2"]
	include_databases = ["include1"]
	disable_default_metrics = true
	query_path = "/tmp/queries.yaml"
`

	var args Arguments
	err := river.Unmarshal([]byte(exampleRiverConfig), &args)
	require.NoError(t, err)

	require.Equal(t, []rivertypes.Secret{rivertypes.Secret("postgresql://username:password@localhost:5432/database?sslmode=disable")}, args.DataSourceNames)
	require.True(t, args.DisableSettingsMetrics)
	require.False(t, args.AutodiscoverDatabases)
	require.Equal(t, []string{"exclude1", "exclude2"}, args.ExcludeDatabases)
	require.Equal(t, []string{"include1"}, args.IncludeDatabases)
	require.True(t, args.DisableDefaultMetrics)
	require.Equal(t, "/tmp/queries.yaml", args.QueryPath)
}

func TestRiverConfigConvert(t *testing.T) {
	var exampleRiverConfig = `
	data_source_names = ["postgresql://username:password@localhost:5432/database?sslmode=disable"]

	disable_settings_metrics = true
	autodiscover_databases = false
	exclude_databases = ["exclude1", "exclude2"]
	include_databases = ["include1"]
	disable_default_metrics = true
	query_path = "/tmp/queries.yaml"
`

	var args Arguments
	err := river.Unmarshal([]byte(exampleRiverConfig), &args)
	require.NoError(t, err)

	c := args.Convert()
	require.Equal(t, []config_util.Secret{config_util.Secret("postgresql://username:password@localhost:5432/database?sslmode=disable")}, c.DataSourceNames)
	require.True(t, c.DisableSettingsMetrics)
	require.False(t, c.AutodiscoverDatabases)
	require.Equal(t, []string{"exclude1", "exclude2"}, c.ExcludeDatabases)
	require.Equal(t, []string{"include1"}, c.IncludeDatabases)
	require.True(t, c.DisableDefaultMetrics)
	require.Equal(t, "/tmp/queries.yaml", c.QueryPath)
}
