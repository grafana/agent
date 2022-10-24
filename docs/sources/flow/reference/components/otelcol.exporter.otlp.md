---
aliases:
- /docs/agent/latest/flow/reference/components/otelcol.exporter.otlp
title: otelcol.exporter.otlp
---

# otelcol.exporter.otlp

`otelcol.exporter.otlp` accepts telemetry data from other `otelcol` components
and writes them over the network using the OTLP gRPC protocol.

> **NOTE**: `otelcol.exporter.otlp` is a wrapper over the upstream
> OpenTelemetry Collector `otlp` exporter. Bug reports or feature requests will
> be redirected to the upstream repository, if necessary.

Multiple `otelcol.exporter.otlp` components can be specified by giving them
different labels.

## Usage

```river
otelcol.exporter.otlp "LABEL" {
  client {
    endpoint = "HOST:PORT"
  }
}
```

## Arguments

`otelcol.exporter.otlp` supports the following arguments:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`timeout` | `duration` | Time to wait before marking a request as failed. | `"5s"` | no

## Blocks

The following blocks are supported inside the definition of
`otelcol.exporter.otlp`:

Hierarchy | Block | Description | Required
--------- | ----- | ----------- | --------
client | [client][] | Configures the gRPC server to send telemetry data to. | yes
client > tls | [tls][] | Configures TLS for the gRPC client. | no
client > keepalive | [keepalive][] | Configures keepalive settings for the gRPC client. | no
queue | [queue][] | Configures batching of data before sending. | no
retry | [retry][] | Configures retry mechanism for failed requests. | no

The `>` symbol indicates deeper levels of nesting. For example, `client > tls`
refers to a `tls` block defined inside a `client` block.

[client]: #client-block
[tls]: #tls-block
[keepalive]: #keepalive-block
[queue]: #queue-block
[retry]: #retry-block

### client block

The `client` block configures the gRPC client used by the component.

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`endpoint` | `string` | `host:port` to send telemetry data to. | | yes
`compression` | `string` | Compression mechanism to use for requests. | `"gzip"` | no
`read_buffer_size` | `string` | Size of the read buffer the gRPC client to use for reading server responses. | | no
`write_buffer_size` | `string` | Size of the write buffer the gRPC client to use for writing requests. | `"512KiB"` | no
`wait_for_ready` | `boolean` | Waits for gRPC connection to be in the `READY` state before sending data. | `false` | no
`headers` | `map(string)` | Additional headers to send with the request. | `{}` | no
`balancer_name` | `string` | Which gRPC client-side load balancer to use for requests. | | no
`auth` | `capsule(otelcol.Handler)` | Handler from an `otelcol.auth` component to use for authenticating requests. | | no

By default, requests are compressed with gzip. The `compression` argument
controls which compression mechanism to use. Supported strings are:

* `"gzip"`
* `"zlib"`
* `"deflate"`
* `"snappy"`
* `"zstd"`

If `compression` is set to `"none"` or an empty string `""`, no compression is
used.

The `balancer_name` argument controls what client-side load balancing mechanism
to use. See the gRPC documentation on [Load balancing][] for more information.
When unspecified, `pick_first` is used.

An HTTP proxy can be configured through the following environment variables:

* `HTTPS_PROXY`
* `NO_PROXY`

Connections to the proxy are established via [the `HTTP CONNECT` method][HTTP
CONNECT].

Because `otelcol.exporter.otlp` uses gRPC, the configured proxy server must be
able to handle and proxy HTTP/2 traffic.

[Load balancing]: https://github.com/grpc/grpc-go/blob/master/examples/features/load_balancing/README.md#pick_first
[HTTP CONNECT]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods/CONNECT

### tls block

The `tls` block configures TLS settings used for the connection to the gRPC
server.

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`ca_file` | `string` | Path to the CA file. | | no
`cert_file` | `string` | Path to the TLS certificate. | | no
`key_file` | `string` | Path to the TLS certificate key. | | no
`min_version` | `string` | Minimum acceptable TLS version for connections. | `"TLS 1.2"` | no
`max_version` | `string` | Maximum acceptable TLS version for connections. | `"TLS 1.3"` | no
`insecure` | `boolean` | Disables TLS when connecting to the gRPC server. | | no
`insecure_skip_verify` | `boolean` | Ignores insecure server TLS sertificates. | | no
`server_name` | `string` | Verifies the hostname of server certificates when set. | | no

The `tls` block should always be provided, even if the server doesn't support
TLS. To disable `tls` for connections to the gRPC server, set the `insecure`
argument to `true`.

### keepalive block

The `keepalive` block configures keepalive settings for gRPC client
connections.

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`ping_wait` | `duration` | How often to ping the server after no activity. | | no
`ping_response_timeout` | `duration` | Time to wait before closing inactive connections if the server does not respond to a ping. | | no
`ping_without_stream` | `boolean` | Send pings even if there is no active stream request. | | no

### queue block

The `queue` block configures an in-memory buffer of batches before data is sent
to the gRPC server.

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`enabled` | `boolean` | Enables an in-memory buffer before sending data to the client. | `true` | no
`num_consumers` | `number` | Number of readers to send batches written to the queue in parallel. | `10` | no
`queue_size` | `number` | Maximum number of unwritten batches allowed in the queue at once. | `5000` | no

When `enabled` is `true`, data is first written to an in-memory buffer before
sending it to the configured gRPC server. Batches sent to the component's
`input` exported field are added to the buffer as long as the number of unsent
batches does not exceed the configured `queue_size`.

`queue_size` is used to determine how long an endpoint outage is tolerated for.
Assuming 100 requests/second, the default queue size `5000` provides about 50
seconds of outage tolerance. To calculate the correct value for `queue_size`,
multiply the average number of outgoing requests per second by the amount of
time in seconds outages should be tolerated for.

The `num_consumers` argument controls how many readers read from the buffer and
send data in parallel. Larger values of `num_consumers` allow data to be sent
more quickly at the expense of increased network traffic.

### retry block

The `retry` block configures how failed requests to the gRPC server are
retried.

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`enabled` | `boolean` | Enables retrying failed requests. | `true` | no
`initial_interval` | `duration` | Initial time to wait before retrying a failed request. | `"5s"` | no
`max_interval` | `duration` | Maximum time to wait between retries. | `"30s"` | no
`max_elapsed_time` | `duration` | Maximum amount of time to wait before discarding a failed batch. | `"5m"` | no

When `enabled` is `true`, failed batches are retried after a given interval.
The `initial_interval` argument specifies how long to wait before the first
retry attempt. If requests continue to fail, the time to wait before retrying
increases exponentially. The `max_interval` argument specifies the upper bound
of how long to wait between retries.

If a batch has not sent successfully, it is discarded after the time specified
by `max_elapsed_time` elapses. If `max_elapsed_time` is set to `"0s"`, failed
requests are retried forever until they succeed.

## Exported fields

The following fields are exported and can be referenced by other components:

Name | Type | Description
---- | ---- | -----------
`input` | `otelcol.Consumer` | A value that other components can use to send telemetry data to.

`input` accepts `otelcol.Consumer` data for any telemetry signal (metrics,
logs, or traces).

## Component health

`otelcol.exporter.otlp` is only reported as unhealthy if given an invalid
configuration.

## Debug information

`otelcol.exporter.otlp` does not expose any component-specific debug
information.

## Example

This example creates an exporter to send data to a locally running Grafana
Tempo without TLS:

```river
otelcol.exporter.otlp "tempo" {
    client {
        endpoint = "tempo:4317"
        tls {
            insecure             = true
            insecure_skip_verify = true
        }
    }
}
```
