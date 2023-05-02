---
title: loki.source.api
---

# loki.source.api

`loki.source.api` listens for HTTP requests containing Loki log entries and forwards them to other components capable
of receiving log entries.

The HTTP API exposed is compatible with [Loki push API][loki-push-api], using Loki's `logproto` format. This means that
other [`loki.write`][loki.write] components can be used as a client and send requests to `loki.source.api`. This enables
Grafana Agents to send log entries over the network to other Grafana Agents.

[loki.write]: {{< relref "./loki.write.md" >}}
[loki-push-api]: https://grafana.com/docs/loki/latest/api/#push-log-entries-to-loki

## Usage

```river
loki.source.api LABEL {
    http_address = "LISTEN_ADDRESS"
    http_port = PORT
    forward_to = RECEIVER_LIST
}
```

## Arguments

`loki.source.api` supports the following arguments:

 Name                     | Type                 | Description                                                | Default | Required 
--------------------------|----------------------|------------------------------------------------------------|---------|----------
 `http_port`              | `int`                | The host port for the HTTP server to listen on.            |         | yes      
 `http_address`           | `string`             | The host address for the HTTP server to listen on.         |         | yes      
 `forward_to`             | `list(LogsReceiver)` | List of receivers to send log entries to.                  |         | yes      
 `use_incoming_timestamp` | `bool`               | Whether or not to use the timestamp received from request. | `false` | no       
 `labels`                 | `map(string)`        | The labels to associate with each received logs record.    | `{}`    | no       
 `relabel_rules`          | `RelabelRules`       | Relabeling rules to apply on log entries.                  | `{}`    | no       

The `relabel_rules` field can make use of the `rules` export value from a
[`loki.relabel`][loki.relabel] component to apply one or more relabeling rules to log entries
before they're forwarded to the list of receivers in `forward_to`.

[loki.relabel]: {{< relref "./loki.relabel.md" >}}

## Exported fields

`loki.source.api` does not export any fields.

## Component health

`loki.source.api` is only reported as unhealthy if given an invalid configuration.

## Debug Metrics

The following are some of the metrics that are exposed when this component is used. Note that the metrics include labels
such as `status_code` where relevant, which can be used to measure request success rates.

* `loki_source_api_request_duration_seconds` (histogram): Time (in seconds) spent serving HTTP requests.
* `loki_source_api_request_message_bytes` (histogram): Size (in bytes) of messages received in the request.
* `loki_source_api_response_message_bytes` (histogram): Size (in bytes) of messages sent in response.
* `loki_source_api_tcp_connections` (gauge): Current number of accepted TCP connections.

## Example

This example starts an HTTP server on `localhost` address and port `9999`. The server receives log entries and forwards
them to a `loki.echo` component while
adding a `forwarded="true"` label.

```river
loki.echo "print" {}

loki.source.api "loki_push_api" {
    http_address = "0.0.0.0"
    http_port = 9999
    forward_to = [loki.echo.print.receiver]
    labels = {
        forwarded = "true",
    }
}
```

