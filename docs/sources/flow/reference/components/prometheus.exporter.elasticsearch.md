---
# NOTE(rfratto): the title below has zero-width spaces injected into it to
# prevent it from overflowing the sidebar on the rendered site. Be careful when
# modifying this section to retain the spaces.
#
# Ideally, in the future, we can fix the overflow issue with css rather than
# injecting special characters.

title: prometheus.exporter.elasticsearch
---

# prometheus.exporter.elasticsearch
The `prometheus.exporter.elasticsearch` component embeds
[elasticsearch_exporter](https://github.com/prometheus-community/elasticsearch_exporter) for
the collection of metrics from ElasticSearch servers.

Note that currently, an Agent can only collect metrics from a single ElasticSearch server.
However, the exporter is able to collect the metrics from all nodes through that server configured.

We strongly recommend that you configure a separate user for the Agent, and give it only the strictly mandatory
security privileges necessary for monitoring your node, as per the [official documentation](https://github.com/prometheus-community/elasticsearch_exporter#elasticsearch-7x-security-privileges).

## Usage

```river
prometheus.exporter.elasticsearch "LABEL" {
    address = "ELASTICSEARCH_ADDRESS"
}
```

## Arguments
The following arguments can be used to configure the exporter's behavior.
Omitted fields take their default values.

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`address` | `string` | HTTP API address of an Elasticsearch node. | `"http://localhost:9200"` | no
`timeout` | `duration` | Timeout for trying to get stats from Elasticsearch. | `"5s"` | no
`all` | `bool` | Export stats for all nodes in the cluster. If used, this flag will override the flag `node`. | | no
`node` | `string` | Node's name of which metrics should be exposed |  | no
`indices` | bool | Export stats for indices in the cluster. |  | no
`indices_settings` | bool | Export stats for settings of all indices of the cluster. |  | no
`cluster_settings` | bool | Export stats for cluster settings. |  | no
`shards` | bool | Export stats for shards in the cluster (implies indices). |  | no
`snapshots` | bool | Export stats for the cluster snapshots. |  | no
`clusterinfo_interval` | `duration` | Cluster info update interval for the cluster label. | `"5m"` | no
`ca` | `string` | Path to PEM file that contains trusted Certificate Authorities for the Elasticsearch connection. |  | no
`client_private_key` | `string` | Path to PEM file that contains the private key for client auth when connecting to Elasticsearch. |  | no
`client_cert` | `string` | Path to PEM file that contains the corresponding cert for the private key to connect to Elasticsearch. |  | no
`ssl_skip_verify` | `bool` | Skip SSL verification when connecting to Elasticsearch. | | no
`aliases` | `bool` | Include informational aliases metrics. |  | no
`data_streams` | `bool` | Export stats for Data Streams. |  | no
`slm` | `bool` | Export stats for SLM (Snapshot Lifecycle Management). |  | no



## Exported fields
The following fields are exported and can be referenced by other components.

Name      | Type                | Description
--------- | ------------------- | -----------
`targets` | `list(map(string))` | The targets that can be used to collect `elasticsearch` metrics.

For example, `targets` can either be passed to a `prometheus.relabel`
component to rewrite the metrics' label set, or to a `prometheus.scrape`
component that collects the exposed metrics.

The exported targets will use the configured [in-memory traffic][] address
specified by the [run command][].

[in-memory traffic]: {{< relref "../../concepts/component_controller.md#in-memory-traffic" >}}
[run command]: {{< relref "../cli/run.md" >}}

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
  address = "localhost:9200"
}

// Configure a prometheus.scrape component to collect Elasticsearch metrics.
prometheus.scrape "demo" {
  targets    = prometheus.exporter.elasticsearch.example.targets
  forward_to = [ /* ... */ ]
}
```

[scrape]: {{< relref "./prometheus.scrape.md" >}}
