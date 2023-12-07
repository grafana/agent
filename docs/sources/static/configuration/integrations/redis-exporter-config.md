---
aliases:
- ../../../configuration/integrations/redis-exporter-config/
- /docs/grafana-cloud/monitor-infrastructure/agent/static/configuration/integrations/redis-exporter-config/
- /docs/grafana-cloud/send-data/agent/static/configuration/integrations/redis-exporter-config/
canonical: https://grafana.com/docs/agent/latest/static/configuration/integrations/redis-exporter-config/
description: Learn about redis_exporter_config
title: redis_exporter_config
---

# redis_exporter_config

The `redis_exporter_config` block configures the `redis_exporter` integration, which is an embedded version of [`redis_exporter`](https://github.com/oliver006/redis_exporter). This allows for the collection of metrics from Redis servers.

Note that currently, an Agent can only collect metrics from a single Redis server. If you want to collect metrics from multiple Redis servers, you can run multiple Agents and add labels using `relabel_configs` to differentiate between the Redis servers:

```yaml
redis_exporter:
  enabled: true
  redis_addr: "redis-2:6379"
  relabel_configs:
  - source_labels: [__address__]
    target_label: instance
    replacement: redis-2
```

We strongly recommend that you configure a separate user for the Agent, and give it only the strictly mandatory
security privileges necessary for monitoring your node, as per the [official documentation](https://github.com/oliver006/redis_exporter#authenticating-with-redis).

Full reference of options:
```yaml
  # Enables the redis_exporter integration, allowing the Agent to automatically
  # collect system metrics from the configured redis address
  [enabled: <boolean> | default = false]

  # Sets an explicit value for the instance label when the integration is
  # self-scraped. Overrides inferred values.
  #
  # The default value for this integration is inferred from the hostname
  # portion of redis_addr.
  [instance: <string>]

  # Automatically collect metrics from this integration. If disabled,
  # the redis_exporter integration will be run but not scraped and thus not
  # remote-written. Metrics for the integration will be exposed at
  # /integrations/redis_exporter/metrics and can be scraped by an external
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

  # Monitor the exporter itself and include those metrics in the results.
  [include_exporter_metrics: <bool> | default = false]

  # exporter-specific configuration options

  # Address of the redis instance.
  redis_addr: <string>

  # User name to use for authentication (Redis ACL for Redis 6.0 and newer).
  [redis_user: <string>]

  # Password of the redis instance.
  [redis_password: <string>]

  # Path of a file containing a passord. If this is defined, it takes precedece
  # over redis_password.
  [redis_password_file: <string>]

  # Path of a file containing a JSON object which maps Redis URIs [string] to passwords [string]
  # (e.g. {"redis://localhost:6379": "sample_password"}).
  [redis_password_map_file: <string>]

  # Namespace for the metrics.
  [namespace: <string> | default = "redis"]

  # What to use for the CONFIG command.
  [config_command: <string> | default = "CONFIG"]

  # Comma separated list of key-patterns to export value and length/size, searched for with SCAN.
  [check_keys: <string>]

  # Comma separated list of LUA regex for grouping keys. When unset, no key
  # groups will be made.
  [check_key_groups: <string>]

  # Check key or key groups batch size hint for the underlying SCAN. Keeping the same name for backwards compatibility, but this applies to both key and key groups batch size configuration.
  [check_key_groups_batch_size: <int> | default = 10000]

  # The maximum number of distinct key groups with the most memory utilization
  # to present as distinct metrics per database. The leftover key groups will be
  # aggregated in the 'overflow' bucket.
  [max_distinct_key_groups: <int> | default = 100]

  # Comma separated list of single keys to export value and length/size.
  [check_single_keys: <string>]

  # Comma separated list of stream-patterns to export info about streams, groups and consumers, searched for with SCAN.
  [check_streams: <string>]

  # Comma separated list of single streams to export info about streams, groups and consumers.
  [check_single_streams: <string>]

  # Whether to export key values as labels when using `check_keys` or `check_single_keys`.
  [export_key_values: <bool> | default = true]

  # Comma separated list of individual keys to export counts for.
  [count_keys: <string>]

  # Comma-separated list of paths to Lua Redis scripts for collecting extra metrics.
  [script_path: <string>]

  # Timeout for connection to Redis instance (in Golang duration format).
  [connection_timeout: <time.Duration> | default = "15s"]

  # Name of the client key file (including full path) if the server requires TLS client authentication.
  [tls_client_key_file: <string>]

  # Name of the client certificate file (including full path) if the server requires TLS client authentication.
  [tls_client_cert_file: <string>]

  # Name of the CA certificate file (including full path) if the server requires TLS client authentication.
  [tls_ca_cert_file: <string>]

  # Whether to set client name to redis_exporter.
  [set_client_name: <bool>]

  # Whether to scrape Tile38 specific metrics.
  [is_tile38: <bool>]

  # Whether this is a redis cluster (Enable this if you need to fetch key level data on a Redis Cluster).
  [is_cluster: <bool> | default = false]

  # Whether to scrape Client List specific metrics.
  [export_client_list: <bool>]

  # Whether to include the client's port when exporting the client list. Note
  # that including this will increase the cardinality of all redis metrics.
  [export_client_port: <bool>]

  # Whether to also export go runtime metrics.
  [redis_metrics_only: <bool>]

  # Whether to ping the redis instance after connecting.
  [ping_on_connect: <bool>]

  # Whether to include system metrics like e.g. redis_total_system_memory_bytes.
  [incl_system_metrics: <bool>]

  # Whether to to skip TLS verification.
  [skip_tls_verification: <bool>]
```
