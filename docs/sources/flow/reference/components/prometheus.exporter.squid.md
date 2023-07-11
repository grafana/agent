---
# NOTE(rfratto): the title below has zero-width spaces injected into it to
# prevent it from overflowing the sidebar on the rendered site. Be careful when
# modifying this section to retain the spaces.
#
# Ideally, in the future, we can fix the overflow issue with css rather than
# injecting special characters.

title: prometheus.exporter.squid
---

# prometheus.exporter.squid
The `prometheus.exporter.squid` component embeds
[squid_exporter](https://github.com/boynux/squid-exporter) for collecting metrics from a squid instance.

## Usage

```river
prometheus.exporter.squid "LABEL" {
    address = SQUID_ADDRESS
}
```

## Arguments

You can use the following arguments to configure the exporter's behavior.
Omitted fields take their default values.

| Name           | Type     | Description                                           | Default          | Required |
|----------------|----------|-------------------------------------------------------|------------------|----------|
| `address`      | `string` | The squid address to collect metrics from.            |                  | yes      |
| `username`     | `string` | The username for the user used when querying metrics. |                  | no       |
| `password`     | `secret` | The password for the user used when querying metrics. |                  | no       |


## Blocks

The `prometheus.exporter.squid` component does not support any blocks, and is configured
fully through arguments.

## Exported fields

The following fields are exported and can be referenced by other components.

| Name      | Type                | Description                                                  |
|-----------|---------------------|--------------------------------------------------------------|
| `targets` | `list(map(string))` | The targets that can be used to collect `squid` metrics. |

For example, the `targets` can either be passed to a `prometheus.relabel`
component to rewrite the metric's label set, or to a `prometheus.scrape`
component that collects the exposed metrics.

The exported targets will use the configured [in-memory traffic][] address
specified by the [run command][].

[in-memory traffic]: {{< relref "../../concepts/component_controller.md#in-memory-traffic" >}}
[run command]: {{< relref "../cli/run.md" >}}

## Component health

`prometheus.exporter.squid` is only reported as unhealthy if given
an invalid configuration. In those cases, exported fields retain their last
healthy values.

## Debug information

`prometheus.exporter.squid` does not expose any component-specific
debug information.

## Debug metrics

`prometheus.exporter.squid` does not expose any component-specific
debug metrics.

## Example

This example uses a [`prometheus.scrape` component][scrape] to collect metrics
from `prometheus.exporter.squid`:

```river
prometheus.exporter.squid "example" {
  address = "localhost:3128"
}

// Configure a prometheus.scrape component to collect squid metrics.
prometheus.scrape "demo" {
  targets    = prometheus.exporter.squid.example.targets
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
