---
aliases:
- ../../../configuration/integrations/memcached-exporter-config/
- /docs/grafana-cloud/monitor-infrastructure/agent/static/configuration/integrations/memcached-exporter-config/
- /docs/grafana-cloud/send-data/agent/static/configuration/integrations/memcached-exporter-config/
canonical: https://grafana.com/docs/agent/latest/static/configuration/integrations/memcached-exporter-config/
description: Learn about memcached_exporter_config
title: memcached_exporter_config
---

# memcached_exporter_config

The `memcached_exporter_config` block configures the `memcached_exporter`
integration, which is an embedded version of
[`memcached_exporter`](https://github.com/prometheus/memcached_exporter). This
allows for the collection of metrics from memcached servers.

Note that currently, an Agent can only collect metrics from a single memcached
server. If you want to collect metrics from multiple servers, you can run
multiple Agents and add labels using `relabel_configs` to differentiate between
the servers:

```yaml
memcached_exporter:
  enabled: true
  memcached_address: memcached-a:53
  relabel_configs:
  - source_labels: [__address__]
    target_label: instance
    replacement: memcached-a
```

Full reference of options:

```yaml
  # Enables the memcached_exporter integration, allowing the Agent to automatically
  # collect system metrics from the configured memcached server address
  [enabled: <boolean> | default = false]

  # Sets an explicit value for the instance label when the integration is
  # self-scraped. Overrides inferred values.
  #
  # The default value for this integration is inferred from
  # memcached_address.
  [instance: <string>]

  # Automatically collect metrics from this integration. If disabled,
  # the memcached_exporter integration will be run but not scraped and thus not
  # remote-written. Metrics for the integration will be exposed at
  # /integrations/memcached_exporter/metrics and can be scraped by an external
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

  # Address of the memcached server in host:port form.
  [memcached_address: <string> | default = "localhost:53"]

  # Timeout for connecting to memcached.
  [timeout: <duration> | default = "1s"]

  # TLS configuration for requests to the memcached server.
  tls_config:
    # The CA cert to use.
    [ca: <string>]
    # The client cert to use.
    [cert: <string>]
    # The client key to use.
    [key: <string>]

    # Path to the CA cert file to use.
    [ca_file: <string>]
    # Path to the client cert file to use.
    [cert_file: <string>]
    # Path to the client key file to use.
    [key_file: <string>]

    # Used to verify the hostname for the memcached server.
    [server_name: <string>]

    # Disable memcached server certificate validation.
    [insecure_skip_verify: <boolean> | default = false]

    # Minimum TLS version.
    [min_version: <tls_version>]
    # Maximum TLS version.
    [max_version: <tls_version>]
```
