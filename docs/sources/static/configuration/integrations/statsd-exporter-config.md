---
aliases:
- ../../../configuration/integrations/statsd-exporter-config/
- /docs/grafana-cloud/monitor-infrastructure/agent/static/configuration/integrations/statsd-exporter-config/
- /docs/grafana-cloud/send-data/agent/static/configuration/integrations/statsd-exporter-config/
canonical: https://grafana.com/docs/agent/latest/static/configuration/integrations/statsd-exporter-config/
description: Learn about statsd_exporter_config
title: statsd_exporter_config
---

# statsd_exporter_config

The `statsd_exporter_config` block configures the `statsd_exporter`
integration, which is an embedded version of
[`statsd_exporter`](https://github.com/prometheus/statsd_exporter). This allows
for the collection of statsd metrics and exposing them as Prometheus metrics.

Full reference of options:

```yaml
  # Enables the statsd_exporter integration, allowing the Agent to automatically
  # collect system metrics from the configured statsd server address
  [enabled: <boolean> | default = false]

  # Sets an explicit value for the instance label when the integration is
  # self-scraped. Overrides inferred values.
  #
  # The default value for this integration is inferred from the agent hostname
  # and HTTP listen port, delimited by a colon.
  [instance: <string>]

  # Automatically collect metrics from this integration. If disabled,
  # the statsd_exporter integration will be run but not scraped and thus not
  # remote-written. Metrics for the integration will be exposed at
  # /integrations/statsd_exporter/metrics and can be scraped by an external
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

  # The UDP address on which to receive statsd metric lines. An empty string
  # will disable UDP collection.
  [listen_udp: <string> | default = ":9125"]

  # The TCP address on which to receive statsd metric lines. An empty string
  # will disable TCP collection.
  [listen_tcp: <string> | default = ":9125"]

  # The Unixgram socket path to receive statsd metric lines. An empty string
  # will disable unixgram collection.
  [listen_unixgram: <string> | default = ""]

  # The permission mode of the unixgram socket, when enabled.
  [unix_socket_mode: <string> | default = "755"]

  # An optional mapping config that can translate dot-separated StatsD metrics
  # into labeled Prometheus metrics. For full instructions on how to write this
  # object, see the official documentation from the statsd_exporter:
  #
  # https://github.com/prometheus/statsd_exporter#metric-mapping-and-configuration
  #
  # Note that a SIGHUP will not reload this config.
  [mapping_config: <statsd_exporter.mapping_config>]

  # Size (in bytes) of the operating system's transmit read buffer associated
  # with the UDP or unixgram connection. Please make sure the kernel parameters
  # net.core.rmem_max is set to a value greater than the value specified.
  [read_buffer: <int> | default = 0]

  # Maximum size of your metric mapping cache. Relies on least recently used
  # replacement policy if max size is reached.
  [cache_size: <int> | default = 1000]

  # Metric mapping cache type. Valid values are "lru" and "random".
  [cache_type: <string> | default = "lru"]

  # Size of internal queue for processing events.
  [event_queue_size: <int> | default = 10000]

  # Number of events to hold in queue before flushing.
  [event_flush_threshold: <int> | default = 1000]

  # Number of events to hold in queue before flushing.
  [event_flush_interval: <duration> | default = "200ms"]

  # Parse DogStatsd style tags.
  [parse_dogstatsd_tags: <bool> | default = true]

  # Parse InfluxDB style tags.
  [parse_influxdb_tags: <bool> | default = true]

  # Parse Librato style tags.
  [parse_librato_tags: <bool> | default = true]

  # Parse SignalFX style tags.
  [parse_signalfx_tags: <bool> | default = true]

  # Optional: Relay address configuration. This setting, if provided,
  # specifies the destination to forward your metrics.

  # Note that it must be a UDP endpoint in the format 'host:port'.
  [relay_address: <string>]

  # Maximum relay output packet length to avoid fragmentation.
  [relay_packet_length: <int> | default = 1400]
```
