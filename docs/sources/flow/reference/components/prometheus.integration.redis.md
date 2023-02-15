---
# NOTE(rfratto): the title below has zero-width spaces injected into it to
# prevent it from overflowing the sidebar on the rendered site. Be careful when
# modifying this section to retain the spaces.
#
# Ideally, in the future, we can fix the overflow issue with css rather than
# injecting special characters.

title: prometheus.​integration.redis
---

# prometheus.integration.redis
The `prometheus.integration.redis` component embeds
[redis_exporter](https://github.com/oliver006/redis_exporter) for collecting metrics from a redis database.

## Usage

```river
prometheus.integration.redis "LABEL"{
}
```

## Arguments
The following arguments can be used to configure the exporter's behavior.
Omitted fields take their default values.

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`redis_addr`                  | `string`   | Address (host and port) of the Redis instance to connect to. This can be a local Redis instance (e.g., `localhost:6379`) or the address of a remote server. | | yes
`redis_user`                  | `string`   | User name to use for authentication (Redis ACL for Redis 6.0 and newer).  | | no
`redis_password`              | `string`   | Password of the Redis instance. | | no
`redis_password_file`         | `string`   | Path of a file containing a password. If defined, it takes precedence over redis_password. | | no
`redis_password_map_file`     | `string`   | Path of a file containing a JSON object which maps Redis URIs to passwords (e.g., `{"redis://localhost:6379": "sample_password"}`). | | no
`namespace`                   | `string`   | Namespace for the metrics.  | `"redis"` | no
`config_command`              | `string`   | What to use for the CONFIG command. | `"CONFIG"` | no
`check_keys`                  | `string`   | Comma-separated list of key-patterns to export value and length/size, searched for with SCAN. | | no
`check_key_groups`            | `string`   | Comma-separated list of Lua regular expressions (regex) for grouping keys. When unset, no key groups are made. | | no
`check_key_groups_batch_size` | `int`      | Check key or key groups batch size hint for the underlying SCAN. The option name reflects key groups for backwards compatibility, but this applies to both key and key groups. | `10000` | no
`max_distinct_key_groups`     | `int`      | The maximum number of distinct key groups with the most memory utilization to present as distinct metrics per database. The leftover key groups are aggregated in the 'overflow' bucket. | `100` | no
`check_single_keys`           | `string`   | Comma separated list of single keys to export value and length/size. | | no
`check_streams`               | `string`   | Comma separated list of stream-patterns to export info about streams, groups and consumers, searched for with SCAN. | | no
`check_single_streams`        | `string`   | Comma separated list of single streams to export info about streams, groups and consumers. | | no
`count_keys`                  | `string`   | Comma separated list of individual keys to export counts for. | | no
`script_path`                 | `string`   | Path to Lua Redis script for collecting extra metrics. | | no
`connection_timeout`          | `duration` | Timeout for connection to Redis instance (in Golang duration format). | `"15s"` | no
`tls_client_key_file`         | `string`   | Name of the client key file (including full path) if the server requires TLS client authentication. | | no
`tls_client_cert_file`        | `string`   | Name of the client certificate file (including full path) if the server requires TLS client authentication. | | no
`tls_ca_cert_file`            | `string`   | Name of the CA certificate file (including full path) if the server requires TLS client authentication. | | no
`set_client_name`             | `bool`     | Whether to set client name to redis_exporter. | `true` | no
`is_tile38`                   | `bool`     | Whether to scrape Tile38 specific metrics. | | no
`export_client_list`          | `bool`     | Whether to scrape Client List specific metrics. | | no
`export_client_port`          | `bool`     | Whether to include the client's port when exporting the client list. Note that including this will increase the cardinality of all redis metrics. | | no
`redis_metrics_only`          | `bool`     | Whether to also export go runtime metrics. | | no
`ping_on_connect`             | `bool`     | Whether to ping the redis instance after connecting. | | no
`incl_system_metrics`         | `bool`     | Whether to include system metrics like e.g. redis_total_system_memory_bytes. | | no
`skip_tls_verification`       | `bool`     | Whether to to skip TLS verification. | | no



## Exported fields
The following fields are exported and can be referenced by other components.

Name      | Type                | Description
--------- | ------------------- | -----------
`targets` | `list(map(string))` | The targets that can be used to collect `redis` metrics.

For example, the `targets` could either be passed to a `prometheus.relabel`
component to rewrite the metrics' label set, or to a `prometheus.scrape`
component that collects the exposed metrics.

## Component health

`prometheus.integration.redis` is only reported as unhealthy if given
an invalid configuration. In those cases, exported fields retain their last
healthy values.

## Debug information

`prometheus.integration.redis` does not expose any component-specific
debug information.

## Debug metrics

`prometheus.integration.redis` does not expose any component-specific
debug metrics.

## Example

This example uses a [`prometheus.scrape` component][scrape] to collect metrics
from `prometheus.integration.redis`:

```river
prometheus.integration.redis "example" {
  redis_addr = "localhost:6379"
}

// Configure a prometheus.scrape component to collect redis metrics.
prometheus.scrape "demo" {
  targets    = prometheus.integration.redis.example.targets
  forward_to = [ /* ... */ ]
}
```

[scrape]: {{< relref "./prometheus.scrape.md" >}}
