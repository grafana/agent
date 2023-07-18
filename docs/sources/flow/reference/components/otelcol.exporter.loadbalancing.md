---
canonical: https://grafana.com/docs/agent/latest/flow/reference/components/otelcol.exporter.loadbalancing/
labels:
  stage: beta
title: otelcol.exporter.loadbalancing
---

# otelcol.exporter.loadbalancing

{{< docs/shared lookup="flow/stability/beta.md" source="agent" >}}

`otelcol.exporter.loadbalancing` accepts logs and traces from other `otelcol` components
and writes them over the network using the OpenTelemetry Protocol (OTLP) protocol. 

> **NOTE**: `otelcol.exporter.loadbalancing` is a wrapper over the upstream
> OpenTelemetry Collector `loadbalancing` exporter. Bug reports or feature requests will
> be redirected to the upstream repository, if necessary.

Multiple `otelcol.exporter.loadbalancing` components can be specified by giving them
different labels.

The decision which backend to use depends on the trace ID or the service name. 
The backend load doesn't influence the choice. Even though this load-balancer won't do 
round-robin balancing of the batches, the load distribution should be very similar among backends, 
with a standard deviation under 5% at the current configuration.

`otelcol.exporter.loadbalancing` is especially useful for backends configured with tail-based samplers
which choose a backend based on the view of the full trace.

When a list of backends is updated, some of the signals will be rerouted to different backends. 
Around R/N of the "routes" will be rerouted differently, where:

* A "route" is either a trace ID or a service name mapped to a certain backend.
* "R" is the total number of routes.
* "N" is the total number of backends.

This should be stable enough for most cases, and the larger the number of backends, the less disruption it should cause.

## Usage

```river
otelcol.exporter.loadbalancing "LABEL" {
  resolver {
    ...
  }
  protocol {
    otlp {
      client {}
    }
  }
}
```

## Arguments

`otelcol.exporter.loadbalancing` supports the following arguments:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`routing_key` | `string` | Routing strategy for load balancing. | `"traceID"` | no

The `routing_key` attribute determines how to route signals across endpoints. Its value could be one of the following:
* `"service"`: spans with the same `service.name` will be exported to the same backend.
This is useful when using processors like the span metrics, so all spans for each service are sent to consistent Agent instances 
for metric collection. Otherwise, metrics for the same services would be sent to different Agents, making aggregations inaccurate.
* `"traceID"`: spans belonging to the same traceID will be exported to the same backend.

## Blocks

The following blocks are supported inside the definition of
`otelcol.exporter.loadbalancing`:

Hierarchy | Block | Description | Required
--------- | ----- | ----------- | --------
resolver | [resolver][] | Configures discovering the endpoints to export to. | yes
resolver > static | [static][] | Static list of endpoints to export to. | no
resolver > dns | [dns][] | DNS-sourced list of endpoints to export to. | no
protocol | [protocol][] | Protocol settings. Only OTLP is supported at the moment. | no
protocol > otlp | [otlp][] | Configures an OTLP exporter. | no
protocol > otlp > client | [client][] | Configures the exporter gRPC client. | no
protocol > otlp > client > tls | [tls][] | Configures TLS for the gRPC client. | no
protocol > otlp > client > keepalive | [keepalive][] | Configures keepalive settings for the gRPC client. | no
protocol > otlp > queue | [queue][] | Configures batching of data before sending. | no
protocol > otlp > retry | [retry][] | Configures retry mechanism for failed requests. | no

The `>` symbol indicates deeper levels of nesting. For example, `resolver > static`
refers to a `static` block defined inside a `resolver` block.

[resolver]: #resolver-block
[static]: #static-block
[dns]: #dns-block
[protocol]: #protocol-block
[otlp]: #otlp-block
[client]: #client-block
[tls]: #tls-block
[keepalive]: #keepalive-block
[queue]: #queue-block
[retry]: #retry-block

### resolver block

The `resolver` block configures how to retrieve the endpoint to which this exporter will send data.

Inside the `resolver` block, either the [dns][] block or the [static][] block 
should be specified. If both `dns` and `static` are specified, `dns` takes precedence.

### static block

The `static` block configures a list of endpoints which this exporter will send data to.

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`hostnames` | `list(string)` | List of endpoints to export to. |  | yes

### dns block

The `dns` block periodically resolves an IP address via the DNS `hostname` attribute. This IP address 
and the port specified via the `port` attribute will then be used by the gRPC exporter 
as the endpoint to which to export data to.

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`hostname` | `string`   | DNS hostname to resolve. |  | yes
`interval` | `duration` | Resolver interval. | `"5s"` | no
`timeout`  | `duration` | Resolver timeout. | `"1s"`  | no
`port`     | `string`   | Port to be used with the IP addresses resolved from the DNS hostname. | `"4317"` | no

### protocol block

The `protocol` block configures protocol-related settings for exporting.
At the moment only the OTLP protocol is supported.

### otlp block

The `otlp` block configures OTLP-related settings for exporting.

### client block

The `client` block configures the gRPC client used by the component. 
The endpoints used by the client block are the ones from the `resolver` block

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
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

You can configure an HTTP proxy with the following environment variables:

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

Because `otelcol.exporter.loadbalancing` uses gRPC, the configured proxy server must be
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

### queue block

The `queue` block configures an in-memory buffer of batches before data is sent
to the gRPC server.

{{< docs/shared lookup="flow/reference/components/otelcol-queue-block.md" source="agent" >}}

### retry block

The `retry` block configures how failed requests to the gRPC server are
retried.

{{< docs/shared lookup="flow/reference/components/otelcol-retry-block.md" source="agent" >}}

## Exported fields

The following fields are exported and can be referenced by other components:

Name | Type | Description
---- | ---- | -----------
`input` | `otelcol.Consumer` | A value that other components can use to send telemetry data to.

`input` accepts `otelcol.Consumer` OTLP-formatted data for telemetry signals of these types:
* logs
* traces

## Component health

`otelcol.exporter.loadbalancing` is only reported as unhealthy if given an invalid
configuration.

## Debug information

`otelcol.exporter.loadbalancing` does not expose any component-specific debug
information.

## Example

This example accepts OTLP logs and traces over gRPC. It then sends them in a load-balanced 
way to "localhost:55690" or "localhost:55700".

```river
otelcol.receiver.otlp "default" {
    grpc {}
    output {
        traces  = [otelcol.exporter.loadbalancing.default.input]
        logs    = [otelcol.exporter.loadbalancing.default.input]
    }
}

otelcol.exporter.loadbalancing "default" {
    resolver {
        static {
            hostnames = ["localhost:55690", "localhost:55700"]
        }
    }
    protocol {
        otlp {
            client {}
        }
    }
}
```
