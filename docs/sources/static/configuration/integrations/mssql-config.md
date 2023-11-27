---
aliases:
- ../../../configuration/integrations/mssql-config/
- /docs/grafana-cloud/monitor-infrastructure/agent/static/configuration/integrations/mssql-config/
- /docs/grafana-cloud/send-data/agent/static/configuration/integrations/mssql-config/
canonical: https://grafana.com/docs/agent/latest/static/configuration/integrations/mssql-config/
description: Learn about mssql_config
title: mssql_config
---

# mssql_config

The `mssql_config` block configures the `mssql` integration, an embedded version of [`sql_exporter`](https://github.com/burningalchemist/sql_exporter) that lets you collect [Microsoft SQL Server](https://www.microsoft.com/en-us/sql-server) metrics.

It is recommended that you have a dedicated user set up for monitoring an mssql instance.
The user for monitoring must have the following grants in order to populate the metrics:
```
GRANT VIEW ANY DEFINITION TO <MONITOR_USER>
GRANT VIEW SERVER STATE TO <MONITOR_USER>
```

## Quick configuration example

To get started, define the MSSQL connection string in Grafana Agent's integration block:

```yaml
metrics:
  wal_directory: /tmp/wal
integrations:
  mssql:
    enabled: true
    connection_string: "sqlserver://[user]:[pass]@localhost:1433"
```

Full reference of options:

```yaml
  # Enables the MSSQL integration, allowing the Agent to automatically
  # collect metrics for the specified MSSQL instance.
  [enabled: <boolean> | default = false]

  # Sets an explicit value for the instance label when the integration is
  # self-scraped. Overrides inferred values.
  #
  # The default value for this integration is the host:port of the provided connection_string.
  [instance: <string>]

  # Automatically collect metrics from this integration. If disabled,
  # the MSSQL integration is run but not scraped and thus not
  # remote-written. Metrics for the integration are exposed at
  # /integrations/mssql/metrics and can be scraped by an external
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

  # Relabel metrics coming from the integration, lets you drop series
  # that you don't care about from the integration.
  metric_relabel_configs:
    [ - <relabel_config> ... ]

  # How frequently the WAL is truncated for this integration.
  [wal_truncate_frequency: <duration> | default = "60m"]

  #
  # Exporter-specific configuration options
  #

  # The connection_string to use to connect to the MSSQL instance.
  # It is specified in the form of: "sqlserver://<USERNAME>:<PASSWORD>@<HOST>:<PORT>"
  connection_string: <string>

  # The maximum number of open database connections to the MSSQL instance.
  [max_open_connections: <int> | default = 3]

  # The maximum number of idle database connections to the MSSQL instance.
  [max_idle_connections: <int> | default = 3]

  # The timeout for scraping metrics from the MSSQL instance.
  [timeout: <duration> | default = "10s"]

  # Embedded MSSQL query configuration for specifying custom MSSQL Prometheus metrics.
  # See https://github.com/burningalchemist/sql_exporter#collectors for more details how to specify your metric configurations.
  query_config: 
    [- <metrics> ... ]
    [- <queries> ... ]]
```

## Custom metrics
You can use the optional `query_config` parameter to retrieve custom Prometheus metrics for a MSSQL instance.

If this is defined, the new configuration will be used to query your MSSQL instance and create whatever Prometheus metrics are defined.
If you want additional metrics on top of the default metrics, the default configuration must be used as a base.

The default configuration used by this integration is as follows:
```
collector_name: mssql_standard

metrics:
  - metric_name: mssql_local_time_seconds
    type: gauge
    help: 'Local time in seconds since epoch (Unix time).'
    values: [unix_time]
    query: |
      SELECT DATEDIFF(second, '19700101', GETUTCDATE()) AS unix_time
  - metric_name: mssql_connections
    type: gauge
    help: 'Number of active connections.'
    key_labels:
      - db
    values: [count]
    query: |
      SELECT DB_NAME(sp.dbid) AS db, COUNT(sp.spid) AS count
      FROM sys.sysprocesses sp
      GROUP BY DB_NAME(sp.dbid)
  #
  # Collected from sys.dm_os_performance_counters
  #
  - metric_name: mssql_deadlocks_total
    type: counter
    help: 'Number of lock requests that resulted in a deadlock.'
    values: [cntr_value]
    query: |
      SELECT cntr_value
      FROM sys.dm_os_performance_counters WITH (NOLOCK)
      WHERE counter_name = 'Number of Deadlocks/sec' AND instance_name = '_Total'
  - metric_name: mssql_user_errors_total
    type: counter
    help: 'Number of user errors.'
    values: [cntr_value]
    query: |
      SELECT cntr_value
      FROM sys.dm_os_performance_counters WITH (NOLOCK)
      WHERE counter_name = 'Errors/sec' AND instance_name = 'User Errors'
  - metric_name: mssql_kill_connection_errors_total
    type: counter
    help: 'Number of severe errors that caused SQL Server to kill the connection.'
    values: [cntr_value]
    query: |
      SELECT cntr_value
      FROM sys.dm_os_performance_counters WITH (NOLOCK)
      WHERE counter_name = 'Errors/sec' AND instance_name = 'Kill Connection Errors'
  - metric_name: mssql_page_life_expectancy_seconds
    type: gauge
    help: 'The minimum number of seconds a page will stay in the buffer pool on this node without references.'
    values: [cntr_value]
    query: |
      SELECT top(1) cntr_value
      FROM sys.dm_os_performance_counters WITH (NOLOCK)
      WHERE counter_name = 'Page life expectancy'
  - metric_name: mssql_batch_requests_total
    type: counter
    help: 'Number of command batches received.'
    values: [cntr_value]
    query: |
      SELECT cntr_value
      FROM sys.dm_os_performance_counters WITH (NOLOCK)
      WHERE counter_name = 'Batch Requests/sec'
  - metric_name: mssql_log_growths_total
    type: counter
    help: 'Number of times the transaction log has been expanded, per database.'
    key_labels:
      - db
    values: [cntr_value]
    query: |
      SELECT rtrim(instance_name) AS db, cntr_value
      FROM sys.dm_os_performance_counters WITH (NOLOCK)
      WHERE counter_name = 'Log Growths' AND instance_name <> '_Total'
  - metric_name: mssql_buffer_cache_hit_ratio
    type: gauge
    help: 'Ratio of requests that hit the buffer cache'
    values: [BufferCacheHitRatio]
    query: |
      SELECT (a.cntr_value * 1.0 / b.cntr_value) * 100.0 as BufferCacheHitRatio
      FROM sys.dm_os_performance_counters  a
      JOIN  (SELECT cntr_value, OBJECT_NAME 
          FROM sys.dm_os_performance_counters  
          WHERE counter_name = 'Buffer cache hit ratio base'
              AND OBJECT_NAME = 'SQLServer:Buffer Manager') b ON  a.OBJECT_NAME = b.OBJECT_NAME
      WHERE a.counter_name = 'Buffer cache hit ratio'
      AND a.OBJECT_NAME = 'SQLServer:Buffer Manager'

  - metric_name: mssql_checkpoint_pages_sec
    type: gauge
    help: 'Checkpoint Pages Per Second'
    values: [cntr_value]
    query: |
      SELECT cntr_value
      FROM sys.dm_os_performance_counters
      WHERE [counter_name] = 'Checkpoint pages/sec'
  #
  # Collected from sys.dm_io_virtual_file_stats
  #
  - metric_name: mssql_io_stall_seconds_total
    type: counter
    help: 'Stall time in seconds per database and I/O operation.'
    key_labels:
      - db
    value_label: operation
    values:
      - read
      - write
    query_ref: mssql_io_stall

  #
  # Collected from sys.dm_os_process_memory
  #
  - metric_name: mssql_resident_memory_bytes
    type: gauge
    help: 'SQL Server resident memory size (AKA working set).'
    values: [resident_memory_bytes]
    query_ref: mssql_process_memory

  - metric_name: mssql_virtual_memory_bytes
    type: gauge
    help: 'SQL Server committed virtual memory size.'
    values: [virtual_memory_bytes]
    query_ref: mssql_process_memory

  - metric_name: mssql_available_commit_memory_bytes
    type: gauge
    help: 'SQL Server available to be committed memory size.'
    values: [available_commit_limit_bytes]
    query_ref: mssql_process_memory

  - metric_name: mssql_memory_utilization_percentage
    type: gauge
    help: 'The percentage of committed memory that is in the working set.'
    values: [memory_utilization_percentage]
    query_ref: mssql_process_memory

  - metric_name: mssql_page_fault_count_total
    type: counter
    help: 'The number of page faults that were incurred by the SQL Server process.'
    values: [page_fault_count]
    query_ref: mssql_process_memory

  #
  # Collected from sys.dm_os_sys_info
  #
  - metric_name: mssql_server_total_memory_bytes
    type: gauge
    help: 'SQL Server committed memory in the memory manager.'
    values: [committed_memory_bytes]
    query_ref: mssql_os_sys_info

  - metric_name: mssql_server_target_memory_bytes
    type: gauge
    help: 'SQL Server target committed memory set for the memory manager.'
    values: [committed_memory_target_bytes]
    query_ref: mssql_os_sys_info

  #
  # Collected from sys.dm_os_sys_memory
  #
  - metric_name: mssql_os_memory
    type: gauge
    help: 'OS physical memory, used and available.'
    value_label: 'state'
    values: [used, available]
    query: |
      SELECT
        (total_physical_memory_kb - available_physical_memory_kb) * 1024 AS used,
        available_physical_memory_kb * 1024 AS available
      FROM sys.dm_os_sys_memory
  - metric_name: mssql_os_page_file
    type: gauge
    help: 'OS page file, used and available.'
    value_label: 'state'
    values: [used, available]
    query: |
      SELECT
        (total_page_file_kb - available_page_file_kb) * 1024 AS used,
        available_page_file_kb * 1024 AS available
      FROM sys.dm_os_sys_memory
queries:
  # Populates `mssql_io_stall` and `mssql_io_stall_total`
  - query_name: mssql_io_stall
    query: |
      SELECT
        cast(DB_Name(a.database_id) as varchar) AS [db],
        sum(io_stall_read_ms) / 1000.0 AS [read],
        sum(io_stall_write_ms) / 1000.0 AS [write]
      FROM
        sys.dm_io_virtual_file_stats(null, null) a
      INNER JOIN sys.master_files b ON a.database_id = b.database_id AND a.file_id = b.file_id
      GROUP BY a.database_id
  # Populates `mssql_resident_memory_bytes`, `mssql_virtual_memory_bytes`, mssql_available_commit_memory_bytes,
  # and `mssql_memory_utilization_percentage`, and `mssql_page_fault_count_total`
  - query_name: mssql_process_memory
    query: |
      SELECT
        physical_memory_in_use_kb * 1024 AS resident_memory_bytes,
        virtual_address_space_committed_kb * 1024 AS virtual_memory_bytes,
        available_commit_limit_kb * 1024 AS available_commit_limit_bytes,
        memory_utilization_percentage,
        page_fault_count
      FROM sys.dm_os_process_memory
  # Populates `mssql_server_total_memory_bytes` and `mssql_server_target_memory_bytes`.
  - query_name: mssql_os_sys_info
    query: |
      SELECT
        committed_kb * 1024 AS committed_memory_bytes,
        committed_target_kb * 1024 AS committed_memory_target_bytes
      FROM sys.dm_os_sys_info
```
