---
# NOTE(rfratto): the title below has zero-width spaces injected into it to
# prevent it from overflowing the sidebar on the rendered site. Be careful when
# modifying this section to retain the spaces.
#
# Ideally, in the future, we can fix the overflow issue with css rather than
# injecting special characters.

title: prometheus.exporter.statsd
---

# prometheus.exporter.statsd
The `prometheus.exporter.statsd` component embeds
[statsd_exporter](https://github.com/prometheus/statsd_exporter) for collecting StatsD-style metrics and exporting them as Prometheus metrics.

## Usage

```river
prometheus.exporter.statsd "LABEL" {

}
```

## Arguments
The following arguments can be used to configure the exporter's behavior.
All arguments are optional. Omitted fields take their default values.

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`listen_udp`                                      | `string`       | The UDP address on which to receive statsd metric lines. Use "" to disable it. | `:9125` | no
`listen_tcp`                                      | `string`       | The TCP address on which to receive statsd metric lines. Use "" to disable it. | `:9125` | no
`listen_unixgram`                                 | `string`       | The Unixgram socket path to receive statsd metric lines in datagram. Use "" to disable it. | | no
`unix_socket_mode`                                | `string`       | The permission mode of the unix socket. | `755` | no
`mapping_config_path`                             | `string`       | The path to a YAML mapping file used to translate specific dot-separated StatsD metrics into labeled Prometheus metrics. | | no
`read_buffer`                                     | `int`          | Size (in bytes) of the operating system's transmit read buffer associated with the UDP or Unixgram connection. | | no
`cache_size`                                      | `int`          | Maximum size of your metric mapping cache. Relies on least recently used replacement policy if max size is reached. | `1000` | no
`cache_type`                                      | `string`       | Metric mapping cache type. Valid options are "lru" and "random". | `lru` | no
`event_queue_size`                                | `int`          | Size of internal queue for processing events. | `10000` | no
`event_flush_threshold`                           | `int`          | Number of events to hold in queue before flushing. | `1000`| no
`event_flush_interval`                            | `string`       | Maximum time between event queue flushes. | `200ms`| no
`parse_dogstatsd_tags`                            | `string`       | Parse DogStatsd style tags. | `true`| no
`parse_influxdb_tags`                             | `string`       | Parse InfluxDB style tags. | `true`| no
`parse_librato_tags`                              | `string`       | Parse Librato style tags. | `true`| no
`parse_signalfx_tags`                             | `string`       | Parse SignalFX style tags. | `true`| no

At least one of `listen_udp`, `listen_tcp`, or `listen_unixgram` should be enabled.
For details on how to use the mapping config file, please check the official
[statsd_exporter docs](https://github.com/prometheus/statsd_exporter#metric-mapping-and-configuration).
Please make sure the kernel parameter `net.core.rmem_max` is set to a value greater
than the value specified in `read_buffer`.

### Blocks

The `prometheus.exporter.statsd` component does not support any blocks, and is configured
fully through arguments.

## Exported fields
The following fields are exported and can be referenced by other components.

Name      | Type                | Description
--------- | ------------------- | -----------
`targets` | `list(map(string))` | The targets that can be used to collect `statsd` metrics.

For example, the `targets` could either be passed to a `prometheus.relabel`
component to rewrite the metrics' label set, or to a `prometheus.scrape`
component that collects the exposed metrics.

The exported targets will use the configured [in-memory traffic][] address
specified by the [run command][].

[in-memory traffic]: {{< relref "../../concepts/component_controller.md#in-memory-traffic" >}}
[run command]: {{< relref "../cli/run.md" >}}

## Component health

`prometheus.exporter.statsd` is only reported as unhealthy if given
an invalid configuration. In those cases, exported fields retain their last
healthy values.

## Debug information

`prometheus.exporter.statsd` does not expose any component-specific
debug information.

## Debug metrics

`prometheus.exporter.statsd` does not expose any component-specific
debug metrics.

## Example

This example uses a [`prometheus.scrape` component][scrape] to collect metrics
from `prometheus.exporter.statsd`:

```river
prometheus.exporter.statsd "example" {
  listen_udp = ""
  listen_tcp = ":9125"
  listen_unixgram = ""
  unix_socket_mode = "755"
  mapping_config_path = "mapTest.yaml"
  read_buffer = 1
  cache_size = 1000
  cache_type = "lru"
  event_queue_size = 10000
  event_flush_threshold = 1000
  event_flush_interval = "200ms"
  parse_dogstatsd_tags = true
  parse_influxdb_tags = true
  parse_librato_tags = true
  parse_signalfx_tags = true
}

// Configure a prometheus.scrape component to collect statsd metrics.
prometheus.scrape "demo" {
  targets    = prometheus.exporter.statsd.example.targets
  forward_to = [ /* ... */ ]
}
```

[scrape]: {{< relref "./prometheus.scrape.md" >}}
