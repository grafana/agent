---
aliases:
- /docs/grafana-cloud/agent/flow/reference/components/discovery.marathon/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/reference/components/discovery.marathon/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/reference/components/discovery.marathon/
canonical: https://grafana.com/docs/agent/latest/flow/reference/components/discovery.marathon/
title: discovery.marathon
description: Learn about discovery.marathon
---

# discovery.marathon

`discovery.marathon` allows you to retrieve scrape targets from [Marathon's](https://mesosphere.github.io/marathon/) Service API.

## Usage

```river
discovery.marathon "LABEL" {
  servers = [MARATHON_SERVER1, MARATHON_SERVER2...]
}
```

## Arguments

The following arguments are supported:

| Name               | Type           | Description                                                  | Default | Required |
| ------------------ | -------------- | ------------------------------------------------------------ | ------- | -------- |
| `servers`          | `list(string)` | List of Marathon servers.                                    |         | yes      |
| `refresh_interval` | `duration`     | Interval at which to refresh the list of targets.            | `"30s"` | no       |
| `auth_token`       | `secret`       | Auth token to authenticate with.                             |         | no       |
| `auth_token_file`  | `string`       | File containing an auth token to authenticate with.          |         | no       |
| `proxy_url`        | `string`       | HTTP proxy to proxy requests through.                        |         | no       |
| `follow_redirects` | `bool`         | Whether redirects returned by the server should be followed. | `true`  | no       |
| `enable_http2`     | `bool`         | Whether HTTP2 is supported for requests.                     | `true`  | no       |

You can provide one of the following arguments for authentication:

- [`auth_token` argument](#arguments).
- [`auth_token_file` argument](#arguments).
- [`basic_auth` block][basic_auth].
- [`authorization` block][authorization].
- [`oauth2` block][oauth2].

[arguments]: #arguments

## Blocks

The following blocks are supported inside the definition of
`discovery.marathon`:

| Hierarchy           | Block             | Description                                              | Required |
| ------------------- | ----------------- | -------------------------------------------------------- | -------- |
| basic_auth          | [basic_auth][]    | Configure basic_auth for authenticating to the endpoint. | no       |
| authorization       | [authorization][] | Configure generic authorization to the endpoint.         | no       |
| oauth2              | [oauth2][]        | Configure OAuth2 for authenticating to the endpoint.     | no       |
| oauth2 > tls_config | [tls_config][]    | Configure TLS settings for connecting to the endpoint.   | no       |

The `>` symbol indicates deeper levels of nesting. For example,
`oauth2 > tls_config` refers to a `tls_config` block defined inside
an `oauth2` block.

[basic_auth]: #basic_auth-block
[authorization]: #authorization-block
[oauth2]: #oauth2-block
[tls_config]: #tls_config-block

### basic_auth block

{{< docs/shared lookup="flow/reference/components/basic-auth-block.md" source="agent" version="<AGENT VERSION>" >}}

### authorization block

{{< docs/shared lookup="flow/reference/components/authorization-block.md" source="agent" version="<AGENT VERSION>" >}}

### oauth2 block

{{< docs/shared lookup="flow/reference/components/oauth2-block.md" source="agent" version="<AGENT VERSION>" >}}

### tls_config block

{{< docs/shared lookup="flow/reference/components/tls-config-block.md" source="agent" version="<AGENT VERSION>" >}}

## Exported fields

The following fields are exported and can be referenced by other components:

| Name      | Type                | Description                                              |
| --------- | ------------------- | -------------------------------------------------------- |
| `targets` | `list(map(string))` | The set of targets discovered from the Marathon servers. |

Each target includes the following labels:

- `__meta_marathon_app`: the name of the app (with slashes replaced by dashes).
- `__meta_marathon_image`: the name of the Docker image used (if available).
- `__meta_marathon_task`: the ID of the Mesos task.
- `__meta_marathon_app_label_<labelname>`: any Marathon labels attached to the app.
- `__meta_marathon_port_definition_label_<labelname>`: the port definition labels.
- `__meta_marathon_port_mapping_label_<labelname>`: the port mapping labels.
- `__meta_marathon_port_index`: the port index number (e.g. 1 for PORT1).

## Component health

`discovery.marathon` is only reported as unhealthy when given an invalid
configuration. In those cases, exported fields retain their last healthy
values.

## Debug information

`discovery.marathon` does not expose any component-specific debug information.

## Debug metrics

`discovery.marathon` does not expose any component-specific debug metrics.

## Example

This example discovers targets from a Marathon server:

```river
discovery.marathon "example" {
  servers = ["localhost:8500"]
}

prometheus.scrape "demo" {
  targets    = discovery.marathon.example.targets
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
