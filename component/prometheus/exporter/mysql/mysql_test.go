package mysql

import (
	"testing"

	"github.com/grafana/agent/pkg/integrations/mysqld_exporter"
	"github.com/grafana/agent/pkg/river"
	"github.com/stretchr/testify/require"
)

func TestRiverConfigUnmarshal(t *testing.T) {
	var exampleRiverConfig = `
	data_source_name = "DataSourceName"
	enable_collectors = ["collector1"]
	disable_collectors = ["collector2"]
	set_collectors = ["collector3", "collector4"]
	lock_wait_timeout = 1
	log_slow_filter = false
	info_schema_processlist_min_time = 2
	info_schema_processlist_processes_by_user = true
	info_schema_processlist_processes_by_host = false
	info_schema_tables_databases = "schema"
	perf_schema_eventsstatements_limit = 3
	perf_schema_eventsstatements_time_limit = 4
	perf_schema_eventsstatements_digtext_text_limit = 5
	perf_schema_file_instances_filter = "instances_filter"
	perf_schema_file_instances_remove_prefix = "instances_remove"
	heartbeat_database = "heartbeat_database"
	heartbeat_table = "heartbeat_table"
	heartbeat_utc = true
	mysql_user_privileges = false
`

	var args Arguments
	err := river.Unmarshal([]byte(exampleRiverConfig), &args)
	require.NoError(t, err)

	require.Equal(t, "DataSourceName", string(args.DataSourceName))
	require.Equal(t, []string{"collector1"}, args.EnableCollectors)
	require.Equal(t, []string{"collector2"}, args.DisableCollectors)
	require.Equal(t, []string{"collector3", "collector4"}, args.SetCollectors)
	require.Equal(t, 1, args.LockWaitTimeout)
	require.False(t, args.LogSlowFilter)
	require.Equal(t, 2, args.InfoSchemaProcessListMinTime)
	require.True(t, args.InfoSchemaProcessListProcessesByUser)
	require.False(t, args.InfoSchemaProcessListProcessesByHost)
	require.Equal(t, "schema", args.InfoSchemaTablesDatabases)
	require.Equal(t, 3, args.PerfSchemaEventsStatementsLimit)
	require.Equal(t, 4, args.PerfSchemaEventsStatementsTimeLimit)
	require.Equal(t, 5, args.PerfSchemaEventsStatementsTextLimit)
	require.Equal(t, "instances_filter", args.PerfSchemaFileInstancesFilter)
	require.Equal(t, "instances_remove", args.PerfSchemaFileInstancesRemovePrefix)
	require.Equal(t, "heartbeat_database", args.HeartbeatDatabase)
	require.Equal(t, "heartbeat_table", args.HeartbeatTable)
	require.True(t, args.HeartbeatUTC)
	require.False(t, args.MySQLUserPrivileges)
}

func TestRiverConfigConvert(t *testing.T) {
	var exampleRiverConfig = `
	data_source_name = "DataSourceName"
	enable_collectors = ["collector1"]
	disable_collectors = ["collector2"]
	set_collectors = ["collector3", "collector4"]
	lock_wait_timeout = 1
	log_slow_filter = false
	info_schema_processlist_min_time = 2
	info_schema_processlist_processes_by_user = true
	info_schema_processlist_processes_by_host = false
	info_schema_tables_databases = "schema"
	perf_schema_eventsstatements_limit = 3
	perf_schema_eventsstatements_time_limit = 4
	perf_schema_eventsstatements_digtext_text_limit = 5
	perf_schema_file_instances_filter = "instances_filter"
	perf_schema_file_instances_remove_prefix = "instances_remove"
	heartbeat_database = "heartbeat_database"
	heartbeat_table = "heartbeat_table"
	heartbeat_utc = true
	mysql_user_privileges = false
`

	var args Arguments
	err := river.Unmarshal([]byte(exampleRiverConfig), &args)
	require.NoError(t, err)

	c := args.Convert()
	require.Equal(t, "DataSourceName", string(c.DataSourceName))
	require.Equal(t, []string{"collector1"}, c.EnableCollectors)
	require.Equal(t, []string{"collector2"}, c.DisableCollectors)
	require.Equal(t, []string{"collector3", "collector4"}, c.SetCollectors)
	require.Equal(t, 1, c.LockWaitTimeout)
	require.False(t, c.LogSlowFilter)
	require.Equal(t, 2, c.InfoSchemaProcessListMinTime)
	require.True(t, c.InfoSchemaProcessListProcessesByUser)
	require.False(t, c.InfoSchemaProcessListProcessesByHost)
	require.Equal(t, "schema", c.InfoSchemaTablesDatabases)
	require.Equal(t, 3, c.PerfSchemaEventsStatementsLimit)
	require.Equal(t, 4, c.PerfSchemaEventsStatementsTimeLimit)
	require.Equal(t, 5, c.PerfSchemaEventsStatementsTextLimit)
	require.Equal(t, "instances_filter", c.PerfSchemaFileInstancesFilter)
	require.Equal(t, "instances_remove", c.PerfSchemaFileInstancesRemovePrefix)
	require.Equal(t, "heartbeat_database", c.HeartbeatDatabase)
	require.Equal(t, "heartbeat_table", c.HeartbeatTable)
	require.True(t, c.HeartbeatUTC)
	require.False(t, c.MySQLUserPrivileges)
}

// Checks that the flow and static default configs have not drifted
func TestDefaultsSame(t *testing.T) {
	convertedDefaults := DefaultArguments.Convert()
	require.Equal(t, mysqld_exporter.DefaultConfig, *convertedDefaults)
}
