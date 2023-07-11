---
# NOTE(rfratto): the title below has zero-width spaces injected into it to
# prevent it from overflowing the sidebar on the rendered site. Be careful when
# modifying this section to retain the spaces.
#
# Ideally, in the future, we can fix the overflow issue with css rather than
# injecting special characters.

title: prometheus.exporter.memcached
---

# prometheus.exporter.memcached
The `prometheus.exporter.memcached` component embeds
[memcached_exporter](https://github.com/prometheus/memcached_exporter) for collecting metrics from a Memcached server.

## Usage
```river
prometheus.exporter.memcached "LABEL" {
}
```

## Arguments
The following arguments are supported:

Name             | Type       | Description                                         | Default               | Required |
---------------- | ---------- | --------------------------------------------------- | --------------------- | -------- |
`address`        | `string`   | The Memcached server address.                       | `"localhost:11211"`   | no       |
`timeout`        | `duration` | The timeout for connecting to the Memcached server. | `"1s"`                | no       |

## Blocks
The `prometheus.exporter.memcached` component does not support any blocks, and is configured
fully through arguments.

## Exported fields
The following fields are exported and can be referenced by other components:

Name      | Type                | Description
--------- | ------------------- | -----------
`targets` | `list(map(string))` | The targets that can be used to collect `memcached` metrics.

For example, `targets` can either be passed to a `prometheus.relabel`
component to rewrite the metrics' label set, or to a `prometheus.scrape`
component that collects the exposed metrics.

The exported targets will use the configured [in-memory traffic][] address
specified by the [run command][].

[in-memory traffic]: {{< relref "../../concepts/component_controller.md#in-memory-traffic" >}}
[run command]: {{< relref "../cli/run.md" >}}

## Component health
`prometheus.exporter.memcached` is only reported as unhealthy if given
an invalid configuration. In those cases, exported fields retain their last
healthy values.

## Debug information
`prometheus.exporter.memcached` does not expose any component-specific
debug information.

## Debug metrics
`prometheus.exporter.memcached` does not expose any component-specific
debug metrics.

## Example

This example uses a `prometheus.exporter.memcached` component to collect metrics from a Memcached
server running locally, and scrapes the metrics using a [prometheus.scrape][scrape] component:

```river
prometheus.exporter.memcached "example" {
  address = "localhost:13321"
  timeout = "5s"
}

prometheus.scrape "example" {
  targets    = [prometheus.exporter.memcached.example.targets]
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
