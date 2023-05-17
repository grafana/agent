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
`disable_default_metrics`            | `bool`              | When `true`, only exposes metrics supplied from `custom_queries_config_path`. | `false` | no
`custom_queries_config_path`         | `string`            | Path to YAML file containing custom queries to expose as metrics. | "" | no

The format for connection strings in `data_source_names` can be found in the [official postgresql documentation](https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING).

See examples for the `custom_queries_config_path` file in the [postgres_exporter repository](https://github.com/prometheus-community/postgres_exporter/blob/master/queries.yaml).

**NOTE**: There are a number of environment variables that are not recommended for use, as they will affect _all_ `prometheus.exporter.postgres` components. A full list can be found in the [postgres_exporter repository](https://github.com/prometheus-community/postgres_exporter#environment-variables).

## Blocks
The following blocks are supported:

Hierarchy                    | Block             | Description                         | Required |
---------------------------- | ----------------- | ----------------------------------- | -------- |
autodiscovery                | [autodiscovery][] | Database discovery settings.        | no       |

[autodiscovery]: #autodiscovery-block

### autodiscovery block
The `autodiscovery` block configures discovery of databases, outside of any specified in `data_source_names`.

The following arguments are supported:

Name                 | Type           | Description                                                   | Default | Required |
-------------------- | -------------- | ------------------------------------------------------------- | ------- | -------- |
`enabled`            | `bool`         | Whether to autodiscover other databases                       | `false` | no       |
`database_allowlist` | `list(string)` | List of databases to filter for, meaning only these databases will be scraped. | | no
`database_denylist`  | `list(string)` | List of databases to filter out, meaning all other databases will be scraped. | | no

If `enabled` is set to `true` and no allowlist or denylist is specified, the exporter will scrape from all databases.

If `autodiscovery` is disabled, neither `database_allowlist` nor `database_denylist` will have any effect.

## Exported fields
The following fields are exported and can be referenced by other components:

Name      | Type                | Description
--------- | ------------------- | -----------
`targets` | `list(map(string))` | The targets that can be used to collect `postgres` metrics.

For example, `targets` can either be passed to a `prometheus.relabel`
component to rewrite the metrics' label set, or to a `prometheus.scrape`
component that collects the exposed metrics.

The exported targets will use the configured [in-memory traffic][] address
specified by the [run command][].

[in-memory traffic]: {{< relref "../../concepts/component_controller.md#in-memory-traffic" >}}
[run command]: {{< relref "../cli/run.md" >}}

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
// Because no autodiscovery is defined, this will only scrape the 'database_name' database, as defined
// in the DSN below.
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

This example uses a `prometheus.exporter.postgres` component to collect custom metrics from a set of
specific databases, replacing default metrics with custom metrics derived from queries in `/etc/agent/custom-postgres-metrics.yaml`:

```river
prometheus.exporter.postgres "example" {
  data_source_names       = ["postgresql://username:password@localhost:5432/database_name?sslmode=disable"]

  // This block configures autodiscovery to check for databases outside of the 'database_name' db
  // specified in the DSN above. The database_allowlist field means that only the 'frontend_app' and 'backend_app'
  // databases will be scraped.
  autodiscovery {
    enabled = true
    database_allowlist = ["frontend_app", "backend_app"]
  }

  disable_default_metrics    = true
  custom_queries_config_path = "/etc/agent/custom-postgres-metrics.yaml"
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

This example uses a `prometheus.exporter.postgres` component to collect custom metrics from all databases except
for the `secrets` database.
```river
prometheus.exporter.postgres "example" {
  data_source_names       = ["postgresql://username:password@localhost:5432/database_name?sslmode=disable"]

  // The database_denylist field will filter out those databases from the list of databases to scrape,
  // meaning that all databases *except* these will be scraped.
  //
  // In this example it will scrape all databases except for the one named 'secrets'.
  autodiscovery {
    enabled = true
    database_denylist = ["secrets"]
  }
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
