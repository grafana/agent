---
# NOTE(rfratto): the title below has zero-width spaces injected into it to
# prevent it from overflowing the sidebar on the rendered site. Be careful when
# modifying this section to retain the spaces.
#
# Ideally, in the future, we can fix the overflow issue with css rather than
# injecting special characters.

title: prometheus.exporter.kafka
---

# prometheus.exporter.kafka
The `prometheus.exporter.kafka` component embeds
[kafka_exporter](https://github.com/davidmparrott/kafka_exporter/v2/exporter) for collecting metrics from a kafka server.

## Usage

```river
prometheus.exporter.kafka "LABEL" {
    kafka_uris = ["localhost:9200"]
}
```

## Arguments
The following arguments can be used to configure the exporter's behavior.
Omitted fields take their default values.

| Name                        | Type            | Description                                                                                                                                                                         | Default | Required |
|-----------------------------|-----------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------|----------|
| `kafka_uris`                | `array(string)` | Address array (host:port) of Kafka server.                                                                                                                                          |         | yes      |
| `instance`                  | `string`        | The`instance`label for metrics, default is the hostname:port of the first kafka_uris, If there is more than one string in kafka_uris, the instance value must be manually provided. |         | no       |
| `use_sasl`                  | `bool`          | Connect using SASL/PLAIN.                                                                                                                                                           |         | no       |
| `use_sasl_handshake`        | `bool`          | Only set this to false if using a non-Kafka SASL proxy.                                                                                                                             | `false` | no       |
| `sasl_username`             | `string`        | SASL user name.                                                                                                                                                                     |         | no       |
| `sasl_password`             | `string`        | SASL user password.                                                                                                                                                                 |         | no       |
| `sasl_mechanism`            | `string`        | The SASL SCRAM SHA algorithm sha256 or sha512 as mechanism.                                                                                                                         |         | no       |
| `use_tls`                   | `bool`          | Connect using TLS.                                                                                                                                                                  |         | no       |
| `ca_file`                   | `string`        | The optional certificate authority file for TLS client authentication.                                                                                                              |         | no       |
| `cert_file`                 | `string`        | The optional certificate file for TLS client authentication.                                                                                                                        |         | no       |
| `key_file`                  | `string`        | The optional key file for TLS client authentication.                                                                                                                                |         | no       |
| `insecure_skip_verify`      | `bool`          | If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure.                                                                 |         | no       |
| `kafka_version`             | `string`        | Kafka broker version.                                                                                                                                                               | `2.0.0` | no       |
| `use_zookeeper_lag`         | `bool`          | if you need to use a group from zookeeper.                                                                                                                                          |         | no       |
| `zookeeper_uris`            | `array(string)` | Address array (hosts) of zookeeper server..                                                                                                                                         |         | no       |
| `kafka_cluster_name`        | `string`        | Kafka cluster name.                                                                                                                                                                 |         | no       |
| `metadata_refresh_interval` | `duration`      | Metadata refresh interval.                                                                                                                                                          | `1m`    | no       |
| `allow_concurrency`         | `bool`          | If true, all scrapes will trigger kafka operations otherwise, they will share results. WARN: This should be disabled on large clusters.                                             | `true`  | no       |
| `max_offsets`               | `int`           | Maximum number of offsets to store in the interpolation table for a partition.                                                                                                      | `1000`  | no       |
| `prune_interval_seconds`    | `int`           | How frequently should the interpolation table be pruned, in seconds.                                                                                                                | `30`    | no       |
| `topics_filter_regex`       | `string`        | Regex filter for topics to be monitored.                                                                                                                                            | `.*`    | no       |
| `groups_filter_regex`       | `string`        | Regex filter for consumer groups to be monitored.                                                                                                                                   | `.*`    | no       |

## Exported fields
The following fields are exported and can be referenced by other components.

| Name      | Type                | Description                                              |
|-----------|---------------------|----------------------------------------------------------|
| `targets` | `list(map(string))` | The targets that can be used to collect `kafka` metrics. |

For example, `targets` can either be passed to a `prometheus.relabel`
component to rewrite the metrics' label set, or to a `prometheus.scrape`
component that collects the exposed metrics.

The exported targets will use the configured [in-memory traffic][] address
specified by the [run command][].

[in-memory traffic]: {{< relref "../../concepts/component_controller.md#in-memory-traffic" >}}
[run command]: {{< relref "../cli/run.md" >}}

## Component health

`prometheus.exporter.kafka` is only reported as unhealthy if given
an invalid configuration. In those cases, exported fields retain their last
healthy values.

## Debug information

`prometheus.exporter.kafka` does not expose any component-specific
debug information.

## Debug metrics

`prometheus.exporter.kafka` does not expose any component-specific
debug metrics.

## Example

This example uses a [`prometheus.scrape` component][scrape] to collect metrics
from `prometheus.exporter.kafka`:

```river
prometheus.exporter.kafka "example" {
  kafka_uris = ["localhost:9200"]
}

// Configure a prometheus.scrape component to send metrics to.
prometheus.scrape "demo" {
  targets    = [prometheus.exporter.kafka.example.targets]
  forward_to = [prometheus.remote_write.demo.receiver]
}

// prometheus.remote_write component.
prometheus.remote_write "demo" {
  endpoint {
    url = "http://mimir:9009/api/v1/push"
    basic_auth {
      username = "example-user"
      password = "example-password"
    }
  }
}

```

[scrape]: {{< relref "./prometheus.scrape.md" >}}
