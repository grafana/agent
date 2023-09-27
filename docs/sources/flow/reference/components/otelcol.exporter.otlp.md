---
aliases:
- /docs/grafana-cloud/agent/flow/reference/components/otelcol.exporter.otlp/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/reference/components/otelcol.exporter.otlp/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/reference/components/otelcol.exporter.otlp/
canonical: https://grafana.com/docs/agent/latest/flow/reference/components/otelcol.exporter.otlp/
title: otelcol.exporter.otlp
description: Learn about otelcol.exporter.otlp
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
sending_queue | [sending_queue][] | Configures batching of data before sending. | no
retry_on_failure | [retry_on_failure][] | Configures retry mechanism for failed requests. | no
debug_metrics | [debug_metrics][] | Configures the metrics that this component generates to monitor its state. | no

The `>` symbol indicates deeper levels of nesting. For example, `client > tls`
refers to a `tls` block defined inside a `client` block.

[client]: #client-block
[tls]: #tls-block
[keepalive]: #keepalive-block
[sending_queue]: #sending_queue-block
[retry_on_failure]: #retry_on_failure-block
[debug_metrics]: #debug_metrics-block

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
`balancer_name` | `string` | Which gRPC client-side load balancer to use for requests. | `pick_first` | no
`authority` | `string` | Overrides the default `:authority` header in gRPC requests from the gRPC client. | | no
`auth` | `capsule(otelcol.Handler)` | Handler from an `otelcol.auth` component to use for authenticating requests. | | no

{{< docs/shared lookup="flow/reference/components/otelcol-compression-field.md" source="agent" version="<AGENT VERSION>" >}}

{{< docs/shared lookup="flow/reference/components/otelcol-grpc-balancer-name.md" source="agent" version="<AGENT VERSION>" >}}

{{< docs/shared lookup="flow/reference/components/otelcol-grpc-authority.md" source="agent" version="<AGENT VERSION>" >}}

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

Because `otelcol.exporter.otlp` uses gRPC, the configured proxy server must be
able to handle and proxy HTTP/2 traffic.

[HTTP CONNECT]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods/CONNECT

### tls block

The `tls` block configures TLS settings used for the connection to the gRPC
server.

{{< docs/shared lookup="flow/reference/components/otelcol-tls-config-block.md" source="agent" version="<AGENT VERSION>" >}}

> **NOTE**: `otelcol.exporter.otlp` uses gRPC, which does not allow you to send sensitive credentials (like `auth`) over insecure channels.
> Sending sensitive credentials over insecure non-TLS connections is supported by non-gRPC exporters such as [otelcol.exporter.otlphttp][].

[otelcol.exporter.otlphttp]: {{< relref "./otelcol.exporter.otlphttp.md" >}}

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

{{< docs/shared lookup="flow/reference/components/otelcol-queue-block.md" source="agent" version="<AGENT VERSION>" >}}

### retry_on_failure block

The `retry_on_failure` block configures how failed requests to the gRPC server are
retried.

{{< docs/shared lookup="flow/reference/components/otelcol-retry-block.md" source="agent" version="<AGENT VERSION>" >}}

### debug_metrics block

{{< docs/shared lookup="flow/reference/components/otelcol-debug-metrics-block.md" source="agent" version="<AGENT VERSION>" >}}

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

## Debug metrics

* `exporter_sent_spans_ratio_total` (counter): Number of spans successfully sent to destination.
* `exporter_send_failed_spans_ratio_total` (counter): Number of spans in failed attempts to send to destination.

## Examples

The following examples show you how to create an exporter to send data to different destinations.

### Send data to a local Tempo instance

You can create an exporter that sends your data to a local Grafana Tempo instance without TLS:

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

### Send data to a managed service

You can create an `otlp` exporter that sends your data to a managed service, for example, Grafana Cloud. The Tempo username and Grafana Cloud API Key are injected in this example through environment variables.

```river
otelcol.exporter.otlp "grafana_cloud_tempo" {
    client {
        endpoint = "https://tempo-xxx.grafana.net/tempo"
        auth     = otelcol.auth.basic.grafana_cloud_tempo.handler
    }
}
otelcol.auth.basic "grafana_cloud_tempo" {
    username = env("TEMPO_USERNAME")
    password = env("GRAFANA_CLOUD_API_KEY")
}
```
