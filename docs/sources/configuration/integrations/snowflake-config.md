---
aliases:
- /docs/agent/latest/configuration/integrations/snowflake-config/
title: snowflake_config
---

# snowflake_config

The `snowflake_config` block configures the `snowflake` integration,
which is an embedded version of
[`snowflake-prometheus-exporter`](https://github.com/grafana/snowflake-prometheus-exporter). This allows the collection of [Snowflake](https://www.snowflake.com/) metrics.

Full reference of options:

```yaml
  # Enables the apache_http integration, allowing the Agent to automatically
  # collect metrics for the specified snowflake account.
  [enabled: <boolean> | default = false]

  # Sets an explicit value for the instance label when the integration is
  # self-scraped. Overrides inferred values.
  #
  # The default value for this integration is the configured account_name.
  [instance: <string>]

  # Automatically collect metrics from this integration. If disabled,
  # the snowflake integration will be run but not scraped and thus not
  # remote-written. Metrics for the integration will be exposed at
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

  # Relabel metrics coming from the integration, allowing to drop series
  # from the integration that you don't care about.
  metric_relabel_configs:
    [ - <relabel_config> ... ]

  # How frequent to truncate the WAL for this integration.
  [wal_truncate_frequency: <duration> | default = "60m"]

  #
  # Exporter-specific configuration options
  #
     
  # The account name of the snowflake account to monitor
  account_name: <string>

  # Username for the database user used to scrape metrics.
  username: <string>

  # Password for the database user used to scrape metrics
  password: <string>

  # The warehouse to use when querying metrics. 
  warehouse: <string>

  # The role to use when connecting to the database. The ACCOUNTADMIN role is used by default.
  [role: <string> | default = "ACCOUNTADMIN"]

```
