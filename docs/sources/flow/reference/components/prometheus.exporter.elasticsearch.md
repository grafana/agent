---
aliases:
- /docs/grafana-cloud/agent/flow/reference/components/prometheus.exporter.elasticsearch/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/reference/components/prometheus.exporter.elasticsearch/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/reference/components/prometheus.exporter.elasticsearch/
canonical: https://grafana.com/docs/agent/latest/flow/reference/components/prometheus.exporter.elasticsearch/
title: prometheus.exporter.elasticsearch
description: Learn about prometheus.exporter.elasticsearch
---

# prometheus.exporter.elasticsearch

The `prometheus.exporter.elasticsearch` component embeds
[elasticsearch_exporter](https://github.com/prometheus-community/elasticsearch_exporter) for
the collection of metrics from ElasticSearch servers.

{{% admonition type="note" %}}
Currently, an Agent can only collect metrics from a single ElasticSearch server.
However, the exporter can collect the metrics from all nodes through that server configured.
{{% /admonition %}}

We strongly recommend that you configure a separate user for the Agent, and give it only the strictly mandatory
security privileges necessary for monitoring your node, as per the [official documentation](https://github.com/prometheus-community/elasticsearch_exporter#elasticsearch-7x-security-privileges).

## Usage

```river
prometheus.exporter.elasticsearch "LABEL" {
    address = "ELASTICSEARCH_ADDRESS"
}
```

## Arguments

You can use the following arguments to configure the exporter's behavior.
Omitted fields take their default values.

| Name                   | Type       | Description                                                                                            | Default                   | Required |
| ---------------------- | ---------- | ------------------------------------------------------------------------------------------------------ | ------------------------- | -------- |
| `address`              | `string`   | HTTP API address of an Elasticsearch node.                                                             | `"http://localhost:9200"` | no       |
| `timeout`              | `duration` | Timeout for trying to get stats from Elasticsearch.                                                    | `"5s"`                    | no       |
| `all`                  | `bool`     | Export stats for all nodes in the cluster. If used, this flag will override the flag `node`.           |                           | no       |
| `node`                 | `string`   | Node's name of which metrics should be exposed                                                         |                           | no       |
| `indices`              | bool       | Export stats for indices in the cluster.                                                               |                           | no       |
| `indices_settings`     | bool       | Export stats for settings of all indices of the cluster.                                               |                           | no       |
| `cluster_settings`     | bool       | Export stats for cluster settings.                                                                     |                           | no       |
| `shards`               | bool       | Export stats for shards in the cluster (implies indices).                                              |                           | no       |
| `snapshots`            | bool       | Export stats for the cluster snapshots.                                                                |                           | no       |
| `clusterinfo_interval` | `duration` | Cluster info update interval for the cluster label.                                                    | `"5m"`                    | no       |
| `ca`                   | `string`   | Path to PEM file that contains trusted Certificate Authorities for the Elasticsearch connection.       |                           | no       |
| `client_private_key`   | `string`   | Path to PEM file that contains the private key for client auth when connecting to Elasticsearch.       |                           | no       |
| `client_cert`          | `string`   | Path to PEM file that contains the corresponding cert for the private key to connect to Elasticsearch. |                           | no       |
| `ssl_skip_verify`      | `bool`     | Skip SSL verification when connecting to Elasticsearch.                                                |                           | no       |
| `aliases`              | `bool`     | Include informational aliases metrics.                                                                 |                           | no       |
| `data_streams`         | `bool`     | Export stats for Data Streams.                                                                         |                           | no       |
| `slm`                  | `bool`     | Export stats for SLM (Snapshot Lifecycle Management).                                                  |                           | no       |

## Exported fields

{{< docs/shared lookup="flow/reference/components/exporter-component-exports.md" source="agent" version="<AGENT VERSION>" >}}

## Component health

`prometheus.exporter.elasticsearch` is only reported as unhealthy if given
an invalid configuration. In those cases, exported fields retain their last
healthy values.

## Debug information

`prometheus.exporter.elasticsearch` does not expose any component-specific
debug information.

## Debug metrics

`prometheus.exporter.elasticsearch` does not expose any component-specific
debug metrics.

## Example

This example uses a [`prometheus.scrape` component][scrape] to collect metrics
from `prometheus.exporter.elasticsearch`:

```river
prometheus.exporter.elasticsearch "example" {
  address = "http://localhost:9200"
}

// Configure a prometheus.scrape component to collect Elasticsearch metrics.
prometheus.scrape "demo" {
  targets    = prometheus.exporter.elasticsearch.example.targets
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
