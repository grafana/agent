---
# NOTE(rfratto): the title below has zero-width spaces injected into it to
# prevent it from overflowing the sidebar on the rendered site. Be careful when
# modifying this section to retain the spaces.
#
# Ideally, in the future, we can fix the overflow issue with css rather than
# injecting special characters.

title: prometheus.exporter.mongodb
---

# prometheus.exporter.mongodb
The `prometheus.exporter.mongodb` component embeds percona's [`mongodb_exporter`](https://github.com/percona/mongodb_exporter).

{{% admonition type="note" %}}
For this integration to work properly, you must have connect each node of your MongoDB cluster to an agent instance.
That's because this exporter does not collect metrics from multiple nodes.

Additionally, you need to define two custom label for your metrics using `relabel_configs`.

The first one is `service_name`, which is how you identify this node in your cluster (example: `ReplicaSet1-Node1`).

The second one is `mongodb_cluster`, which is the name of your mongodb cluster, and must be set the same value for all nodes composing the cluster (example: `prod-cluster`).
{{% /admonition %}}

We strongly recommend that you configure a separate user for the Agent, and give it only the strictly mandatory
security privileges necessary for monitoring your node, as per the [official documentation](https://github.com/percona/mongodb_exporter#permissions).

## Usage

```river
prometheus.exporter.mongodb "LABEL" {
    mongodb_uri = "MONGODB_URI"
}
```

## Arguments
You can use the following arguments to configure the exporter's behavior.
Omitted fields take their default values.

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`mongodb_uri` | `string` | MongoDB node connection URI. | | Yes

MongoDB node connection URI must be in the [`Standard Connection String Format`](https://docs.mongodb.com/manual/reference/connection-string/#std-label-connections-standard-connection-string-format)



## Exported fields
The following fields are exported and can be referenced by other components.

Name      | Type                | Description
--------- | ------------------- | -----------
`targets` | `list(map(string))` | The targets that can be used to collect `mongodb` metrics.

For example, `targets` can either be passed to a `prometheus.relabel`
component to rewrite the metrics' label set, or to a `prometheus.scrape`
component that collects the exposed metrics.

The exported targets will use the configured [in-memory traffic][] address
specified by the [run command][].

[in-memory traffic]: {{< relref "../../concepts/component_controller.md#in-memory-traffic" >}}
[run command]: {{< relref "../cli/run.md" >}}

## Component health

`prometheus.exporter.mongodb` is only reported as unhealthy if given
an invalid configuration. In those cases, exported fields retain their last
healthy values.

## Debug information

`prometheus.exporter.mongodb` does not expose any component-specific
debug information.

## Debug metrics

`prometheus.exporter.mongodb` does not expose any component-specific
debug metrics.

## Example

This example uses a [`prometheus.scrape` component][scrape] to collect metrics
from `prometheus.exporter.mongodb`:

```river
prometheus.exporter.mongodb "example" {
  mongodb_uri = "mongodb://127.0.0.1:27017"
}

// Configure a prometheus.scrape component to collect MongoDB metrics.
prometheus.scrape "demo" {
  targets    = prometheus.exporter.mongodb.example.targets
  forward_to = [ prometheus.remote_write.default.receiver ]
}

// Configure a prometheus.relabel component to rewrite the metrics' label set.
prometheus.relabel "mongodb_relabel" {
  forward_to = [prometheus.remote_write.default.receiver]

  rule {
    source_labels = ["__address__"]
    target_label  = "service_name"
    replacement   = "ReplicaSet1-Node1"
  }
  rule {
    source_labels = ["__address__"]
    target_label  = "mongodb_cluster"
    replacement   = "prod-cluster"
  }
}

prometheus.remote_write "default" {
  endpoint {
    url = "REMOTE_WRITE_URL"
  }
}
```

[scrape]: {{< relref "./prometheus.scrape.md" >}}
