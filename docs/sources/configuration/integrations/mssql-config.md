---
title: mssql_config
---

# mssql_config

The `mssql_configs` block configures the `mssql` integration, an embedded version of [`sql_exporter`](https://github.com/burningalchemist/sql_exporter) that lets you collect [Microsoft SQL Server](https://www.microsoft.com/en-us/sql-server) metrics.

It is recommended that you have a dedicated user set up for monitoring an mssql instance.
The user for monitoring must have the following grants in order to populate the metrics:
```
GRANT VIEW ANY DEFINITION TO <MONITOR_USER>
GRANT VIEW SERVER STATE TO <MONITOR_USER>
```


Full reference of options:

```yaml
  # Enables the mssql integration, allowing the Agent to automatically
  # collect metrics for the specified mssql instance.
  [enabled: <boolean> | default = false]

  # Sets an explicit value for the instance label when the integration is
  # self-scraped. Overrides inferred values.
  #
  # The default value for this integration is the host:port of the provided connection_string.
  [instance: <string>]

  # Automatically collect metrics from this integration. If disabled,
  # the mssql integration is run but not scraped and thus not
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
     
  # The connection_string to use to connect to the mssql instance.
  # It is specified in the form of: "sqlserver://<USERNAME>:<PASSWORD>@<HOST>:<PORT>"
  connection_string: <string>

  # The maximum number of open database connections to the mssql instance.
  [max_open_connections: <int> | default = 3]

  # The maximum number of idle database connections to the mssql instance.
  [max_idle_connections: <int> | default = 3]

  # The timeout for scraping metrics from the mssql instance.
  [timeout: <duration> | default = "10s"]

```
