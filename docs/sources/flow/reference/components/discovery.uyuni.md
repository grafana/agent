---
aliases:
- /docs/grafana-cloud/agent/flow/reference/components/discovery.uyuni/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/reference/components/discovery.uyuni/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/reference/components/discovery.uyuni/
canonical: https://grafana.com/docs/agent/latest/flow/reference/components/discovery.uyuni/
title: discovery.uyuni
description: Learn about discovery.uyuni
---

# discovery.uyuni

`discovery.uyuni` discovers [Uyuni][] Monitoring Endpoints and exposes them as targets.

[Uyuni]: https://www.uyuni-project.org/

## Usage

```river
discovery.uyuni "LABEL" {
    server   = SERVER
    username = USERNAME
    password = PASSWORD
}
```

## Arguments

The following arguments are supported:

Name                  | Type       | Description                                                            | Default                  | Required
--------------------- | ---------- | ---------------------------------------------------------------------- | ------------------------ | --------
`server`              | `string`   | The primary Uyuni Server.                                              |                          | yes
`username`            | `string`   | The username to use for authentication to the Uyuni API.               |                          | yes
`password`            | `Secret`   | The password to use for authentication to the Uyuni API.               |                          | yes
`entitlement`         | `string`   | The entitlement to filter on when listing targets.                     | `"monitoring_entitled"`  | no
`separator`           | `string`   | The separator to use when building the `__meta_uyuni_groups` label.    | `","`                    | no
`refresh_interval`    | `duration` | Interval at which to refresh the list of targets.                      | `1m`                     | no
`proxy_url`           | `string`   | HTTP proxy to proxy requests through.                                  |                          | no
`follow_redirects`    | `bool`     | Whether redirects returned by the server should be followed.           | `true`                   | no
`enable_http2`        | `bool`     | Whether HTTP2 is supported for requests.                               | `true`                   | no


## Blocks
The following blocks are supported inside the definition of
`discovery.uyuni`:

Hierarchy | Block | Description | Required
--------- | ----- | ----------- | --------
tls_config | [tls_config][] | TLS configuration for requests to the Uyuni API. | no

[tls_config]: #tls_config-block

### tls_config block

{{< docs/shared lookup="flow/reference/components/tls-config-block.md" source="agent" version="<AGENT VERSION>" >}}

## Exported fields

The following fields are exported and can be referenced by other components:

Name      | Type                | Description
--------- | ------------------- | -----------
`targets` | `list(map(string))` | The set of targets discovered from the Uyuni API.

Each target includes the following labels:

* `__meta_uyuni_minion_hostname`: The hostname of the Uyuni Minion.
* `__meta_uyuni_primary_fqdn`: The FQDN of the Uyuni primary.
* `__meta_uyuni_system_id`: The system ID of the Uyuni Minion.
* `__meta_uyuni_groups`: The groups the Uyuni Minion belongs to.
* `__meta_uyuni_endpoint_name`: The name of the endpoint.
* `__meta_uyuni_exporter`: The name of the exporter.
* `__meta_uyuni_proxy_module`: The name of the Uyuni module.
* `__meta_uyuni_metrics_path`: The path to the metrics endpoint.
* `__meta_uyuni_scheme`: `https` if TLS is enabled on the endpoint, `http` otherwise.

These labels are largely derived from a [listEndpoints](https://www.uyuni-project.org/uyuni-docs-api/uyuni/api/system.monitoring.html)
API call to the Uyuni Server.

## Component health

`discovery.uyuni` is only reported as unhealthy when given an invalid
configuration. In those cases, exported fields retain their last healthy
values.

## Debug information

`discovery.uyuni` does not expose any component-specific debug information.

## Debug metrics

`discovery.uyuni` does not expose any component-specific debug metrics.

## Example

```river
discovery.uyuni "example" {
  server    = "https://127.0.0.1/rpc/api"
  username  = UYUNI_USERNAME
  password  = UYUNI_PASSWORD
}

prometheus.scrape "demo" {
  targets    = discovery.uyuni.example.targets
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
  - `UYUNI_USERNAME`: The username to use for authentication to the Uyuni server.
  - `UYUNI_PASSWORD`: The password to use for authentication to the Uyuni server.
  - `PROMETHEUS_REMOTE_WRITE_URL`: The URL of the Prometheus remote_write-compatible server to send metrics to.
  - `USERNAME`: The username to use for authentication to the remote_write API.
  - `PASSWORD`: The password to use for authentication to the remote_write API.
