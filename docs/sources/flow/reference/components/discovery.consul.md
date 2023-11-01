---
aliases:
- /docs/grafana-cloud/agent/flow/reference/components/discovery.consul/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/reference/components/discovery.consul/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/reference/components/discovery.consul/
canonical: https://grafana.com/docs/agent/latest/flow/reference/components/discovery.consul/
title: discovery.consul
description: Learn about discovery.consul
---

# discovery.consul

`discovery.consul` allows retrieving scrape targets from [Consul's Catalog API][].

[Consul's Catalog API]: https://www.consul.io/use-cases/discover-services

## Usage

```river
discovery.consul "LABEL" {
  server = CONSUL_SERVER
}
```

## Arguments

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`server` | `string` | Host and port of the Consul API. | `localhost:8500` | no
`token` | `secret` | Secret token used to access the Consul API. | | no
`datacenter` | `string` | Datacenter to query. If not provided, the default is used. | | no
`namespace` | `string` | Namespace to use (only supported in Consul Enterprise). | | no
`partition` | `string` | Admin partition to use (only supported in Consul Enterprise). | | no
`tag_separator` | `string` | The string by which Consul tags are joined into the tag label. | `,` | no
`scheme` | `string` | The scheme to use when talking to Consul. | `http` | no
`username` | `string` | The username to use (deprecated in favor of the basic_auth configuration). | | no
`password` | `secret` | The password to use (deprecated in favor of the basic_auth configuration). | | no
`allow_stale` | `bool` | Allow stale Consul results (see [official documentation][consistency documentation]). Will reduce load on Consul. | `true` | no
`services` | `list(string)` | A list of services for which targets are retrieved. If omitted, all services are scraped. | | no
`tags` | `list(string)` | An optional list of tags used to filter nodes for a given service. Services must contain all tags in the list. | | no
`node_meta` | `map(string)` | Node metadata key/value pairs to filter nodes for a given service. | | no
`refresh_interval` | `duration` | Frequency to refresh list of containers. | `"30s"` | no
`bearer_token` | `secret` | Bearer token to authenticate with. | | no
`bearer_token_file` | `string` | File containing a bearer token to authenticate with. | | no
`proxy_url` | `string` | HTTP proxy to proxy requests through. | | no
`follow_redirects` | `bool` | Whether redirects returned by the server should be followed. | `true` | no
`enable_http2` | `bool` | Whether HTTP2 is supported for requests. | `true` | no

 At most one of the following can be provided:
 - [`bearer_token` argument](#arguments).
 - [`bearer_token_file` argument](#arguments). 
 - [`basic_auth` block][basic_auth].
 - [`authorization` block][authorization].
 - [`oauth2` block][oauth2].

[consistency documentation]: https://www.consul.io/api/features/consistency.html
[arguments]: #arguments

## Blocks

The following blocks are supported inside the definition of
`discovery.consul`:

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

Name | Type | Description
---- | ---- | -----------
`targets` | `list(map(string))` | The set of targets discovered from the Consul catalog API.

Each target includes the following labels:

* `__meta_consul_address`: the address of the target.
* `__meta_consul_dc`: the datacenter name for the target.
* `__meta_consul_health`: the health status of the service.
* `__meta_consul_partition`: the admin partition name where the service is registered.
* `__meta_consul_metadata_<key>`: each node metadata key value of the target.
* `__meta_consul_node`: the node name defined for the target.
* `__meta_consul_service_address`: the service address of the target.
* `__meta_consul_service_id`: the service ID of the target.
* `__meta_consul_service_metadata_<key>`: each service metadata key value of the target.
* `__meta_consul_service_port`: the service port of the target.
* `__meta_consul_service`: the name of the service the target belongs to.
* `__meta_consul_tagged_address_<key>`: each node tagged address key value of the target.
* `__meta_consul_tags`: the list of tags of the target joined by the tag separator.

## Component health

`discovery.consul` is only reported as unhealthy when given an invalid
configuration. In those cases, exported fields retain their last healthy
values.

## Debug information

`discovery.consul` does not expose any component-specific debug information.

## Debug metrics

`discovery.consul` does not expose any component-specific debug metrics.

## Example

This example discovers targets from Consul for the specified list of services:

```river
discovery.consul "example" {
  server = "localhost:8500"
  services = [
    "service1",
    "service2",
  ]
}

prometheus.scrape "demo" {
  targets    = discovery.consul.example.targets
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
