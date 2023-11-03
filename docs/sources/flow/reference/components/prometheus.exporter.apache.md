---
aliases:
- /docs/grafana-cloud/agent/flow/reference/components/prometheus.exporter.apache/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/reference/components/prometheus.exporter.apache/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/reference/components/prometheus.exporter.apache/
canonical: https://grafana.com/docs/agent/latest/flow/reference/components/prometheus.exporter.apache/
title: prometheus.exporter.apache
description: Learn about prometheus.exporter.apache
---

# prometheus.exporter.apache

The `prometheus.exporter.apache` component embeds
[apache_exporter](https://github.com/Lusitaniae/apache_exporter) for collecting mod_status statistics from an apache server.

## Usage

```river
prometheus.exporter.apache "LABEL" {
}
```

## Arguments

The following arguments can be used to configure the exporter's behavior.
All arguments are optional. Omitted fields take their default values.

| Name            | Type     | Description                               | Default                               | Required |
| --------------- | -------- | ----------------------------------------- | ------------------------------------- | -------- |
| `scrape_uri`    | `string` | URI to Apache stub status page.           | `http://localhost/server-status?auto` | no       |
| `host_override` | `string` | Override for HTTP Host header.            |                                       | no       |
| `insecure`      | `bool`   | Ignore server certificate if using https. | false                                 | no       |

## Exported fields

{{< docs/shared lookup="flow/reference/components/exporter-component-exports.md" source="agent" version="<AGENT VERSION>" >}}

## Component health

`prometheus.exporter.apache` is only reported as unhealthy if given
an invalid configuration. In those cases, exported fields retain their last
healthy values.

## Debug information

`prometheus.exporter.apache` does not expose any component-specific
debug information.

## Debug metrics

`prometheus.exporter.apache` does not expose any component-specific
debug metrics.

## Example

This example uses a [`prometheus.scrape` component][scrape] to collect metrics
from `prometheus.exporter.apache`:

```river
prometheus.exporter.apache "example" {
  scrape_uri = "http://web.example.com/server-status?auto"
}

// Configure a prometheus.scrape component to collect apache metrics.
prometheus.scrape "demo" {
  targets    = prometheus.exporter.apache.example.targets
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
