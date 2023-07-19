---
canonical: https://grafana.com/docs/agent/latest/flow/reference/components/prometheus.exporter.snowflake/
title: prometheus.exporter.snowflake
---

# prometheus.exporter.snowflake
The `prometheus.exporter.snowflake` component embeds
[snowflake_exporter](https://github.com/grafana/snowflake-prometheus-exporter) for collecting warehouse, database, table, and replication statistics from a Snowflake account via HTTP for Prometheus consumption.

## Usage

```river
prometheus.exporter.snowflake "LABEL" {
    account_name = ACCOUNT_NAME
    username =     USERNAME
    password =     PASSWORD
    warehouse =    WAREHOUSE
}
```

## Arguments

The following arguments can be used to configure the exporter's behavior.
Omitted fields take their default values.

| Name           | Type     | Description                                           | Default          | Required |
|----------------|----------|-------------------------------------------------------|------------------|----------|
| `account_name` | `string` | The account to collect metrics for.                   |                  | yes      |
| `username`     | `string` | The username for the user used when querying metrics. |                  | yes      |
| `password`     | `secret` | The password for the user used when querying metrics. |                  | yes      |
| `role`         | `string` | The role to use when querying metrics.                | `"ACCOUNTADMIN"` | no       |
| `warehouse`    | `string` | The warehouse to use when querying metrics.           |                  | yes      |

## Blocks

The `prometheus.exporter.snowflake` component does not support any blocks, and is configured
fully through arguments.

## Exported fields

The following fields are exported and can be referenced by other components.

| Name      | Type                | Description                                                  |
|-----------|---------------------|--------------------------------------------------------------|
| `targets` | `list(map(string))` | The targets that can be used to collect `snowflake` metrics. |

For example, the `targets` can either be passed to a `prometheus.relabel`
component to rewrite the metric's label set, or to a `prometheus.scrape`
component that collects the exposed metrics.

The exported targets will use the configured [in-memory traffic][] address
specified by the [run command][].

[in-memory traffic]: {{< relref "../../concepts/component_controller.md#in-memory-traffic" >}}
[run command]: {{< relref "../cli/run.md" >}}

## Component health

`prometheus.exporter.snowflake` is only reported as unhealthy if given
an invalid configuration. In those cases, exported fields retain their last
healthy values.

## Debug information

`prometheus.exporter.snowflake` does not expose any component-specific
debug information.

## Debug metrics

`prometheus.exporter.snowflake` does not expose any component-specific
debug metrics.

## Example

This example uses a [`prometheus.scrape` component][scrape] to collect metrics
from `prometheus.exporter.snowflake`:

```river
prometheus.exporter.snowflake "example" {
  account_name = "XXXXXXX-YYYYYYY"
  username     = "grafana"
  password     = "snowflake"
  warehouse    = "examples"
}

// Configure a prometheus.scrape component to collect snowflake metrics.
prometheus.scrape "demo" {
  targets    = prometheus.exporter.snowflake.example.targets
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
