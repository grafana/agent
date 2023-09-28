---
aliases:
- /docs/grafana-cloud/agent/flow/reference/components/pyroscope.write/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/reference/components/pyroscope.write/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/reference/components/pyroscope.write/
canonical: https://grafana.com/docs/agent/latest/flow/reference/components/pyroscope.write/
labels:
  stage: beta
title: pyroscope.write
description: Learn about pyroscope.write
---

# pyroscope.write

{{< docs/shared lookup="flow/stability/beta.md" source="agent" version="<AGENT VERSION>" >}}

`pyroscope.write` receives performance profiles from other components and forwards them
to a series of user-supplied endpoints using [Pyroscope' Push API](/oss/pyroscope/).

Multiple `pyroscope.write` components can be specified by giving them
different labels.

## Usage for Grafana Agent flow mode

```river
pyroscope.write "LABEL" {
  endpoint {
    url = PYROSCOPE_URL

    ...
  }

  ...
}
```

## Arguments

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`external_labels` | `map(string)` | Labels to add to profiles sent over the network. | | no

## Blocks

The following blocks are supported inside the definition of
`pyroscope.write`:

Hierarchy | Block | Description | Required
--------- | ----- | ----------- | --------
endpoint | [endpoint][] | Location to send profiles to. | no
endpoint > basic_auth | [basic_auth][] | Configure basic_auth for authenticating to the endpoint. | no
endpoint > authorization | [authorization][] | Configure generic authorization to the endpoint. | no
endpoint > oauth2 | [oauth2][] | Configure OAuth2 for authenticating to the endpoint. | no
endpoint > oauth2 > tls_config | [tls_config][] | Configure TLS settings for connecting to the endpoint. | no
endpoint > tls_config | [tls_config][] | Configure TLS settings for connecting to the endpoint. | no

The `>` symbol indicates deeper levels of nesting. For example, `endpoint >
basic_auth` refers to a `basic_auth` block defined inside an
`endpoint` block.

[endpoint]: #endpoint-block
[basic_auth]: #basic_auth-block
[authorization]: #authorization-block
[oauth2]: #oauth2-block
[tls_config]: #tls_config-block

### endpoint block

The `endpoint` block describes a single location to send profiles to. Multiple
`endpoint` blocks can be provided to send profiles to multiple locations.

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`url` | `string` | Full URL to send metrics to. | | yes
`name` | `string` | Optional name to identify the endpoint in metrics. | | no
`remote_timeout` | `duration` | Timeout for requests made to the URL. | `"10s"` | no
`headers` | `map(string)` | Extra headers to deliver with the request. | | no
`min_backoff_period`  | `duration` | Initial backoff time between retries. | `"500ms"`      | no
`max_backoff_period`  | `duration` | Maximum backoff time between retries. | `"5m"`         | no
`max_backoff_retries` | `int`      | Maximum number of retries. 0 to retry infinitely.      | 10             | no
`bearer_token` | `secret` | Bearer token to authenticate with. | | no
`bearer_token_file` | `string` | File containing a bearer token to authenticate with. | | no
`proxy_url` | `string` | HTTP proxy to proxy requests through. | | no
`follow_redirects` | `bool` | Whether redirects returned by the server should be followed. | `true` | no
`enable_http2` | `bool` | Whether HTTP2 is supported for requests. | `true`  | no

 At most one of the following can be provided:

- [`bearer_token` argument][endpoint].
- [`bearer_token_file` argument][endpoint].
- [`basic_auth` block][basic_auth].
- [`authorization` block][authorization].
- [`oauth2` block][oauth2].

When multiple `endpoint` blocks are provided, profiles are concurrently forwarded to all
configured locations.

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
`receiver` | `receiver` | A value that other components can use to send profiles to.

## Component health

`pyroscope.write` is only reported as unhealthy if given an invalid
configuration. In those cases, exported fields are kept at their last healthy
values.

## Debug information

`pyroscope.write` does not expose any component-specific debug
information.

## Example

```river
pyroscope.write "staging" {
  // Send metrics to a locally running Pyroscope instance.
  endpoint {
    url = "http://pyroscope:4100"
    headers = {
      "X-Scope-OrgID" = "squad-1",
    }
  }
  external_labels = {
    "env" = "staging",
  }
}


pyroscope.scrape "default" {
  targets = [
    {"__address__" = "pyroscope:4100", "service_name"="pyroscope"},
    {"__address__" = "agent:12345", "service_name"="agent"},
  ]
  forward_to = [pyroscope.write.staging.receiver]
}
```
