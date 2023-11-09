---
aliases:
- ../../../configuration/integrations/oracledb-config/
- /docs/grafana-cloud/monitor-infrastructure/agent/static/configuration/integrations/oracledb-config/
- /docs/grafana-cloud/send-data/agent/static/configuration/integrations/oracledb-config/
canonical: https://grafana.com/docs/agent/latest/static/configuration/integrations/oracledb-config/
description: Learn about oracledb_config
title: oracledb_config
---

# oracledb_config

The `oracledb_config` block configures the `oracledb` integration,
which is an embedded version of a forked version of the
[`oracledb_exporter`](https://github.com/observiq/oracledb_exporter). This allows the collection of third party [OracleDB](https://www.oracle.com/database/) metrics.

Full reference of options:

```yaml
  # Enables the oracledb integration, allowing the Agent to automatically
  # collect metrics for the specified oracledb instance.
  [enabled: <boolean> | default = false]

  # Sets an explicit value for the instance label when the integration is
  # self-scraped. Overrides inferred values.
  #
  # The default value for this integration is the configured host:port of the connection string.
  [instance: <string>]

  # Automatically collect metrics from this integration. If disabled,
  # the oracledb integration is run but not scraped and thus not
  # remote-written. Metrics for the integration are exposed at
  # /integrations/oracledb/metrics and can be scraped by an external
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

  # The connection string used to connect to the OracleDB instance in the format
  # of oracle://<MONITOR_USER>:<PASSWORD>@<HOST>:<PORT>/<SERVICE>.
  # i.e. "oracle://user:password@localhost:1521/orcl.localnet"
  [connection_string: <string>]

  # The maximum amount of connections of the exporter allowed to be idle.
  [max_idle_connections: <int>]
  # The maximum amount of connections allowed to be open by the exporter.
  [max_open_connections: <int>]

  # The number of seconds that will act as the query timeout when the exporter is querying against
  # the OracleDB instance.
  [query_timeout: <int> | default = 5]
```

## Configuration example

```yaml
integrations:
  oracledb:
    enabled: true
    connection_string: oracle://user:password@localhost:1521/orcl.localnet
    scrape_interval: 1m
    scrape_timeout: 1m
    scrape_integration: true
metrics:
  wal_directory: /tmp/grafana-agent-wal
server:
  log_level: debug
```
