---
aliases:
- /docs/grafana-cloud/agent/flow/reference/components/otelcol.receiver.otlp/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/reference/components/otelcol.receiver.otlp/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/reference/components/otelcol.receiver.otlp/
canonical: https://grafana.com/docs/agent/latest/flow/reference/components/otelcol.receiver.otlp/
title: otelcol.receiver.otlp
description: Learn about otelcol.receiver.otlp
---

# otelcol.receiver.otlp

`otelcol.receiver.otlp` accepts OTLP-formatted data over the network and
forwards it to other `otelcol.*` components.

> **NOTE**: `otelcol.receiver.otlp` is a wrapper over the upstream
> OpenTelemetry Collector `otlp` receiver. Bug reports or feature requests will
> be redirected to the upstream repository, if necessary.

Multiple `otelcol.receiver.otlp` components can be specified by giving them
different labels.

## Usage

```river
otelcol.receiver.otlp "LABEL" {
  grpc { ... }
  http { ... }

  output {
    metrics = [...]
    logs    = [...]
    traces  = [...]
  }
}
```

## Arguments

`otelcol.receiver.otlp` doesn't support any arguments and is configured fully
through inner blocks.

## Blocks

The following blocks are supported inside the definition of
`otelcol.receiver.otlp`:

Hierarchy | Block | Description | Required
--------- | ----- | ----------- | --------
grpc | [grpc][] | Configures the gRPC server to receive telemetry data. | no
grpc > tls | [tls][] | Configures TLS for the gRPC server. | no
grpc > keepalive | [keepalive][] | Configures keepalive settings for the configured server. | no
grpc > keepalive > server_parameters | [server_parameters][] | Server parameters used to configure keepalive settings. | no
grpc > keepalive > enforcement_policy | [enforcement_policy][] | Enforcement policy for keepalive settings. | no
http | [http][] | Configures the HTTP server to receive telemetry data. | no
http > tls | [tls][] | Configures TLS for the HTTP server. | no
http > cors | [cors][] | Configures CORS for the HTTP server. | no
debug_metrics | [debug_metrics][] | Configures the metrics that this component generates to monitor its state. | no
output | [output][] | Configures where to send received telemetry data. | yes

The `>` symbol indicates deeper levels of nesting. For example, `grpc > tls`
refers to a `tls` block defined inside a `grpc` block.

[grpc]: #grpc-block
[tls]: #tls-block
[keepalive]: #keepalive-block
[server_parameters]: #server_parameters-block
[enforcement_policy]: #enforcement_policy-block
[http]: #http-block
[cors]: #cors-block
[debug_metrics]: #debug_metrics-block
[output]: #output-block

### grpc block

The `grpc` block configures the gRPC server used by the component. If the
`grpc` block isn't provided, a gRPC server isn't started.

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`endpoint` | `string` | `host:port` to listen for traffic on. | `"0.0.0.0:4317"` | no
`transport` | `string` | Transport to use for the gRPC server. | `"tcp"` | no
`max_recv_msg_size` | `string` | Maximum size of messages the server will accept. 0 disables a limit. | | no
`max_concurrent_streams` | `number` | Limit the number of concurrent streaming RPC calls. | | no
`read_buffer_size` | `string` | Size of the read buffer the gRPC server will use for reading from clients. | `"512KiB"` | no
`write_buffer_size` | `string` | Size of the write buffer the gRPC server will use for writing to clients. | | no
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

### keepalive block

The `keepalive` block configures keepalive settings for connections to a gRPC
server.

`keepalive` doesn't support any arguments and is configured fully through inner
blocks.

### server_parameters block

The `server_parameters` block controls keepalive and maximum age settings for gRPC
servers.

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`max_connection_idle` | `duration` | Maximum age for idle connections. | `"infinity"` | no
`max_connection_age` | `duration` | Maximum age for non-idle connections. | `"infinity"` | no
`max_connection_age_grace` | `duration` | Time to wait before forcibly closing connections. | `"infinity"` | no
`time` | `duration` | How often to ping inactive clients to check for liveness. | `"2h"` | no
`timeout` | `duration` | Time to wait before closing inactive clients that do not respond to liveness checks. | `"20s"` | no

### enforcement_policy block

The `enforcement_policy` block configures the keepalive enforcement policy for
gRPC servers. The server will close connections from clients that violate the
configured policy.

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`min_time` | `duration` | Minimum time clients should wait before sending a keepalive ping. | `"5m"` | no
`permit_without_stream` | `boolean` | Allow clients to send keepalive pings when there are no active streams. | `false` | no

### http block

The `http` block configures the HTTP server used by the component. If the
`http` block isn't specified, an HTTP server isn't started.

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`endpoint` | `string` | `host:port` to listen for traffic on. | `"0.0.0.0:4318"` | no
`max_request_body_size` | `string` | Maximum request body size the server will allow. No limit when unset. | | no
`include_metadata` | `boolean` | Propagate incoming connection metadata to downstream consumers. | | no
`traces_url_path` | `string` | The URL path to receive traces on. | `"/v1/traces"`| no
`metrics_url_path` | `string` | The URL path to receive metrics on. | `"/v1/metrics"` | no
`logs_url_path` | `string` | The URL path to receive logs on. | `"/v1/logs"` | no

To send telemetry signals to `otelcol.receiver.otlp` with HTTP/JSON, POST to:
* `[endpoint][traces_url_path]` for traces.
* `[endpoint][metrics_url_path]` for metrics.
* `[endpoint][logs_url_path]` for logs.

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

### debug_metrics block

{{< docs/shared lookup="flow/reference/components/otelcol-debug-metrics-block.md" source="agent" version="<AGENT VERSION>" >}}

### output block

{{< docs/shared lookup="flow/reference/components/output-block.md" source="agent" version="<AGENT VERSION>" >}}

## Exported fields

`otelcol.receiver.otlp` does not export any fields.

## Component health

`otelcol.receiver.otlp` is only reported as unhealthy if given an invalid
configuration.

## Debug information

`otelcol.receiver.otlp` does not expose any component-specific debug
information.

## Debug metrics

* `receiver_accepted_spans_ratio_total` (counter): Number of spans successfully pushed into the pipeline.
* `receiver_refused_spans_ratio_total` (counter): Number of spans that could not be pushed into the pipeline.
* `rpc_server_duration_milliseconds` (histogram): Duration of RPC requests from a gRPC server.

## Example

This example forwards received telemetry data through a batch processor before
finally sending it to an OTLP-capable endpoint:

```river
otelcol.receiver.otlp "default" {
  http {}
  grpc {}

  output {
    metrics = [otelcol.processor.batch.default.input]
    logs    = [otelcol.processor.batch.default.input]
    traces  = [otelcol.processor.batch.default.input]
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

## Technical details

`otelcol.receiver.otlp` supports [gzip](https://en.wikipedia.org/wiki/Gzip) for compression.
