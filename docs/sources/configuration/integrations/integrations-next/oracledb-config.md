---
aliases:
- /docs/agent/latest/configuration/integrations/integrations-next/oracledb-config/
title: oracledb
---

# oracledb_config

The `oracledb_config` block configures the `oracledb` integration,
which is an embedded version of a forked version of the
[`oracledb_exporter`](https://github.com/observiq/oracledb_exporter). This allows the collection of third party [OracleDB](https://www.oracle.com/database/) metrics.

Full reference of options:

```yaml
  # Override autoscrape defaults for this integration.
  autoscrape:
    # Enables autoscrape of integrations.
    [enable: <boolean> | default = true]

    # Specifies the metrics instance name to send metrics to. Instance
    # names are located at metrics.configs[].name from the top-level config.
    # The instance must exist.
    #
    # As it is common to use the name "default" for your primary instance,
    # we assume the same here.
    [metrics_instance: <string> | default = "default"]

    # Autoscrape interval and timeout. Defaults are inherited from the global
    # section of the top-level metrics config.
    [scrape_interval: <duration> | default = <metrics.global.scrape_interval>]
    [scrape_timeout: <duration> | default = <metrics.global.scrape_timeout>]

  # An optional extra set of labels to add to metrics from the integration target. These
  # labels are only exposed via the integration service discovery HTTP API and
  # added when autoscrape is used. They will not be found directly on the metrics
  # page for an integration.
  extra_labels:
    [ <labelname>: <labelvalue> ... ]

  # Sets an explicit value for the instance label when the integration is
  # self-scraped. Overrides inferred values.
  #
  # The default value for this integration is the `host:port` of the connection string
  [instance: <string>]

  #
  # Exporter-specific configuration options
  #
     
  # The connection string used to connect to the OracleDB instance in the format 
  # of oracle://<MONITOR_USER>:<PASSWORD>@<HOST>:<PORT>/<SERVICE>.
  # i.e. "oracle://user:password@localhost:1521/orcl.localnet"
  # If not inside the YAML this value by default is read from the environment variable `$DATA_SOURCE_NAME`
  [connection_string: <string>]

  # The maximum amount of connections allowed to be idle of the exporter.
  [max_idle_connections: <int>]
  # The maximum amount of connections allowed to be open by the exporter.
  [max_open_connections: <int>]

  # This is the interval between each scrape. Default of 0 is to scrape on collect requests. 
  [metrics_scrape_interval: <duration> | default = "0"]

  # The number of seconds that will act as the query timeout when the exporter is querying against
  # the OracleDB instance.
  [query_timeout: <int> | default = 5]

```

## Quick configuration example

```yaml
integrations:
  oracledb_configs:
  - connection_string: oracle://user:password@localhost:1521/orcl.localnet
    metrics_scrape_interval: 1m
    max_idle_connections: 0
    max_open_connections: 10
    query_timeout: 5

metrics:
  wal_directory: /tmp/grafana-agent-wal
server:
  log_level: debug
```
