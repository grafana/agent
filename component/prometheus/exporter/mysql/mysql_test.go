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

	info_schema.processlist {
		min_time = 2
		processes_by_user = true
		processes_by_host = false
	}

	info_schema.tables {
		databases = "schema"
	}

	perf_schema.eventsstatements {
		limit = 3
		time_limit = 4
		text_limit = 5
	}

	perf_schema.file_instances {
		filter = "instances_filter"
		remove_prefix = "instances_remove"
	}

	heartbeat {
		database = "heartbeat_database"
		table = "heartbeat_table"
		utc = true
	}

	mysql.user {
		privileges = false
	}
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
	require.Equal(t, 2, args.InfoSchemaProcessList.MinTime)
	require.True(t, args.InfoSchemaProcessList.ProcessesByUser)
	require.False(t, args.InfoSchemaProcessList.ProcessesByHost)
	require.Equal(t, "schema", args.InfoSchemaTables.Databases)
	require.Equal(t, 3, args.PerfSchemaEventsStatements.Limit)
	require.Equal(t, 4, args.PerfSchemaEventsStatements.TimeLimit)
	require.Equal(t, 5, args.PerfSchemaEventsStatements.TextLimit)
	require.Equal(t, "instances_filter", args.PerfSchemaFileInstances.Filter)
	require.Equal(t, "instances_remove", args.PerfSchemaFileInstances.RemovePrefix)
	require.Equal(t, "heartbeat_database", args.Heartbeat.Database)
	require.Equal(t, "heartbeat_table", args.Heartbeat.Table)
	require.True(t, args.Heartbeat.UTC)
	require.False(t, args.MySQLUser.Privileges)
}

func TestRiverConfigConvert(t *testing.T) {
	var exampleRiverConfig = `
	data_source_name = "DataSourceName"
	enable_collectors = ["collector1"]
	disable_collectors = ["collector2"]
	set_collectors = ["collector3", "collector4"]
	lock_wait_timeout = 1
	log_slow_filter = false
	
	info_schema.processlist {
		min_time = 2
		processes_by_user = true
		processes_by_host = false
	}

	info_schema.tables {
		databases = "schema"
	}

	perf_schema.eventsstatements {
		limit = 3
		time_limit = 4
		text_limit = 5
	}

	perf_schema.file_instances {
		filter = "instances_filter"
		remove_prefix = "instances_remove"
	}

	heartbeat {
		database = "heartbeat_database"
		table = "heartbeat_table"
		utc = true
	}

	mysql.user {
		privileges = false
	}
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
