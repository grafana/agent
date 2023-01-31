---
title: otelcol.auth.bearer
---

# otelcol.auth.bearer

`otelcol.auth.bearer` exposes a `handler` that can be used by other `otelcol`
components to authenticate requests using bearer token authentication.

> **NOTE**: `otelcol.auth.bearer` is a wrapper over the upstream OpenTelemetry
> Collector `bearertokenauth` extension. Bug reports or feature requests will
> be redirected to the upstream repository, if necessary.

Multiple `otelcol.auth.bearer` components can be specified by giving them
different labels.

## Usage

```river
otelcol.auth.bearer "LABEL" {
  token = "TOKEN"
}
```

## Arguments

`otelcol.auth.bearer` supports the following arguments:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`token` | `secret` | Bearer token to use for authenticating requests. | | yes

## Exported fields

The following fields are exported and can be referenced by other components:

Name | Type | Description
---- | ---- | -----------
`handler` | `capsule(otelcol.Handler)` | A value that other components can use to authenticate requests.

## Component health

`otelcol.auth.bearer` is only reported as unhealthy if given an invalid
configuration.

## Debug information

`otelcol.auth.bearer` does not expose any component-specific debug information.

## Example

This example configures [otelcol.exporter.otlp][] to use bearer token
authentication:

```river
otelcol.exporter.otlp "example" {
  client {
    endpoint = "my-otlp-grpc-server:4317"
    auth     = otelcol.auth.bearer.creds.handler
  }
}

otelcol.auth.bearer "creds" {
  token = env("API_KEY")
}
```

[otelcol.exporter.otlp]: {{< relref "./otelcol.exporter.otlp.md" >}}
