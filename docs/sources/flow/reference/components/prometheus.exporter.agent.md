---
aliases:
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/reference/components/prometheus.exporter.agent/
- /docs/grafana-cloud/send-data/agent/flow/reference/components/prometheus.exporter.agent/
canonical: https://grafana.com/docs/agent/latest/flow/reference/components/prometheus.exporter.agent/
description: Learn about prometheus.exporter.agen
title: prometheus.exporter.agent
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

<!-- START GENERATED COMPATIBLE COMPONENTS -->

## Compatible components

`prometheus.exporter.agent` can output data to the following components:

- Components that accept Targets:
  - [`discovery.relabel`]({{< relref "../components/discovery.relabel.md" >}})
  - [`local.file_match`]({{< relref "../components/local.file_match.md" >}})
  - [`loki.source.docker`]({{< relref "../components/loki.source.docker.md" >}})
  - [`loki.source.file`]({{< relref "../components/loki.source.file.md" >}})
  - [`loki.source.kubernetes`]({{< relref "../components/loki.source.kubernetes.md" >}})
  - [`otelcol.processor.discovery`]({{< relref "../components/otelcol.processor.discovery.md" >}})
  - [`prometheus.scrape`]({{< relref "../components/prometheus.scrape.md" >}})
  - [`pyroscope.scrape`]({{< relref "../components/pyroscope.scrape.md" >}})

Note that connecting some components may not be feasible or components may require further configuration to make the connection work correctly. Please refer to the linked documentation for more details.

<!-- END GENERATED COMPATIBLE COMPONENTS -->
