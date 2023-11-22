---
aliases:
- /docs/grafana-cloud/agent/flow/reference/components/discovery.nomad/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/reference/components/discovery.nomad/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/reference/components/discovery.nomad/
- /docs/grafana-cloud/send-data/agent/flow/reference/components/discovery.nomad/
canonical: https://grafana.com/docs/agent/latest/flow/reference/components/discovery.nomad/
description: Learn about discovery.nomad
title: discovery.nomad
---

# discovery.nomad

`discovery.nomad` allows you to retrieve scrape targets from [Nomad's](https://www.nomadproject.io/) Service API.

## Usage

```river
discovery.nomad "LABEL" {
}
```

## Arguments

The following arguments are supported:

Name                | Type       | Description                                                  | Default                 | Required
--------------------|------------|--------------------------------------------------------------|-------------------------|---------
`allow_stale`       | `bool`     | Allow reading from non-leader nomad instances.               | `true`                  | no
`bearer_token_file` | `string`   | File containing a bearer token to authenticate with.         |                         | no
`bearer_token`      | `secret`   | Bearer token to authenticate with.                           |                         | no
`enable_http2`      | `bool`     | Whether HTTP2 is supported for requests.                     | `true`                  | no
`follow_redirects`  | `bool`     | Whether redirects returned by the server should be followed. | `true`                  | no
`namespace`         | `string`   | Nomad namespace to use.                                      | `default`               | no
`proxy_url`         | `string`   | HTTP proxy to proxy requests through.                        |                         | no
`refresh_interval`  | `duration` | Frequency to refresh list of containers.                     | `"30s"`                 | no
`region`            | `string`   | Nomad region to use.                                         | `global`                | no
`server`            | `string`   | Address of nomad server.                                     | `http://localhost:4646` | no
`tag_separator`     | `string`   | Seperator to join nomad tags into Prometheus labels.         | `,`                     | no

 You can provide one of the following arguments for authentication:
- [`authorization` block][authorization].
- [`basic_auth` block][basic_auth].
- [`bearer_token_file` argument](#arguments).
- [`bearer_token` argument](#arguments).
- [`oauth2` block][oauth2].

[arguments]: #arguments

## Blocks

The following blocks are supported inside the definition of
`discovery.nomad`:

Hierarchy           | Block             | Description                                              | Required
--------------------|-------------------|----------------------------------------------------------|---------
authorization       | [authorization][] | Configure generic authorization to the endpoint.         | no
basic_auth          | [basic_auth][]    | Configure basic_auth for authenticating to the endpoint. | no
oauth2              | [oauth2][]        | Configure OAuth2 for authenticating to the endpoint.     | no
oauth2 > tls_config | [tls_config][]    | Configure TLS settings for connecting to the endpoint.   | no

The `>` symbol indicates deeper levels of nesting.
For example, `oauth2 > tls_config` refers to a `tls_config` block defined inside an `oauth2` block.

[basic_auth]: #basic_auth-block
[authorization]: #authorization-block
[oauth2]: #oauth2-block
[tls_config]: #tls_config-block

### authorization

{{< docs/shared lookup="flow/reference/components/authorization-block.md" source="agent" version="<AGENT_VERSION>" >}}

### basic_auth

{{< docs/shared lookup="flow/reference/components/basic-auth-block.md" source="agent" version="<AGENT_VERSION>" >}}

### oauth2

{{< docs/shared lookup="flow/reference/components/oauth2-block.md" source="agent" version="<AGENT_VERSION>" >}}

### oauth2 > tls_config ock

{{< docs/shared lookup="flow/reference/components/tls-config-block.md" source="agent" version="<AGENT_VERSION>" >}}

## Exported fields

The following fields are exported and can be referenced by other components:

Name      | Type                | Description
----------|---------------------|-----------------------------------------------------
`targets` | `list(map(string))` | The set of targets discovered from the nomad server.

Each target includes the following labels:

* `__meta_nomad_address`: The service address of the target.
* `__meta_nomad_dc`: The datacenter name for the target.
* `__meta_nomad_namespace`: The namespace of the target.
* `__meta_nomad_node_id`: The node name defined for the target.
* `__meta_nomad_service_address`: The service address of the target.
* `__meta_nomad_service_id`: The service ID of the target.
* `__meta_nomad_service_port`: The service port of the target.
* `__meta_nomad_service`: The name of the service the target belongs to.
* `__meta_nomad_tags`: The list of tags of the target joined by the tag separator.

## Component health

`discovery.nomad` is only reported as unhealthy when given an invalid configuration.
In those cases, exported fields retain their last healthy values.

## Debug information

`discovery.nomad` doesn't expose any component-specific debug information.

## Debug metrics

`discovery.nomad` doesn't expose any component-specific debug metrics.

## Example

The following example discovers targets from a Nomad server:

```river
discovery.nomad "example" {
}

prometheus.scrape "demo" {
  targets    = discovery.nomad.example.targets
  forward_to = [prometheus.remote_write.demo.receiver]
}

prometheus.remote_write "demo" {
  endpoint {
    url = <PROMETHEUS_REMOTE_WRITE_URL>

    basic_auth {
      username = <USERNAME>
      password = <PASSWORD>
    }
  }
}
```

Replace the following:
- _`<PROMETHEUS_REMOTE_WRITE_URL>`_: The URL of the Prometheus remote_write-compatible server to send metrics to.
- _`<USERNAME>`_: The username to use for authentication to the remote_write API.
- _`<PASSWORD>`_: The password to use for authentication to the remote_write API.
