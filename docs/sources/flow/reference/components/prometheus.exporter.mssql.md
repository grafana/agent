---
aliases:
- /docs/grafana-cloud/agent/flow/reference/components/prometheus.exporter.mssql/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/reference/components/prometheus.exporter.mssql/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/reference/components/prometheus.exporter.mssql/
canonical: https://grafana.com/docs/agent/latest/flow/reference/components/prometheus.exporter.mssql/
title: prometheus.exporter.mssql
description: Learn about prometheus.exporter.mssql
---

# prometheus.exporter.mssql

The `prometheus.exporter.mssql` component embeds
[sql_exporter](https://github.com/burningalchemist/sql_exporter) for collecting stats from a Microsoft SQL Server.

## Usage

```river
prometheus.exporter.mssql "LABEL" {
    connection_string = CONNECTION_STRING
}
```

## Arguments

The following arguments can be used to configure the exporter's behavior.
Omitted fields take their default values.

| Name                   | Type       | Description                                                       | Default | Required |
| ---------------------- | ---------- | ----------------------------------------------------------------- | ------- | -------- |
| `connection_string`    | `secret`   | The connection string used to connect to an Microsoft SQL Server. |         | yes      |
| `max_idle_connections` | `int`      | Maximum number of idle connections to any one target.             | `3`     | no       |
| `max_open_connections` | `int`      | Maximum number of open connections to any one target.             | `3`     | no       |
| `timeout`              | `duration` | The query timeout in seconds.                                     | `"10s"` | no       |

[The sql_exporter examples](https://github.com/burningalchemist/sql_exporter/blob/master/examples/azure-sql-mi/sql_exporter.yml#L21) show the format of the `connection_string` argument:

```conn
sqlserver://USERNAME_HERE:PASSWORD_HERE@SQLMI_HERE_ENDPOINT.database.windows.net:1433?encrypt=true&hostNameInCertificate=%2A.SQL_MI_DOMAIN_HERE.database.windows.net&trustservercertificate=true
```

## Blocks

The `prometheus.exporter.mssql` component does not support any blocks, and is configured
fully through arguments.

## Exported fields

{{< docs/shared lookup="flow/reference/components/exporter-component-exports.md" source="agent" version="<AGENT VERSION>" >}}

## Component health

`prometheus.exporter.mssql` is only reported as unhealthy if given
an invalid configuration. In those cases, exported fields retain their last
healthy values.

## Debug information

`prometheus.exporter.mssql` does not expose any component-specific
debug information.

## Debug metrics

`prometheus.exporter.mssql` does not expose any component-specific
debug metrics.

## Example

This example uses a [`prometheus.scrape` component][scrape] to collect metrics
from `prometheus.exporter.mssql`:

```river
prometheus.exporter.mssql "example" {
  connection_string = "sqlserver://user:pass@localhost:1433"
}

// Configure a prometheus.scrape component to collect mssql metrics.
prometheus.scrape "demo" {
  targets    = prometheus.exporter.mssql.example.targets
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
