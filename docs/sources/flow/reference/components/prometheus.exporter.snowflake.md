---
# NOTE(rfratto): the title below has zero-width spaces injected into it to
# prevent it from overflowing the sidebar on the rendered site. Be careful when
# modifying this section to retain the spaces.
#
# Ideally, in the future, we can fix the overflow issue with css rather than
# injecting special characters.

title: prometheus.exporter.snowflake
---

# prometheus.exporter.snowflake
The `prometheus.exporter.snowflake` component embeds
[snowflake_exporter](https://github.com/grafana/snowflake-prometheus-exporter) for collecting warehouse, database, table, and replication statistics from a Snowflake account via HTTP for Prometheus consumption.

## Usage

```river
prometheus.exporter.snowflake "LABEL" {
    account_name = "ACCOUNT_NAME"
    username =     "USERNAME"
    password =     "PASSWORD"
    warehouse =    "WAREHOUSE"
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

{{< docs/shared lookup="flow/reference/components/exporter-component-exports.md" source="agent" >}}

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
  username     = "USERNAME"
  password     = "PASSWORD"
  warehouse    = "WAREHOUSE"
}

// Configure a prometheus.scrape component to collect snowflake metrics.
prometheus.scrape "demo" {
  targets    = prometheus.exporter.snowflake.example.targets
  forward_to = [ prometheus.remote_write.default.receiver ]
}

prometheus.remote_write "default" {
  endpoint {
    url = "REMOTE_WRITE_URL"
  }
}
```

[scrape]: {{< relref "./prometheus.scrape.md" >}}
