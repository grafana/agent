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

{{< docs/shared lookup="flow/reference/components/exporter-component-exports.md" source="agent" >}}

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

## Examples
This minimal example uses a `prometheus.exporter.memcached` component to collect metrics from a Memcached
server running locally, and scrapes the metrics using a [prometheus.scrape][scrape] component:

```river
prometheus.exporter.memcached "example" {
    address = "localhost:13321"
    timeout = "5s"
}

prometheus.scrape "example" {
    targets    = [prometheus.exporter.memcached.example.targets]
    forward_to = [prometheus.remote_write.default.receiver]
}

prometheus.remote_write "default" {
  endpoint {
    url = "prometheus.example.com/api/v1/write"

    basic_auth {
      username = "user"
      password = "pass"
    }
  }
}
```

[scrape]: {{< relref "./prometheus.scrape.md" >}}
