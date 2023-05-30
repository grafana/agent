---
title: prometheus.exporter.cadvisor
---

# prometheus.exporter.cadvisor
The `prometheus.exporter.cadvisor` component collects container metrics using
[cAdvisor](https://github.com/google/cadvisor).

## Usage

```river
prometheus.exporter.cadvisor "LABEL" {
}
```

## Arguments
The following arguments can be used to configure the exporter's behavior.
All arguments are optional. Omitted fields take their default values.

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`store_container_labels` | `bool` | Whether to store all container labels and environment variables as Prometheus labels. | `true` | no
`allowlisted_container_labels` | `list(string)` | Allowlist of container labels to convert to Prometheus labels. | | no
`env_metadata_allowlist` | `list(string)` | Allowlist of environment variables to convert to Prometheus labels. | | no

## Exported fields
The following fields are exported and can be referenced by other components.

Name      | Type                | Description
--------- | ------------------- | -----------
`targets` | `list(map(string))` | The targets that can be used to collect `cadvisor` metrics.

For example, the `targets` could either be passed to a `prometheus.relabel`
component to rewrite the metrics' label set, or to a `prometheus.scrape`
component that collects the exposed metrics.

The exported targets will use the configured [in-memory traffic][] address
specified by the [run command][].

[in-memory traffic]: {{< relref "../../concepts/component_controller.md#in-memory-traffic" >}}
[run command]: {{< relref "../cli/run.md" >}}

## Component health

`prometheus.exporter.cadvisor` is only reported as unhealthy if given
an invalid configuration. In those cases, exported fields retain their last
healthy values.

## Debug information

`prometheus.exporter.cadvisor` does not expose any component-specific
debug information.

## Debug metrics

`prometheus.exporter.cadvisor` does not expose any component-specific
debug metrics.

## Example

This example uses a [`prometheus.scrape` component][scrape] to collect metrics
from `prometheus.exporter.cadvisor`:

```river
prometheus.exporter.cadvisor "example" {
  docker     = "unix:///var/run/docker.sock"
  docker_tls = false

  storage_duration = "5m"
}

// Configure a prometheus.scrape component to collect cadvisor metrics.
prometheus.scrape "scraper" {
  targets    = prometheus.exporter.cadvisor.example.targets
  forward_to = [ prometheus.remote_write.demo.receiver ]
}

prometheus.remote_write "demo" {
  endpoint {
    url = "http://mimir:9009/api/v1/push"

    basic_auth {
      username = "example-user"
      password = "example-password"
    }
  }
}
```

[scrape]: {{< relref "./prometheus.scrape.md" >}}
