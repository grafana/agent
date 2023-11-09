---
aliases:
- /docs/grafana-cloud/agent/flow/reference/components/otelcol.auth.headers/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/reference/components/otelcol.auth.headers/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/reference/components/otelcol.auth.headers/
- /docs/grafana-cloud/send-data/agent/flow/reference/components/otelcol.auth.headers/
canonical: https://grafana.com/docs/agent/latest/flow/reference/components/otelcol.auth.headers/
description: Learn about otelcol.auth.headers
title: otelcol.auth.headers
---

# otelcol.auth.headers

`otelcol.auth.headers` exposes a `handler` that can be used by other `otelcol`
components to authenticate requests using custom headers.

> **NOTE**: `otelcol.auth.headers` is a wrapper over the upstream OpenTelemetry
> Collector `headerssetter` extension. Bug reports or feature requests will be
> redirected to the upstream repository, if necessary.

Multiple `otelcol.auth.headers` components can be specified by giving them
different labels.

## Usage

```river
otelcol.auth.headers "LABEL" {
  header {
    key   = "HEADER_NAME"
    value = "HEADER_VALUE"
  }
}
```

## Arguments

`otelcol.auth.headers` doesn't support any arguments and is configured fully
through inner blocks.

## Blocks

The following blocks are supported inside the definition of
`otelcol.auth.headers`:

Hierarchy | Block | Description | Required
--------- | ----- | ----------- | --------
header | [header][] | Custom header to attach to requests. | no

[header]: #header-block

### header block

The `header` block defines a custom header to attach to requests. It is valid
to provide multiple `header` blocks to set more than one header.

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`key` | `string` | Name of the header to set. | | yes
`value` | `string` or `secret` | Value of the header. | | no
`from_context` | `string` | Metadata name to get header value from. | | no
`action` | `string` | An action to perform on the header | "upsert" | no

The supported values for `action` are:
* `insert`: Inserts the new header if it does not exist.
* `update`: Updates the header value if it exists.
* `upsert`: Inserts a header if it does not exist and updates the header if it exists.
* `delete`: Deletes the header.

Exactly one of `value` or `from_context` must be provided for each `header`
block.

The `value` attribute sets the value of the header directly.

Alternatively, `from_context` can be used to dynamically retrieve the header
value from request metadata.

## Exported fields

The following fields are exported and can be referenced by other components:

Name | Type | Description
---- | ---- | -----------
`handler` | `capsule(otelcol.Handler)` | A value that other components can use to authenticate requests.

## Component health

`otelcol.auth.headers` is only reported as unhealthy if given an invalid
configuration.

## Debug information

`otelcol.auth.headers` does not expose any component-specific debug information.

## Example

This example configures [otelcol.exporter.otlp][] to use custom headers:

```river
otelcol.exporter.otlp "example" {
  client {
    endpoint = "my-otlp-grpc-server:4317"
    auth     = otelcol.auth.headers.creds.handler
  }
}

otelcol.auth.headers "creds" {
  header {
    key          = "X-Scope-OrgID"
    from_context = "tenant_id"
  }

  header {
    key   = "User-ID"
    value = "user_id"
  }
}
```

[otelcol.exporter.otlp]: {{< relref "./otelcol.exporter.otlp.md" >}}
