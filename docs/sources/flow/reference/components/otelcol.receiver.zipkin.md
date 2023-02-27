---
title: otelcol.receiver.zipkin
---

# otelcol.receiver.zipkin

`otelcol.receiver.zipkin` accepts Zipkin-formatted traces over the network and
forwards it to other `otelcol.*` components.

> **NOTE**: `otelcol.receiver.zipkin` is a wrapper over the upstream
> OpenTelemetry Collector `zipkin` receiver. Bug reports or feature requests
> will be redirected to the upstream repository, if necessary.

Multiple `otelcol.receiver.zipkin` components can be specified by giving them
different labels.

## Usage

```river
otelcol.receiver.zipkin "LABEL" {
  output {
    traces = [...]
  }
}
```

## Arguments

`otelcol.receiver.zipkin` supports the following arguments:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`parse_string_tags` | `bool` | Parse string tags and binary annotations into non-string types. | `false` | no

If `parse_string_tags` is `true`, string tags and binary annotations are
converted to `int`, `bool`, and `float` if possible. String tags and binary
annotations that cannot be converted remain unchanged.

## Blocks

The following blocks are supported inside the definition of
`otelcol.receiver.zipkin`:

Hierarchy | Block | Description | Required
--------- | ----- | ----------- | --------
http | [http][] | Configures the HTTP server to receive telemetry data. | no
http > tls | [tls][] | Configures TLS for the HTTP server. | no
http > cors | [cors][] | Configures CORS for the HTTP server. | no
output | [output][] | Configures where to send received traces. | yes

The `>` symbol indicates deeper levels of nesting. For example, `grpc > tls`
refers to a `tls` block defined inside a `grpc` block.

[http]: #http-block
[tls]: #tls-block
[cors]: #cors-block
[output]: #output-block

### http block

The `http` block configures the HTTP server used by the component.

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`endpoint` | `string` | `host:port` to listen for traffic on. | `"0.0.0.0:9411"` | no
`max_request_body_size` | `string` | Maximum request body size the server will allow. No limit when unset. | | no
`include_metadata` | `boolean` | Propagate incoming connection metadata to downstream consumers. | | no

### tls block

The `tls` block configures TLS settings used for a server. If the `tls` block
isn't provided, TLS won't be used for connections to the server.

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`ca_file` | `string` | Path to the CA file. | | no
`cert_file` | `string` | Path to the TLS certificate. | | no
`key_file` | `string` | Path to the TLS certificate key. | | no
`min_version` | `string` | Minimum acceptable TLS version for connections. | `"TLS 1.2"` | no
`max_version` | `string` | Maximum acceptable TLS version for connections. | `"TLS 1.3"` | no
`reload_interval` | `duration` | Frequency to reload the certificates. | | no
`client_ca_file` | `string` | Path to the CA file used to authenticate client certificates. | | no

### cors block

The `cors` block configures CORS settings for an HTTP server.

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`allowed_origins` | `list(string)` | Allowed values for the `Origin` header. | | no
`allowed_headers` | `list(string)` | Accepted headers from CORS requests. | `["X-Requested-With"]` | no
`max_age` | `number` | Configures the `Access-Control-Max-Age` response header. | | no

The `allowed_headers` argument specifies which headers are acceptable from a
CORS request. The following headers are always implicitly allowed:

* `Accept`
* `Accept-Language`
* `Content-Type`
* `Content-Language`

If `allowed_headers` includes `"*"`, all headers are permitted.

### output block

{{< docs/shared lookup="flow/reference/components/output-block.md" source="agent" >}}

## Exported fields

`otelcol.receiver.zipkin` does not export any fields.

## Component health

`otelcol.receiver.zipkin` is only reported as unhealthy if given an invalid
configuration.

## Debug information

`otelcol.receiver.zipkin` does not expose any component-specific debug
information.

## Example

This example forwards received traces through a batch processor before finally
sending it to an OTLP-capable endpoint:

```river
otelcol.receiver.zipkin "default" {
  output {
    traces = [otelcol.processor.batch.default.input]
  }
}

otelcol.processor.batch "default" {
  output {
    metrics = [otelcol.exporter.otlp.default.input]
    logs    = [otelcol.exporter.otlp.default.input]
    traces  = [otelcol.exporter.otlp.default.input]
  }
}

otelcol.exporter.otlp "default" {
  client {
    endpoint = env("OTLP_ENDPOINT")
  }
}
```
