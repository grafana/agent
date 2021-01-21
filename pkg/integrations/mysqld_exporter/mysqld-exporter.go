// Package mysqld_exporter embeds https://github.com/prometheus/mysqld_exporter
package mysqld_exporter //nolint:golint

import (
	"context"
	"fmt"
	"os"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/integrations/config"
	"github.com/prometheus/mysqld_exporter/collector"
)

var DefaultConfig = Config{
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

// Config controls the mysqld_exporter integration.
type Config struct {
	Common config.Common `yaml:",inline"`

	// DataSourceName to use to connect to MySQL.
	DataSourceName string `yaml:"data_source_name"`

	// Collectors to mark as enabled in addition to the default.
	EnableCollectors []string `yaml:"enable_collectors"`
	// Collectors to explicitly mark as disabled.
	DisableCollectors []string `yaml:"disable_collectors"`

	// Overrides the default set of enabled collectors with the given list.
	SetCollectors []string `yaml:"set_collectors"`

	// Collector-wide options
	LockWaitTimeout int  `yaml:"lock_wait_timeout"`
	LogSlowFilter   bool `yaml:"log_slow_filter"`

	// Collector-specific config options
	InfoSchemaProcessListMinTime         int    `yaml:"info_schema_processlist_min_time"`
	InfoSchemaProcessListProcessesByUser bool   `yaml:"info_schema_processlist_processes_by_user"`
	InfoSchemaProcessListProcessesByHost bool   `yaml:"info_schema_processlist_processes_by_host"`
	InfoSchemaTablesDatabases            string `yaml:"info_schema_tables_databases"`
	PerfSchemaEventsStatementsLimit      int    `yaml:"perf_schema_eventsstatements_limit"`
	PerfSchemaEventsStatementsTimeLimit  int    `yaml:"perf_schema_eventsstatements_time_limit"`
	PerfSchemaEventsStatementsTextLimit  int    `yaml:"perf_schema_eventsstatements_digtext_text_limit"`
	PerfSchemaFileInstancesFilter        string `yaml:"perf_schema_file_instances_filter"`
	PerfSchemaFileInstancesRemovePrefix  string `yaml:"perf_schema_file_instances_remove_prefix"`
	HeartbeatDatabase                    string `yaml:"heartbeat_database"`
	HeartbeatTable                       string `yaml:"heartbeat_table"`
	HeartbeatUTC                         bool   `yaml:"heartbeat_utc"`
	MySQLUserPrivileges                  bool   `yaml:"mysql_user_privileges"`
}

// UnmarshalYAML implements yaml.Unmarshaler for Config.
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultConfig

	type plain Config
	return unmarshal((*plain)(c))
}

func (c *Config) Name() string {
	return "mysqld_exporter"
}

func (c *Config) CommonConfig() config.Common {
	return c.Common
}

func (c *Config) NewIntegration(l log.Logger) (integrations.Integration, error) {
	return New(l, c)
}

func init() {
	integrations.RegisterIntegration(&Config{})
}

// New creates a new mysqld_exporter integration. The integration scrapes
// metrics from a mysqld process.
func New(log log.Logger, c *Config) (integrations.Integration, error) {
	dsn := c.DataSourceName
	if len(dsn) == 0 {
		dsn = os.Getenv("MYSQLD_EXPORTER_DATA_SOURCE_NAME")
	}
	if len(dsn) == 0 {
		return nil, fmt.Errorf("cannot create mysqld_exporter; neither mysqld_exporter.data_source_name or $MYSQLD_EXPORTER_DATA_SOURCE_NAME is set")
	}

	scrapers := GetScrapers(c)
	exporter := collector.New(context.Background(), dsn, collector.NewMetrics(), scrapers, log, collector.Config{
		LockTimeout:   c.LockWaitTimeout,
		SlowLogFilter: c.LogSlowFilter,
	})

	level.Debug(log).Log("msg", "enabled mysqld_exporter scrapers")
	for _, scraper := range scrapers {
		level.Debug(log).Log("scraper", scraper.Name())
	}

	return integrations.NewCollectorIntegration(
		c.Name(),
		integrations.WithCollectors(exporter),
	), nil
}

