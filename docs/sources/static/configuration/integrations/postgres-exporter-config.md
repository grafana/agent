---
aliases:
- ../../../configuration/integrations/postgres-exporter-config/
- /docs/grafana-cloud/monitor-infrastructure/agent/static/configuration/integrations/postgres-exporter-config/
- /docs/grafana-cloud/send-data/agent/static/configuration/integrations/postgres-exporter-config/
canonical: https://grafana.com/docs/agent/latest/static/configuration/integrations/postgres-exporter-config/
description: Learn about postgres_exporter_config
title: postgres_exporter_config
---

# postgres_exporter_config

The `postgres_exporter_config` block configures the `postgres_exporter`
integration, which is an embedded version of
[`postgres_exporter`](https://github.com/prometheus-community/postgres_exporter). This
allows for the collection of metrics from Postgres servers.

We strongly recommend that you configure a separate user for the Agent, and give it only the strictly mandatory
security privileges necessary for monitoring your node, as per the [official documentation](https://github.com/prometheus-community/postgres_exporter#running-as-non-superuser).

Full reference of options:

```yaml
  # Enables the postgres_exporter integration, allowing the Agent to automatically
  # collect system metrics from the configured postgres server address
  [enabled: <boolean> | default = false]

  # Sets an explicit value for the instance label when the integration is
  # self-scraped. Overrides inferred values.
  #
  # The default value for this integration is inferred from a truncated version of
  # the first DSN in data_source_names. The truncated DSN includes the hostname
  # and database name (if used) of the server, but does not include any user
  # information.
  #
  # If data_source_names contains more than one entry, the integration will fail to
  # load and a value for instance must be manually provided.
  [instance: <string>]

  # Automatically collect metrics from this integration. If disabled,
  # the postgres_exporter integration will be run but not scraped and thus not
  # remote-written. Metrics for the integration will be exposed at
  # /integrations/postgres_exporter/metrics and can be scraped by an external
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

  # Data Source Names specifies the Postgres server(s) to connect to. This is
  # REQUIRED but may also be specified by the POSTGRES_EXPORTER_DATA_SOURCE_NAME
  # environment variable, where DSNs the environment variable are separated by
  # commas. If neither are set, the integration will fail to start.
  #
  # The format of this is specified here: https://pkg.go.dev/github.com/lib/pq#ParseURL
  #
  # A working example value for a server with a password is:
  # "postgresql://username:passwword@localhost:5432/database?sslmode=disable"
  #
  # Multiple DSNs may be provided here, allowing for scraping from multiple
  # servers.
  data_source_names:
  - <string>

  # Disables collection of metrics from pg_settings.
  [disable_settings_metrics: <boolean> | default = false]

  # Autodiscover databases to collect metrics from. If false, only collects
  # metrics from databases collected from data_source_names.
  [autodiscover_databases: <boolean> | default = false]

  # Excludes specific databases from being collected when autodiscover_databases
  # is true.
  exclude_databases:
  [ - <string> ]

  # Includes only specific databases (excluding all others) when autodiscover_databases
  # is true.
  include_databases:
  [ - <string> ]

  # Path to a YAML file containing custom queries to run. Check out
  # postgres_exporter's queries.yaml for examples of the format:
  # https://github.com/prometheus-community/postgres_exporter/blob/master/queries.yaml
  [query_path: <string> | default = ""]

  # When true, only exposes metrics supplied from query_path.
  [disable_default_metrics: <boolean> | default = false]
```
