---
title: snowflake_config
---

# snowflake_config (beta)

>**Note:** This config is associated with the [Integrations revamp experimental feature](_index.md) for multi-instance monitoring using a single agent. A GA snowflake_config for single-instance monitoring can be found [here](../snowflake-config.md).

The `snowflake_configs` block configures the `snowflake` integration,
which is an embedded version of
[`snowflake-prometheus-exporter`](https://github.com/grafana/snowflake-prometheus-exporter). This allows the collection of [Snowflake](https://www.snowflake.com/) metrics.

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
  # The default value for this integration is the configured account_name.
  [instance: <string> | default = <account_name>]

  # The account name of the snowflake account to monitor.
  account_name: <string>

  # Username for the database user used to scrape metrics.
  username: <string>

  # Password for the database user used to scrape metrics.
  password: <string>

  # The warehouse to use when querying metrics. 
  warehouse: <string>

  # The role to use when connecting to the database. The ACCOUNTADMIN role is used by default.
  [role: <string> | default = "ACCOUNTADMIN"]

```

## Quick configuration example

```yaml
integrations:
  snowflake_configs:
    - account_name: XXXXXXX-YYYYYYY
      username: snowflake-user
      password: snowflake-pass
      warehouse: SNOWFLAKE_WAREHOUSE
      role: ACCOUNTADMIN
      autoscrape:
        enable: true
        metrics_instance: default
        scrape_interval: 30m

metrics:
  wal_directory: /tmp/grafana-agent-wal
server:
  log_level: debug
```
