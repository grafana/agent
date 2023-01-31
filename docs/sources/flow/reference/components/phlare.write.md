---
title: phlare.write
---

# phlare.write

`phlare.write` receives performance profiles from other components and forwards them
to a series of user-supplied endpoints using [Phlare' Push API](https://grafana.com/oss/phlare/).

Multiple `phlare.write` components can be specified by giving them
different labels.

## Usage

```river
phlare.write "LABEL" {
  endpoint {
    url = PHLARE_URL

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
`phlare.write`:

Hierarchy | Block | Description | Required
--------- | ----- | ----------- | --------
endpoint | [endpoint][] | Location to send profiles to. | no
endpoint > http_client_config | [http_client_config][] | HTTP client settings when connecting to the endpoint. | no
endpoint > http_client_config > basic_auth | [basic_auth][] | Configure basic_auth for authenticating to the endpoint. | no
endpoint > http_client_config > authorization | [authorization][] | Configure generic authorization to the endpoint. | no
endpoint > http_client_config > oauth2 | [oauth2][] | Configure OAuth2 for authenticating to the endpoint. | no
endpoint > http_client_config > oauth2 > tls_config | [tls_config][] | Configure TLS settings for connecting to the endpoint. | no
endpoint > http_client_config > tls_config | [tls_config][] | Configure TLS settings for connecting to the endpoint. | no

The `>` symbol indicates deeper levels of nesting. For example, `endpoint >
http_client_config` refers to an `http_client_config` block defined inside an
`endpoint` block.

### endpoint block

The `endpoint` block describes a single location to send profiles to. Multiple
`endpoint` blocks can be provided to send profiles to multiple locations.

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`url` | `string` | Full URL to send metrics to. | | yes
`name` | `string` | Optional name to identify the endpoint in metrics. | | no
`remote_timeout` | `duration` | Timeout for requests made to the URL. | `"30s"` | no
`headers` | `map(string)` | Extra headers to deliver with the request. | | no

When multiple `endpoint` blocks are provided, profiles are concurrently forwarded to all
configured locations.

### http_client_config block

The `http_client_config` configures settings used to connect to the
remote_write server.

{{< docs/shared lookup="flow/reference/components/http-client-config-block.md" source="agent" >}}

### basic_auth block

{{< docs/shared lookup="flow/reference/components/basic-auth-block.md" source="agent" >}}

### authorization block

{{< docs/shared lookup="flow/reference/components/authorization-block.md" source="agent" >}}

### oauth2 block

{{< docs/shared lookup="flow/reference/components/oauth2-block.md" source="agent" >}}

### tls_config block

{{< docs/shared lookup="flow/reference/components/tls-config-block.md" source="agent" >}}

## Exported fields

The following fields are exported and can be referenced by other components:

Name | Type | Description
---- | ---- | -----------
`receiver` | `receiver` | A value that other components can use to send profiles to.

## Component health

`phlare.write` is only reported as unhealthy if given an invalid
configuration. In those cases, exported fields are kept at their last healthy
values.

## Debug information

`phlare.write` does not expose any component-specific debug
information.

## Example

```river
phlare.write "staging" {
  // Send metrics to a locally running Phlare instance.
  endpoint {
    url = "http://phlare:4100"
    headers = {
      "X-Scope-Org-ID" = "squad-1",
    }
  }
  external_labels = {
    "env" = "staging",
  }
}


phlare.scrape "default" {
  targets = [
    {"__address__" = "phlare:4100", "app"="phlare"},
    {"__address__" = "agent:12345", "app"="agent"},
  ]
  forward_to = [phlare.write.staging.receiver]
}
```
