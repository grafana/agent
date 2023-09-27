---
aliases:
- /docs/grafana-cloud/agent/flow/reference/components/prometheus.exporter.memcached/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/reference/components/prometheus.exporter.memcached/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/reference/components/prometheus.exporter.memcached/
canonical: https://grafana.com/docs/agent/latest/flow/reference/components/prometheus.exporter.memcached/
title: prometheus.exporter.memcached
description: Learn about prometheus.exporter.memcached
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

| Name      | Type       | Description                                         | Default             | Required |
| --------- | ---------- | --------------------------------------------------- | ------------------- | -------- |
| `address` | `string`   | The Memcached server address.                       | `"localhost:11211"` | no       |
| `timeout` | `duration` | The timeout for connecting to the Memcached server. | `"1s"`              | no       |

## Blocks

The following blocks are supported inside the definition of `prometheus.exporter.memcached`:

| Hierarchy  | Block          | Description                                             | Required |
| ---------- | -------------- | ------------------------------------------------------- | -------- |
| tls_config | [tls_config][] | TLS configuration for requests to the Memcached server. | no       |

[tls_config]: #tls_config-block

### tls_config block

{{< docs/shared lookup="flow/reference/components/tls-config-block.md" source="agent" version="<AGENT VERSION>" >}}

## Exported fields

{{< docs/shared lookup="flow/reference/components/exporter-component-exports.md" source="agent" version="<AGENT VERSION>" >}}

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
