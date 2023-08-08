---
canonical: https://grafana.com/docs/agent/latest/flow/reference/components/prometheus.exporter.mongodb/
title: prometheus.exporter.mongodb
---

# prometheus.exporter.mongodb
The `prometheus.exporter.mongodb` component embeds percona's [`mongodb_exporter`](https://github.com/percona/mongodb_exporter).

{{% admonition type="note" %}}
For this integration to work properly, you must have connect each node of your MongoDB cluster to an agent instance.
That's because this exporter does not collect metrics from multiple nodes.
{{% /admonition %}}

We strongly recommend configuring a separate user for the Grafana Agent, giving it only the strictly mandatory security privileges necessary for monitoring your node. 
Refer to the [Percona documentation](https://github.com/percona/mongodb_exporter#permissions) for more information.

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
`direct-connect` | `boolean` | Whether or not a direct connect should be made. Direct connections are not valid if multiple hosts are specified or an SRV URI is used. | true | No
`discovering-mode` | `boolean` | Wheter or not to enable autodiscover collections. | false | No
`tls-basic-auth-config-path` | `string` | Path to the file having Prometheus TLS config for basic auth. Only enable if you want to use TLS based authentication. Check [`tls_config`](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#tls_config) for reference on the file format | | No

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

prometheus.remote_write "default" {
  endpoint {
    url = "REMOTE_WRITE_URL"
  }
}
```

[scrape]: {{< relref "./prometheus.scrape.md" >}}
