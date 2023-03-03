---
# NOTE(rfratto): the title below has zero-width spaces injected into it to
# prevent it from overflowing the sidebar on the rendered site. Be careful when
# modifying this section to retain the spaces.
#
# Ideally, in the future, we can fix the overflow issue with css rather than
# injecting special characters.

title: prometheus.exporter.postgres
labels:
  stage: beta
---

# prometheus.exporter.postgres
The `prometheus.exporter.postgres` component embeds
[postgres_exporter](https://github.com/prometheus-community/postgres_exporter) for collecting metrics from a postgres database.

Multiple `prometheus.exporter.postgres` components can be specified by giving them different 
labels.

## Usage

```river
prometheus.exporter.postgres "LABEL" {
    data_source_names = DATA_SOURCE_NAMES_LIST
}
```

## Arguments
The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`data_source_names`                  | `list(secret)`      | Specifies the Postgres server(s) to connect to.  |         | yes
`disable_settings_metrics`           | `bool`              | Disables collection of metrics from pg_settings. | `false` | no
`disable_default_metrics`            | `bool`              | When `true`, only exposes metrics supplied from `custom_queries_path`. | `false` | no
`autodiscover_databases`             | `bool`              | Whether to automatically discover databases to scrape metrics from. | `false` | no
`exclude_databases`                  | `list(string)`      | A list of databases to ignore when `autodiscover_databases` is `true` | | no
`include_databases`                  | `list(string)`      | Includes only specific databases (excluding all others) when autodiscover_databases is `true` | | no
`custom_queries_path`                | `string`            | Path to YAML file containing custom queries to expose as metrics. | | no

The format for connection strings in `data_source_names` can be found in the [official postgresql documentation](https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING).

See examples for the `custom_queries_path` file in the [postgres_exporter repository](https://github.com/prometheus-community/postgres_exporter/blob/master/queries.yaml).

**NOTE**: There are a number of environment variables that are not recommended for use, as they will affect _all_ `prometheus.exporter.postgres` components. A full list can be found in the [postgres_exporter repository](https://github.com/prometheus-community/postgres_exporter#environment-variables).

## Blocks
The `prometheus.exporter.postgres` component does not support any blocks, and is configured fully through arguments.

## Exported fields
The following fields are exported and can be referenced by other components:

Name      | Type                | Description
--------- | ------------------- | -----------
`targets` | `list(map(string))` | The targets that can be used to collect `postgres` metrics.

For example, `targets` can either be passed to a `prometheus.relabel`
component to rewrite the metrics' label set, or to a `prometheus.scrape`
component that collects the exposed metrics.

## Component health

`prometheus.exporter.postgres` is only reported as unhealthy if given
an invalid configuration.

## Debug information

`prometheus.exporter.postgres` does not expose any component-specific
debug information.

## Debug metrics

`prometheus.exporter.postgres` does not expose any component-specific
debug metrics.

## Examples

This minimal example uses a `prometheus.exporter.postgres` component to collect metrics from a Postgres
server running locally with all default settings:

```river
prometheus.exporter.postgres "example" {
  data_source_names = ["postgresql://username:password@localhost:5432/database_name?sslmode=disable"]
}

prometheus.scrape "default" {
  targets    = prometheus.exporter.postgres.example.targets
  forward_to = [prometheus.remote_write.default.receiver]
}

prometheus.remote_write "default" {
  client {
    url = env("PROMETHEUS_URL")
  }
}
```

This example uses a `prometheus.exporter.postgres` component to collect custom queries from a set of
specific databases, without any of the default metrics:

```river
prometheus.exporter.postgres "example" {
  data_source_names       = ["postgresql://username:password@localhost:5432/database_name?sslmode=disable"]
  
  disable_default_metrics = true
  autodiscover_databases  = true

  include_databases       = ["payments", "users"]
  custom_queries_path     = "/etc/agent/custom-postgres-metrics.yaml"
}

prometheus.scrape "default" {
  targets    = prometheus.exporter.postgres.example.targets
  forward_to = [prometheus.remote_write.default.receiver]
}

prometheus.remote_write "default" {
  client {
    url = env("PROMETHEUS_URL")
  }
}
```

[scrape]: {{< relref "./prometheus.scrape.md" >}}
