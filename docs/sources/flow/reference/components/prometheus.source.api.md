---
title: prometheus.source.api
---

# prometheus.source.api

`prometheus.source.api` listens for HTTP requests containing Prometheus metric samples and forwards them to other components capable of receiving metrics.

The HTTP API exposed is compatible with [Prometheus `remote_write` API][prometheus-remote-write-docs]. This means that other [`prometheus.remote_write`][prometheus.remote_write] components can be used as a client and send requests to `prometheus.source.api` which enables using the Agent as a proxy for prometheus metrics.

[prometheus.remote_write]: {{< relref "./prometheus.remote_write.md" >}}
[prometheus-remote-write-docs]: https://prometheus.io/docs/prometheus/latest/querying/api/#remote-write-receiver

## Usage

```river
prometheus.source.api "LABEL" {
  http {
    listen_address = "LISTEN_ADDRESS"
    listen_port = PORT 
  }
  forward_to = RECEIVER_LIST
}
```

The component will start an HTTP server supporting the following endpoint:

- `POST /api/v1/metrics/write` - send metrics to the component, which in turn will be forwarded to the receivers as configured in `forward_to` argument. The request format must match that of [Prometheus `remote_write` API][prometheus-remote-write-docs]. One way to send valid requests to this component is to use another Grafana Agent with a [`prometheus.remote_write`][prometheus.remote_write] component.

## Arguments

`prometheus.source.api` supports the following arguments:

 Name         | Type             | Description                           | Default | Required 
--------------|------------------|---------------------------------------|---------|----------
 `forward_to` | `list(receiver)` | List of receivers to send metrics to. |         | yes      

## Blocks

The following blocks are supported inside the definition of `prometheus.source.api`:

 Hierarchy | Name     | Description                                        | Required 
-----------|----------|----------------------------------------------------|----------
 `http`    | [http][] | Configures the HTTP server that receives requests. | no       

[http]: #http

### http

{{< docs/shared lookup="flow/reference/components/loki-server-http.md" source="agent" >}}

## Exported fields

`prometheus.source.api` does not export any fields.

## Component health

`prometheus.source.api` is reported as unhealthy if it is given an invalid configuration.

## Debug metrics

The following are some of the metrics that are exposed when this component is used. Note that the metrics include labels such as `status_code` where relevant, which can be used to measure request success rates.

* `prometheus_source_api_request_duration_seconds` (histogram): Time (in seconds) spent serving HTTP requests.
* `prometheus_source_api_request_message_bytes` (histogram): Size (in bytes) of messages received in the request.
* `prometheus_source_api_response_message_bytes` (histogram): Size (in bytes) of messages sent in response.
* `prometheus_source_api_tcp_connections` (gauge): Current number of accepted TCP connections.
* `agent_prometheus_fanout_latency` (histogram): Write latency for sending metrics to other components.
* `agent_prometheus_forwarded_samples_total` (counter): Total number of samples sent to downstream components.

## Example

This example creates a `prometheus.source.api` component which starts an HTTP server on `0.0.0.0` address and port `9999`. The server receives metrics and forwards them to a `prometheus.remote_write` component which writes these metrics to a specified HTTP endpoint.

```river
// Receives metrics over HTTP
prometheus.source.api "api" {
  http {
    listen_address = "0.0.0.0"
    listen_port = 9999 
  }
  forward_to = [prometheus.remote_write.local.receiver]
}

// Writes metrics to a specified address, e.g. cloud-hosted Prometheus instance
prometheus.remote_write "local" {
  endpoint {
    url = "http://my-cloud-prometheus-instance.com/api/prom/push"
  }
}
```

In order to send metrics to the `prometheus.source.api` component defined above, another Grafana Agent can run with the following configuration:

```river
// Collects metrics of localhost:12345
prometheus.scrape "agent_self" {
  targets = [
    {"__address__" = "localhost:12345", "job" = "agent"},
  ]
  forward_to = [prometheus.remote_write.local.receiver]
}

// Writes metrics to localhost:9999/api/v1/metrics/write - e.g. served by 
// the prometheus.source.api component from the example above.
prometheus.remote_write "local" {
  endpoint {
    url = "http://localhost:9999/api/v1/metrics/write"
  }  
}
```