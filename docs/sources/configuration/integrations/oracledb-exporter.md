---
aliases:
- /docs/agent/latest/configuration/integrations/oracledb-exporter-config/
title: oracledb
---

# oracledb_config

The `oracledb_config` block configures the `oracle` integration,
which is an embedded version of a forked version of the
[`oracledb_exporter`](https://github.com/observiq/oracledb_exporter). This allows the collection of third party [OracleDB](https://www.oracle.com/database/) metrics.

Full reference of options:

```yaml
  # Enables the snowflake integration, allowing the Agent to automatically
  # collect metrics for the specified snowflake account.
  [enabled: <boolean> | default = false]

  # Sets an explicit value for the instance label when the integration is
  # self-scraped. Overrides inferred values.
  #
  # The default value for this integration is the configured account_name.
  [instance: <string>]

  # Automatically collect metrics from this integration. If disabled,
  # the snowflake integration is run but not scraped and thus not
  # remote-written. Metrics for the integration are exposed at
  # /integrations/snowflake/metrics and can be scraped by an external
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
  # from the integration that you don't care about from the integration.
  metric_relabel_configs:
    [ - <relabel_config> ... ]

  # How frequently the WAL is truncated for this integration.
  [wal_truncate_frequency: <duration> | default = "60m"]

  #
  # Exporter-specific configuration options
  #
     
  # The connection string used to connect to the OracleDB instance in the format 
  # of <MONITOR_USER>/<PASSWORD>@//<HOST>:<PORT>/<SERVICE>.
  # i.e. "oracle://user:password@localhost:1521/orcl.localnet"
  [connection_string: <string>]

  # The maximum amount of connections allowed to be idle of the exporter.
  [max_idle_connections: <int>]
  # The maximum amount of connections allowed to be open by the exporter.
  [max_open_connections: <int>]
  # The number of seconds that will act as the query timeout when the exporter is querying against
  # the OracleDB instance.
  [query_timeout: <int> | default = 5]

```
