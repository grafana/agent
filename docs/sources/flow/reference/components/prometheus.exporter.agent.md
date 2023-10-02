---
canonical: https://grafana.com/docs/agent/latest/flow/reference/components/prometheus.exporter.agent/
title: prometheus.exporter.agent
description: Learn about prometheus.exporter.agen
---

# prometheus.exporter.agent
The `prometheus.exporter.agent` component collects and exposes metrics about the agent itself.

## Usage

```river
prometheus.exporter.agent "agent" {
}
```

## Arguments
`prometheus.exporter.agent` accepts no arguments.

## Exported fields

{{< docs/shared lookup="flow/reference/components/exporter-component-exports.md" source="agent" version="<AGENT VERSION>" >}}

## Component health

`prometheus.exporter.agent` is only reported as unhealthy if given
an invalid configuration.

## Debug information

`prometheus.exporter.agent` does not expose any component-specific
debug information.

## Debug metrics

`prometheus.exporter.agent` does not expose any component-specific
debug metrics.

## Example

This example uses a [`prometheus.scrape` component][scrape] to collect metrics
from `prometheus.exporter.agent`:

```river
prometheus.exporter.agent "agent" {}

// Configure a prometheus.scrape component to collect agent metrics.
prometheus.scrape "demo" {
  targets    = prometheus.exporter.agent.example.targets
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