---
title: otelcol.exporter.jaeger
---

# otelcol.exporter.jaeger

`otelcol.exporter.jaeger` accepts telemetry data from other `otelcol` components
and writes them over the network using the Jaeger protocol.

> **NOTE**: `otelcol.exporter.jaeger` is a wrapper over the 
> [upstream](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter/jaegerexporter) 
> OpenTelemetry Collector `jaeger` exporter. The upstream
> exporter has been deprecated and will be removed from future versions of 
> both OpenTelemetry Collector and Grafana Agent because Jaeger supports OTLP directly.

Multiple `otelcol.exporter.jaeger` components can be specified by giving them
different labels.

## Usage

```river
otelcol.exporter.jaeger "LABEL" {
  client {
    endpoint = "HOST:PORT"
  }
}
```

## Arguments

`otelcol.exporter.jaeger` supports the following arguments:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`timeout` | `duration` | Time to wait before marking a request as failed. | `"5s"` | no

## Blocks

The following blocks are supported inside the definition of
`otelcol.exporter.jaeger`:

Hierarchy | Block | Description | Required
--------- | ----- | ----------- | --------
client | [client][] | Configures the gRPC server to send telemetry data to. | yes
client > tls | [tls][] | Configures TLS for the gRPC client. | no
client > keepalive | [keepalive][] | Configures keepalive settings for the gRPC client. | no
sending_queue | [sending_queue][] | Configures batching of data before sending. | no
retry_on_failure | [retry_on_failure][] | Configures retry mechanism for failed requests. | no

The `>` symbol indicates deeper levels of nesting. For example, `client > tls`
refers to a `tls` block defined inside a `client` block.

[client]: #client-block
[tls]: #tls-block
[keepalive]: #keepalive-block
[sending_queue]: #sending_queue-block
[retry_on_failure]: #retry_on_failure-block

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

{{< docs/shared lookup="flow/reference/components/otelcol-compression-field.md" source="agent" >}}

The `balancer_name` argument controls what client-side load balancing mechanism
to use. See the gRPC documentation on [Load balancing][] for more information.
When unspecified, `pick_first` is used.

An HTTP proxy can be configured through the following environment variables:

* `HTTPS_PROXY`
* `NO_PROXY`

The `HTTPS_PROXY` environment variable specifies a URL to use for proxying
requests. Connections to the proxy are established via [the `HTTP CONNECT`
method][HTTP CONNECT].

The `NO_PROXY` environment variable is an optional list of comma-separated
hostnames for which the HTTPS proxy should _not_ be used. Each hostname can be
provided as an IP address (`1.2.3.4`), an IP address in CIDR notation
(`1.2.3.4/8`), a domain name (`example.com`), or `*`. A domain name matches
that domain and all subdomains. A domain name with a leading "."
(`.example.com`) matches subdomains only. `NO_PROXY` is only read when
`HTTPS_PROXY` is set.

Because `otelcol.exporter.jaeger` uses gRPC, the configured proxy server must be
able to handle and proxy HTTP/2 traffic.

[Load balancing]: https://github.com/grpc/grpc-go/blob/master/examples/features/load_balancing/README.md#pick_first
[HTTP CONNECT]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods/CONNECT

### tls block

The `tls` block configures TLS settings used for the connection to the gRPC
server.

{{< docs/shared lookup="flow/reference/components/otelcol-tls-config-block.md" source="agent" >}}

### keepalive block

The `keepalive` block configures keepalive settings for gRPC client
connections.

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`ping_wait` | `duration` | How often to ping the server after no activity. | | no
`ping_response_timeout` | `duration` | Time to wait before closing inactive connections if the server does not respond to a ping. | | no
`ping_without_stream` | `boolean` | Send pings even if there is no active stream request. | | no

### sending_queue block

The `sending_queue` block configures an in-memory buffer of batches before data is sent
to the gRPC server.

{{< docs/shared lookup="flow/reference/components/otelcol-queue-block.md" source="agent" >}}

### retry_on_failure block

The `retry_on_failure` block configures how failed requests to the gRPC server are
retried.

{{< docs/shared lookup="flow/reference/components/otelcol-retry-block.md" source="agent" >}}

## Exported fields

The following fields are exported and can be referenced by other components:

Name | Type | Description
---- | ---- | -----------
`input` | `otelcol.Consumer` | A value that other components can use to send telemetry data to.

`input` accepts `otelcol.Consumer` data for any telemetry signal (metrics,
logs, or traces).

## Component health

`otelcol.exporter.jaeger` is only reported as unhealthy if given an invalid
configuration.

## Debug information

`otelcol.exporter.jaeger` does not expose any component-specific debug
information.

## Example

This example accepts OTLP traces over gRPC, sends them to a batch processor and forwards to Jaeger without TLS:

```river
otelcol.receiver.otlp "default" {
    grpc {}
    output {
        traces  = [otelcol.processor.batch.default.input]
    }
}

otelcol.processor.batch "default" {
    output {
        traces  = [otelcol.exporter.jaeger.default.input]
    }
}

otelcol.exporter.jaeger "default" {
    client {
        endpoint = "jaeger:14250"
        tls {
            insecure             = true
            insecure_skip_verify = true
        }
    }
}
```
