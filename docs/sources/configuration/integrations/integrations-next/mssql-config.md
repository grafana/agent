---
aliases:
  - /docs/agent/latest/configuration/integrations/integrations-next/mssql-config/
title: mssql_config
---

# mssql_config (beta)

The `mssql_configs` block configures the `mssql` integration,
which is an embedded version of [`sql_exporter`](https://github.com/burningalchemist/sql_exporter).
This allows the collection of [Microsoft SQL Server](https://www.microsoft.com/en-us/sql-server) metrics.

Configuration reference:

```yaml
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

  # Integration instance name. 
  # The default value for this integration is the host:port of the provided connection_string.
  [instance: <string> | default = <host:port of connection_string>]

  # The connection_string to use to connect to the mssql instance.
  # It is specified in the form of: "sqlserver://<USERNAME>:<PASSWORD>@<HOST>:<PORT>"
  connection_string: <string>

  # The maximum number of open database connections to the mssql instance.
  [max_connections: <int> | default = 3]

  # The maximum number of idle database connections to the mssql instance.
  [max_idle_connections: <int> | default = 3]

  # The timeout for scraping metrics from the mssql instance.
  [timeout: <duration> | default = "10s"]
```

## Quick configuration example

```yaml
integrations:
  mssql_configs:
    - connection_string: "sqlserver://user:pass@localhost:1433"
      autoscrape:
        enable: true
        metrics_instance: default

metrics:
  wal_directory: /tmp/grafana-agent-wal
server:
  log_level: debug
```
