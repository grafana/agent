---
aliases:
- ../../../configuration/integrations/mysqld-exporter-config/
- /docs/grafana-cloud/monitor-infrastructure/agent/static/configuration/integrations/mysqld-exporter-config/
- /docs/grafana-cloud/send-data/agent/static/configuration/integrations/mysqld-exporter-config/
canonical: https://grafana.com/docs/agent/latest/static/configuration/integrations/mysqld-exporter-config/
description: Learn about mysqld_exporter_config
title: mysqld_exporter_config
---

# mysqld_exporter_config

The `mysqld_exporter_config` block configures the `mysqld_exporter` integration,
which is an embedded version of
[`mysqld_exporter`](https://github.com/prometheus/mysqld_exporter)
and allows for collection metrics from MySQL servers.

Note that currently, an Agent can only collect metrics from a single MySQL
server. If you want to collect metrics from multiple servers, run multiple
Agents and add labels using `relabel_configs` to differentiate between the MySQL
servers:

```yaml
mysqld_exporter:
  enabled: true
  data_source_name: root@(server-a:3306)/
  relabel_configs:
  - source_labels: [__address__]
    target_label: instance
    replacement: server-a
```

We strongly recommend that you configure a separate user for the Agent, and give it only the strictly mandatory
security privileges necessary for monitoring your node, as per the [official documentation](https://github.com/prometheus/mysqld_exporter#required-grants).

Full reference of options:

```yaml
  # Enables the mysqld_exporter integration, allowing the Agent to collect
  # metrics from a MySQL server.
  [enabled: <boolean> | default = false]

  # Sets an explicit value for the instance label when the integration is
  # self-scraped. Overrides inferred values.
  #
  # The default value for this integration is a truncated version of the
  # connection DSN, containing only the server and db name. (Credentials
  # are not included.)
  [instance: <string>]

  # Automatically collect metrics from this integration. If disabled,
  # the mysqld_exporter integration will be run but not scraped and thus not
  # remote-written. Metrics for the integration will be exposed at
  # /integrations/mysqld_exporter/metrics and can be scraped by an external
  # process.
  [scrape_integration: <boolean> | default = <integrations_config.scrape_integrations>]

  # How often should the metrics be collected? Defaults to
  # prometheus.global.scrape_interval.
  [scrape_interval: <duration> | default = <global_config.scrape_interval>]

  # The timeout before considering the scrape a failure. Defaults to
  # prometheus.global.scrape_timeout.
  [scrape_timeout: <duration> | default = <global_config.scrape_timeout>]

  # Allows for relabeling labels on the target.
  relabel_configs:
    [- <relabel_config> ... ]

  # Relabel metrics coming from the integration, allowing to drop series
  # from the integration that you don't care about.
  metric_relabel_configs:
    [ - <relabel_config> ... ]

  # How frequent to truncate the WAL for this integration.
  [wal_truncate_frequency: <duration> | default = "60m"]

  # Data Source Name specifies the MySQL server to connect to. This is REQUIRED
  # but may also be specified by the MYSQLD_EXPORTER_DATA_SOURCE_NAME
  # environment variable. If neither are set, the integration will fail to
  # start.
  #
  # The format of this is specified here: https://github.com/go-sql-driver/mysql#dsn-data-source-name
  #
  # A working example value for a server with no required password
  # authentication is: "root@(localhost:3306)/"
  data_source_name: <string>

  # A list of collector names to enable on top of the default set.
  enable_collectors:
    [ - <string> ]
  # A list of collector names to disable from the default set.
  disable_collectors:
    [ - <string> ]
  # A list of collectors to run. Fully overrides the default set.
  set_collectors:
    [ - <string> ]

  # Set a lock_wait_timeout on the connection to avoid long metadata locking.
  [lock_wait_timeout: <int> | default = 2]
  # Add a low_slow_filter to avoid slow query logging of scrapes. NOT supported
  # by Oracle MySQL.
  [log_slow_filter: <bool> | default = false]

  ## Collector-specific options

  # Minimum time a thread must be in each state to be counted.
  [info_schema_processlist_min_time: <int> | default = 0]
  # Enable collecting the number of processes by user.
  [info_schema_processlist_processes_by_user: <bool> | default = true]
  # Enable collecting the number of processes by host.
  [info_schema_processlist_processes_by_host: <bool> | default = true]
  # The list of databases to collect table stats for. * for all
  [info_schema_tables_databases: <string> | default = "*"]
  # Limit the number of events statements digests by response time.
  [perf_schema_eventsstatements_limit: <int> | default = 250]
  # Limit how old the 'last_seen' events statements can be, in seconds.
  [perf_schema_eventsstatements_time_limit: <int> | default = 86400]
  # Maximum length of the normalized statement text.
  [perf_schema_eventsstatements_digtext_text_limit: <int> | default = 120]
  # Regex file_name filter for performance_schema.file_summary_by_instance
  [perf_schema_file_instances_filter: <string> | default = ".*"]
  # Remove path prefix in performance_schema.file_summary_by_instance
  [perf_schema_file_instances_remove_prefix: <string> | default = "/var/lib/mysql"]
  # Remove instrument prefix in performance_schema.memory_summary_global_by_event_name
  [perf_schema_memory_events_remove_prefix: <string> | default = "memory/"]
  # Database from where to collect heartbeat data.
  [heartbeat_database: <string> | default = "heartbeat"]
  # Table from where to collect heartbeat data.
  [heartbeat_table: <string> | default = "heartbeat"]
  # Use UTC for timestamps of the current server (`pt-heartbeat` is called with `--utc`)
  [heartbeat_utc: <bool> | default = false]
  # Enable collecting user privileges from mysql.user
  [mysql_user_privileges: <bool> | default = false]
```

The full list of collectors that are supported for `mysqld_exporter` is:

| Name                                             | Description | Enabled by default |
| ------------------------------------------------ | ----------- | ------------------ |
| auto_increment.columns                           | Collect auto_increment columns and max values from information_schema | no |
| binlog_size                                      | Collect the current size of all registered binlog files | no |
| engine_innodb_status                             | Collect from SHOW ENGINE INNODB STATUS | no |
| engine_tokudb_status                             | Collect from SHOW ENGINE TOKUDB STATUS | no |
| global_status                                    | Collect from SHOW GLOBAL STATUS | yes |
| global_variables                                 | Collect from SHOW GLOBAL VARIABLES | yes |
| heartbeat                                        | Collect from heartbeat | no |
| info_schema.clientstats                          | If running with userstat=1, enable to collect client statistics | no |
| info_schema.innodb_cmpmem                        | Collect metrics from information_schema.innodb_cmpmem | yes |
| info_schema.innodb_metrics                       | Collect metrics from information_schema.innodb_metrics | yes |
| info_schema.innodb_tablespaces                   | Collect metrics from information_schema.innodb_sys_tablespaces | no |
| info_schema.processlist                          | Collect current thread state counts from the information_schema.processlist | no |
| info_schema.query_response_time                  | Collect query response time distribution if query_response_time_stats is ON | yes |
| info_schema.replica_host                         | Collect metrics from information_schema.replica_host_status | no |
| info_schema.schemastats                          | If running with userstat=1, enable to collect schema statistics | no |
| info_schema.tables                               | Collect metrics from information_schema.tables | no |
| info_schema.tablestats                           | If running with userstat=1, enable to collect table statistics | no |
| info_schema.userstats                            | If running with userstat=1, enable to collect user statistics | no |
| mysql.user                                       | Collect data from mysql.user | no |
| perf_schema.eventsstatements                     | Collect metrics from performance_schema.events_statements_summary_by_digest | no |
| perf_schema.eventsstatementssum                  | Collect metrics of grand sums from performance_schema.events_statements_summary_by_digest | no |
| perf_schema.eventswaits                          | Collect metrics from performance_schema.events_waits_summary_global_by_event_name | no |
| perf_schema.file_events                          | Collect metrics from performance_schema.file_summary_by_event_name | no |
| perf_schema.file_instances                       | Collect metrics from performance_schema.file_summary_by_instance | no |
| perf_schema.indexiowaits                         | Collect metrics from performance_schema.table_io_waits_summary_by_index_usage | no |
| perf_schema.memory_events                        | Collect metrics from performance_schema.memory_summary_global_by_event_name |no |
| perf_schema.replication_applier_status_by_worker | Collect metrics from performance_schema.replication_applier_status_by_worker | no |
| perf_schema.replication_group_member_stats       | Collect metrics from performance_schema.replication_group_member_stats | no |
| perf_schema.replication_group_members            | Collect metrics from performance_schema.replication_group_members | no |
| perf_schema.tableiowaits                         | Collect metrics from performance_schema.table_io_waits_summary_by_table | no |
| perf_schema.tablelocks                           | Collect metrics from performance_schema.table_lock_waits_summary_by_table | no |
| slave_hosts                                      | Scrape information from 'SHOW SLAVE HOSTS' | no |
| slave_status                                     | Scrape information from SHOW SLAVE STATUS | yes |
