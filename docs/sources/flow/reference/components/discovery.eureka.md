---
aliases:
- /docs/grafana-cloud/agent/flow/reference/components/discovery.eureka/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/reference/components/discovery.eureka/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/reference/components/discovery.eureka/
canonical: https://grafana.com/docs/agent/latest/flow/reference/components/discovery.eureka/
title: discovery.eureka
description: Learn about discovery.eureka
---

# discovery.eureka

`discovery.eureka` discovers instances in a [Eureka][] Registry and exposes them as targets.

[Eureka]: https://github.com/Netflix/eureka/

## Usage

```river
discovery.eureka "LABEL" {
    server = SERVER
}
```

## Arguments

The following arguments are supported:

Name                | Type       | Description                                                            | Default              | Required
------------------- | ---------- | ---------------------------------------------------------------------- | -------------------- | --------
`server`            | `string`   | Eureka server URL.                                                     |                      | yes
`refresh_interval`  | `duration` | Interval at which to refresh the list of targets.                      | `30s`                | no
`enable_http2`      | `bool`     | Whether HTTP2 is supported for requests.                               | `true`               | no
`follow_redirects`  | `bool`     | Whether redirects returned by the server should be followed.           | `true`               | no

## Blocks
The following blocks are supported inside the definition of
`discovery.eureka`:

Hierarchy | Block | Description | Required
--------- | ----- | ----------- | --------
basic_auth | [basic_auth][] | Configure basic_auth for authenticating to the endpoint. | no
authorization | [authorization][] | Configure generic authorization to the endpoint. | no
oauth2 | [oauth2][] | Configure OAuth2 for authenticating to the endpoint. | no
oauth2 > tls_config | [tls_config][] | Configure TLS settings for connecting to the endpoint. | no

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

Name      | Type                | Description
--------- | ------------------- | -----------
`targets` | `list(map(string))` | The set of targets discovered from the Eureka API.

Each target includes the following labels:

* `__meta_eureka_app_name`
* `__meta_eureka_app_instance_hostname`
* `__meta_eureka_app_instance_homepage_url`
* `__meta_eureka_app_instance_statuspage_url`
* `__meta_eureka_app_instance_healthcheck_url`
* `__meta_eureka_app_instance_ip_addr`
* `__meta_eureka_app_instance_vip_address`
* `__meta_eureka_app_instance_secure_vip_address`
* `__meta_eureka_app_instance_status`
* `__meta_eureka_app_instance_port`
* `__meta_eureka_app_instance_port_enabled`
* `__meta_eureka_app_instance_secure_port`
* `__meta_eureka_app_instance_secure_port_enabled`
* `__meta_eureka_app_instance_datacenterinfo_name`
* `__meta_eureka_app_instance_datacenterinfo_metadata_`
* `__meta_eureka_app_instance_country_id`
* `__meta_eureka_app_instance_id`
* `__meta_eureka_app_instance_metadata_`

## Component health

`discovery.eureka` is only reported as unhealthy when given an invalid
configuration. In those cases, exported fields retain their last healthy
values.

## Debug information

`discovery.eureka` does not expose any component-specific debug information.

## Debug metrics

`discovery.eureka` does not expose any component-specific debug metrics.

## Example

```river
discovery.eureka "example" {
    server = "https://eureka.example.com/eureka/v1"
}

prometheus.scrape "demo" {
  targets    = discovery.eureka.example.targets
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
