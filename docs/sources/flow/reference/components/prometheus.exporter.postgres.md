---
# NOTE(rfratto): the title below has zero-width spaces injected into it to
# prevent it from overflowing the sidebar on the rendered site. Be careful when
# modifying this section to retain the spaces.
#
# Ideally, in the future, we can fix the overflow issue with css rather than
# injecting special characters.

title: prometheus.exporter.postgres
---

# prometheus.exporter.postgres
The `prometheus.exporter.postgres` component embeds
[postgres_exporter](https://github.com/prometheus-community/postgres_exporter) for collecting metrics from a postgres database.

## Usage

```river
prometheus.exporter.postgres "LABEL" {
    data_source_names = ["DATA_SOURCE_NAMES"...]
}
```

## Arguments
The following arguments can be used to configure the exporter's behavior.
Omitted fields take their default values.

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`redis_addr`                  | `string`       | Address (host and port) of the Redis instance to connect to.  | | yes



## Exported fields
The following fields are exported and can be referenced by other components.

Name      | Type                | Description
--------- | ------------------- | -----------
`targets` | `list(map(string))` | The targets that can be used to collect `postgres` metrics.

For example, `targets` can either be passed to a `prometheus.relabel`
component to rewrite the metrics' label set, or to a `prometheus.scrape`
component that collects the exposed metrics.

## Component health

`prometheus.exporter.postgres` is only reported as unhealthy if given
an invalid configuration. In those cases, exported fields retain their last
healthy values.

## Debug information

`prometheus.exporter.postgres` does not expose any component-specific
debug information.

## Debug metrics

`prometheus.exporter.postgres` does not expose any component-specific
debug metrics.

## Example

This example uses a [`prometheus.scrape` component][scrape] to collect metrics
from `prometheus.exporter.postgres`:

```river
prometheus.exporter.postgres "example" {
  data_source_names = ["postgresql://username:password@localhost:5432/database_name?sslmode=disable"]
}

// Configure a prometheus.scrape component to collect Redis metrics.
prometheus.scrape "demo" {
  targets    = prometheus.exporter.redis.example.targets
  forward_to = [ /* ... */ ]
}
```

[scrape]: {{< relref "./prometheus.scrape.md" >}}
