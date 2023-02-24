package mysql

import (
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/prometheus/exporter"
	"github.com/grafana/agent/pkg/flow/rivertypes"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/integrations/mysqld_exporter"
	config_util "github.com/prometheus/common/config"
)

func init() {
	component.Register(component.Registration{
		Name:    "prometheus.exporter.mysql",
		Args:    Arguments{},
		Exports: exporter.Exports{},
		Build:   exporter.New(createExporter, "mysql"),
	})
}

func createExporter(opts component.Options, args component.Arguments) (integrations.Integration, error) {
	cfg := args.(Arguments)
	return cfg.Convert().NewIntegration(opts.Logger)
}

// DefaultArguments holds the default settings for the mysqld_exporter integration.
var DefaultArguments = Arguments{
	LockWaitTimeout: 2,

	InfoSchemaProcessListProcessesByUser: true,
	InfoSchemaProcessListProcessesByHost: true,
	InfoSchemaTablesDatabases:            "*",

	PerfSchemaEventsStatementsLimit:     250,
	PerfSchemaEventsStatementsTimeLimit: 86400,
	PerfSchemaEventsStatementsTextLimit: 120,
	PerfSchemaFileInstancesFilter:       ".*",
	PerfSchemaFileInstancesRemovePrefix: "/var/lib/mysql",

	HeartbeatDatabase: "heartbeat",
	HeartbeatTable:    "heartbeat",
}

// Arguments controls the mysql component.
type Arguments struct {
	// DataSourceName to use to connect to MySQL.
	DataSourceName rivertypes.Secret `river:"data_source_name,attr,optional"`

	// Collectors to mark as enabled in addition to the default.
	EnableCollectors []string `river:"enable_collectors,attr,optional"`
	// Collectors to explicitly mark as disabled.
	DisableCollectors []string `river:"disable_collectors,attr,optional"`

	// Overrides the default set of enabled collectors with the given list.
	SetCollectors []string `river:"set_collectors,attr,optional"`

	// Collector-wide options
	LockWaitTimeout int  `river:"lock_wait_timeout,attr,optional"`
	LogSlowFilter   bool `river:"log_slow_filter,attr,optional"`

	// Collector-specific config options
	InfoSchemaProcessListMinTime         int    `river:"info_schema_processlist_min_time,attr,optional"`
	InfoSchemaProcessListProcessesByUser bool   `river:"info_schema_processlist_processes_by_user,attr,optional"`
	InfoSchemaProcessListProcessesByHost bool   `river:"info_schema_processlist_processes_by_host,attr,optional"`
	InfoSchemaTablesDatabases            string `river:"info_schema_tables_databases,attr,optional"`
	PerfSchemaEventsStatementsLimit      int    `river:"perf_schema_eventsstatements_limit,attr,optional"`
	PerfSchemaEventsStatementsTimeLimit  int    `river:"perf_schema_eventsstatements_time_limit,attr,optional"`
	PerfSchemaEventsStatementsTextLimit  int    `river:"perf_schema_eventsstatements_digtext_text_limit,attr,optional"`
	PerfSchemaFileInstancesFilter        string `river:"perf_schema_file_instances_filter,attr,optional"`
	PerfSchemaFileInstancesRemovePrefix  string `river:"perf_schema_file_instances_remove_prefix,attr,optional"`
	HeartbeatDatabase                    string `river:"heartbeat_database,attr,optional"`
	HeartbeatTable                       string `river:"heartbeat_table,attr,optional"`
	HeartbeatUTC                         bool   `river:"heartbeat_utc,attr,optional"`
	MySQLUserPrivileges                  bool   `river:"mysql_user_privileges,attr,optional"`
}

// UnmarshalRiver implements River unmarshalling for Config.
func (c *Arguments) UnmarshalRiver(f func(interface{}) error) error {
	*c = DefaultArguments

	type cfg Arguments
	return f((*cfg)(c))
}

func (a *Arguments) Convert() *mysqld_exporter.Config {
	return &mysqld_exporter.Config{
		DataSourceName:                       config_util.Secret(a.DataSourceName),
		EnableCollectors:                     a.EnableCollectors,
		DisableCollectors:                    a.DisableCollectors,
		SetCollectors:                        a.SetCollectors,
		LockWaitTimeout:                      a.LockWaitTimeout,
		LogSlowFilter:                        a.LogSlowFilter,
		InfoSchemaProcessListMinTime:         a.InfoSchemaProcessListMinTime,
		InfoSchemaProcessListProcessesByUser: a.InfoSchemaProcessListProcessesByUser,
		InfoSchemaProcessListProcessesByHost: a.InfoSchemaProcessListProcessesByHost,
		InfoSchemaTablesDatabases:            a.InfoSchemaTablesDatabases,
		PerfSchemaEventsStatementsLimit:      a.PerfSchemaEventsStatementsLimit,
		PerfSchemaEventsStatementsTimeLimit:  a.PerfSchemaEventsStatementsTimeLimit,
		PerfSchemaEventsStatementsTextLimit:  a.PerfSchemaEventsStatementsTextLimit,
		PerfSchemaFileInstancesFilter:        a.PerfSchemaFileInstancesFilter,
		PerfSchemaFileInstancesRemovePrefix:  a.PerfSchemaFileInstancesRemovePrefix,
		HeartbeatDatabase:                    a.HeartbeatDatabase,
		HeartbeatTable:                       a.HeartbeatTable,
		HeartbeatUTC:                         a.HeartbeatUTC,
		MySQLUserPrivileges:                  a.MySQLUserPrivileges,
	}
}