// GetScrapers returns the set of *enabled* scrapers from the config.
// Configurable scrapers will have their configuration filled out matching the
// Config's settings.
func GetScrapers(c *Config) []collector.Scraper {
	scrapers := map[collector.Scraper]bool{
		&collector.ScrapeAutoIncrementColumns{}:                false,
		&collector.ScrapeBinlogSize{}:                          false,
		&collector.ScrapeClientStat{}:                          false,
		&collector.ScrapeEngineInnodbStatus{}:                  false,
		&collector.ScrapeEngineTokudbStatus{}:                  false,
		&collector.ScrapeGlobalStatus{}:                        true,
		&collector.ScrapeGlobalVariables{}:                     true,
		&collector.ScrapeInfoSchemaInnodbTablespaces{}:         false,
		&collector.ScrapeInnodbCmpMem{}:                        true,
		&collector.ScrapeInnodbCmp{}:                           true,
		&collector.ScrapeInnodbMetrics{}:                       false,
		&collector.ScrapePerfEventsStatementsSum{}:             false,
		&collector.ScrapePerfEventsWaits{}:                     false,
		&collector.ScrapePerfFileEvents{}:                      false,
		&collector.ScrapePerfIndexIOWaits{}:                    false,
		&collector.ScrapePerfReplicationApplierStatsByWorker{}: false,
		&collector.ScrapePerfReplicationGroupMemberStats{}:     false,
		&collector.ScrapePerfReplicationGroupMembers{}:         false,
		&collector.ScrapePerfTableIOWaits{}:                    false,
		&collector.ScrapePerfTableLockWaits{}:                  false,
		&collector.ScrapeQueryResponseTime{}:                   true,
		&collector.ScrapeReplicaHost{}:                         false,
		&collector.ScrapeSchemaStat{}:                          false,
		&collector.ScrapeSlaveHosts{}:                          false,
		&collector.ScrapeSlaveStatus{}:                         true,
		&collector.ScrapeTableStat{}:                           false,
		&collector.ScrapeUserStat{}:                            false,

		// Collectors that have configuration
		&collector.ScrapeHeartbeat{
			Database: c.HeartbeatDatabase,
			Table:    c.HeartbeatTable,
			UTC:      c.HeartbeatUTC,
		}: false,

		&collector.ScrapePerfEventsStatements{
			Limit:           c.PerfSchemaEventsStatementsLimit,
			TimeLimit:       c.PerfSchemaEventsStatementsTimeLimit,
			DigestTextLimit: c.PerfSchemaEventsStatementsTextLimit,
		}: false,

		&collector.ScrapePerfFileInstances{
			Filter:       c.PerfSchemaFileInstancesFilter,
			RemovePrefix: c.PerfSchemaFileInstancesRemovePrefix,
		}: false,

		&collector.ScrapeProcesslist{
			ProcessListMinTime:  c.InfoSchemaProcessListMinTime,
			ProcessesByHostFlag: c.InfoSchemaProcessListProcessesByHost,
			ProcessesByUserFlag: c.InfoSchemaProcessListProcessesByUser,
		}: false,

		&collector.ScrapeTableSchema{
			Databases: c.InfoSchemaTablesDatabases,
		}: false,

		&collector.ScrapeUser{
			Privileges: c.MySQLUserPrivileges,
		}: false,
	}

	// Override the defaults with the provided set of collectors if
	// set_collectors has at least one element in it.
	if len(c.SetCollectors) != 0 {
		customDefaults := map[string]struct{}{}
		for _, c := range c.SetCollectors {
			customDefaults[c] = struct{}{}
		}
		for scraper := range scrapers {
			_, enable := customDefaults[scraper.Name()]
			scrapers[scraper] = enable
		}
	}

	// Explicitly disable/enable specific collectors.
	for _, c := range c.DisableCollectors {
		for scraper := range scrapers {
			if scraper.Name() == c {
				scrapers[scraper] = false
				break
			}
		}
	}
	for _, c := range c.EnableCollectors {
		for scraper := range scrapers {
			if scraper.Name() == c {
				scrapers[scraper] = true
				break
			}
		}
	}

	enabledScrapers := []collector.Scraper{}
	for scraper, enabled := range scrapers {
		if enabled {
			enabledScrapers = append(enabledScrapers, scraper)
		}
	}
	return enabledScrapers
}
