---
aliases:
- /docs/agent/latest/flow/reference/components/otelcol.exporter.otlphttphttp
title: otelcol.exporter.otlphttphttp
---

# otelcol.exporter.otlphttp

`otelcol.exporter.otlphttp` accepts telemetry data from other `otelcol`
components and writes them over the network using the OTLP HTTP protocol.

> **NOTE**: `otelcol.exporter.otlphttp` is a wrapper over the upstream
> OpenTelemetry Collector `otlphttp` exporter. Bug reports or feature requests
> will be redirected to the upstream repository, if necessary.

Multiple `otelcol.exporter.otlphttp` components can be specified by giving them
different labels.

## Usage

```river
otelcol.exporter.otlphttp "LABEL" {
  client {
    endpoint = "HOST:PORT"
  }
}
```

## Arguments

`otelcol.exporter.otlphttp` supports the following arguments:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`metrics_endpoint` | `string` | The endpoint to send metrics to | `client.endpoint + "/v1/metrics"` | no
`logs_endpoint`    | `string` | The endpoint to send logs to    | `client.endpoint + "/v1/logs"`    | no
`traces_endpoint`  | `string` | The endpoint to send traces to  | `client.endpoint + "/v1/traces"`  | no

The default value depends on the `endpoint` field set in the required `client`
block. If set, these arguments override the `client.endpoint` field for the
corresponding signal.

## Blocks

The following blocks are supported inside the definition of
`otelcol.exporter.otlphttp`:

Hierarchy | Block | Description | Required
--------- | ----- | ----------- | --------
client           | [client][] | Configures the HTTP server to send telemetry data to. | yes
client > tls     | [tls][] | Configures TLS for the HTTP client. | no
queue            | [queue][] | Configures batching of data before sending. | no
retry            | [retry][] | Configures retry mechanism for failed requests. | no

The `>` symbol indicates deeper levels of nesting. For example, `client > tls`
refers to a `tls` block defined inside a `client` block.

[client]: #client-block
[tls]: #tls-block
[queue]: #queue-block
[retry]: #retry-block

### client block

The `client` block configures the HTTP client used by the component.

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`endpoint`           | `string`      | `host:port` to send telemetry data to. | | yes
`read_buffer_size`   | `string`      | Size of the read buffer the HTTP client uses for reading server responses. | `0` | no
`write_buffer_size`  | `string`      | Size of the write buffer the HTTP client uses for writing requests. | `"512KiB"` | no
`timeout`            | `duration`    | Time to wait before marking a request as failed. | `"30s"` | no
`headers`            | `map(string)` | Additional headers to send with the request. | `{}` | no
`compression`        | `string`      | Compression mechanism to use for requests. | `"gzip"` | no
`max_idle_conns`     | `int`         | Limits the number of idle HTTP connections the client can keep open. | `100` | no
`max_idle_conns_per_host` | `int`    | Limits the number of idle HTTP connections the host can keep open. | `0` | no
`max_conns_per_host` | `int`         | Limits the total (dialing,active, and idle) number of connections per host. | `0` | no
`idle_conn_timeout`  | `duration`    | Time to wait before an idle connection will close itself. | `"90s"` | no
`auth`               | `capsule(otelcol.Handler)` | Handler from an `otelcol.auth` component to use for authenticating requests. | | no

By default, requests are compressed with gzip. The `compression` argument
controls which compression mechanism to use. Supported strings are:

* `"gzip"`
* `"zlib"`
* `"deflate"`
* `"snappy"`
* `"zstd"`

If `compression` is set to `"none"` or an empty string `""`, no compression is
used.

### tls block

The `tls` block configures TLS settings used for the connection to the HTTP
server.

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`ca_file` | `string` | Path to the CA file. | | no
`cert_file` | `string` | Path to the TLS certificate. | | no
`key_file` | `string` | Path to the TLS certificate key. | | no
`min_version` | `string` | Minimum acceptable TLS version for connections. | `"TLS 1.2"` | no
`max_version` | `string` | Maximum acceptable TLS version for connections. | `"TLS 1.3"` | no
`insecure` | `boolean` | Disables TLS when connecting to the HTTP server. | | no
`insecure_skip_verify` | `boolean` | Ignores insecure server TLS certificates. | | no
`server_name` | `string` | Verifies the hostname of server certificates when set. | | no

The `tls` block should always be provided, even if the server doesn't support
TLS. To disable `tls` for connections to the HTTP server, set the `insecure`
argument to `true`.

### queue block

The `queue` block configures an in-memory buffer of batches before data is sent
to the HTTP server.

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`enabled` | `boolean` | Enables an in-memory buffer before sending data to the client. | `true` | no
`num_consumers` | `number` | Number of readers to send batches written to the queue in parallel. | `10` | no
`queue_size` | `number` | Maximum number of unwritten batches allowed in the queue at once. | `5000` | no

When `enabled` is `true`, data is first written to an in-memory buffer before
sending it to the configured HTTP server. Batches sent to the component's
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

The `retry` block configures how failed requests to the HTTP server are
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

`otelcol.exporter.otlphttp` is only reported as unhealthy if given an invalid
configuration.

## Debug information

`otelcol.exporter.otlphttp` does not expose any component-specific debug
information.

## Example

This example creates an exporter to send data to a locally running Grafana
Tempo without TLS:

```river
otelcol.exporter.otlphttp "tempo" {
    client {
        endpoint = "tempo:4317"
        tls {
            insecure             = true
            insecure_skip_verify = true
        }
    }
}
```
