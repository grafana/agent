---
aliases:
- /docs/grafana-cloud/agent/flow/reference/components/prometheus.exporter.kafka/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/reference/components/prometheus.exporter.kafka/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/reference/components/prometheus.exporter.kafka/
canonical: https://grafana.com/docs/agent/latest/flow/reference/components/prometheus.exporter.kafka/
title: prometheus.exporter.kafka
description: Learn about prometheus.exporter.kafka
---

# prometheus.exporter.kafka

The `prometheus.exporter.kafka` component embeds
[kafka_exporter](https://github.com/davidmparrott/kafka_exporter) for collecting metrics from a kafka server.

## Usage

```river
prometheus.exporter.kafka "LABEL" {
    kafka_uris = KAFKA_URI_LIST
}
```

## Arguments

You can use the following arguments to configure the exporter's behavior.
Omitted fields take their default values.

| Name                        | Type            | Description                                                                                                                                                                        | Default | Required |
| --------------------------- | --------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------- | -------- |
| `kafka_uris`                | `array(string)` | Address array (host:port) of Kafka server.                                                                                                                                         |         | yes      |
| `instance`                  | `string`        | The`instance`label for metrics, default is the hostname:port of the first kafka_uris. You must manually provide the instance value if there is more than one string in kafka_uris. |         | no       |
| `use_sasl`                  | `bool`          | Connect using SASL/PLAIN.                                                                                                                                                          |         | no       |
| `use_sasl_handshake`        | `bool`          | Only set this to false if using a non-Kafka SASL proxy.                                                                                                                            | `false` | no       |
| `sasl_username`             | `string`        | SASL user name.                                                                                                                                                                    |         | no       |
| `sasl_password`             | `string`        | SASL user password.                                                                                                                                                                |         | no       |
| `sasl_mechanism`            | `string`        | The SASL SCRAM SHA algorithm sha256 or sha512 as mechanism.                                                                                                                        |         | no       |
| `use_tls`                   | `bool`          | Connect using TLS.                                                                                                                                                                 |         | no       |
| `ca_file`                   | `string`        | The optional certificate authority file for TLS client authentication.                                                                                                             |         | no       |
| `cert_file`                 | `string`        | The optional certificate file for TLS client authentication.                                                                                                                       |         | no       |
| `key_file`                  | `string`        | The optional key file for TLS client authentication.                                                                                                                               |         | no       |
| `insecure_skip_verify`      | `bool`          | If set to true, the server's certificate will not be checked for validity. This makes your HTTPS connections insecure.                                                             |         | no       |
| `kafka_version`             | `string`        | Kafka broker version.                                                                                                                                                              | `2.0.0` | no       |
| `use_zookeeper_lag`         | `bool`          | If set to true, use a group from zookeeper.                                                                                                                                        |         | no       |
| `zookeeper_uris`            | `array(string)` | Address array (hosts) of zookeeper server.                                                                                                                                         |         | no       |
| `kafka_cluster_name`        | `string`        | Kafka cluster name.                                                                                                                                                                |         | no       |
| `metadata_refresh_interval` | `duration`      | Metadata refresh interval.                                                                                                                                                         | `1m`    | no       |
| `allow_concurrency`         | `bool`          | If set to true, all scrapes trigger Kafka operations. Otherwise, they will share results. WARNING: Disable this on large clusters.                                                 | `true`  | no       |
| `max_offsets`               | `int`           | The maximum number of offsets to store in the interpolation table for a partition.                                                                                                 | `1000`  | no       |
| `prune_interval_seconds`    | `int`           | How frequently should the interpolation table be pruned, in seconds.                                                                                                               | `30`    | no       |
| `topics_filter_regex`       | `string`        | Regex filter for topics to be monitored.                                                                                                                                           | `.*`    | no       |
| `groups_filter_regex`       | `string`        | Regex filter for consumer groups to be monitored.                                                                                                                                  | `.*`    | no       |

## Exported fields

{{< docs/shared lookup="flow/reference/components/exporter-component-exports.md" source="agent" version="<AGENT VERSION>" >}}

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
  targets    = prometheus.exporter.kafka.example.targets
  forward_to = [prometheus.remote_write.demo.receiver]
}

prometheus.remote_write "demo" {
  endpoint {
    url = PROMETHEUS_REMOTE_WRITE_URL

    basic_auth {
      username = USERNAME
      password = PASSWORD
    }
  }
}
```

Replace the following:

- `PROMETHEUS_REMOTE_WRITE_URL`: The URL of the Prometheus remote_write-compatible server to send metrics to.
- `USERNAME`: The username to use for authentication to the remote_write API.
- `PASSWORD`: The password to use for authentication to the remote_write API.

[scrape]: {{< relref "./prometheus.scrape.md" >}}
