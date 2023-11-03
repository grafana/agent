---
aliases:
- /docs/grafana-cloud/agent/flow/reference/components/loki.source.api/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/reference/components/loki.source.api/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/reference/components/loki.source.api/
canonical: https://grafana.com/docs/agent/latest/flow/reference/components/loki.source.api/
title: loki.source.api
description: Learn about loki.source.api
---

# loki.source.api

`loki.source.api` receives log entries over HTTP and forwards them to other `loki.*` components.

The HTTP API exposed is compatible with [Loki push API][loki-push-api] and the `logproto` format. This means that other [`loki.write`][loki.write] components can be used as a client and send requests to `loki.source.api` which enables using the Agent as a proxy for logs.

[loki.write]: {{< relref "./loki.write.md" >}}
[loki-push-api]: https://grafana.com/docs/loki/latest/api/#push-log-entries-to-loki

## Usage

```river
loki.source.api "LABEL" {
    http {
        listen_address = "LISTEN_ADDRESS"
        listen_port = PORT 
    }
    forward_to = RECEIVER_LIST
}
```

The component will start HTTP server on the configured port and address with the following endpoints:

- `/loki/api/v1/push` - accepting `POST` requests compatible with [Loki push API][loki-push-api], for example, from another Grafana Agent's [`loki.write`][loki.write] component.
- `/loki/api/v1/raw` - accepting `POST` requests with newline-delimited log lines in body. This can be used to send NDJSON or plaintext logs. This is compatible with promtail's push API endpoint - see [promtail's documentation][promtail-push-api] for more information. NOTE: when this endpoint is used, the incoming timestamps cannot be used and the `use_incoming_timestamp = true` setting will be ignored. 
- `/loki/ready` - accepting `GET` requests - can be used to confirm the server is reachable and healthy.
- `/api/v1/push` - internally reroutes to `/loki/api/v1/push` 
- `/api/v1/raw` - internally reroutes to `/loki/api/v1/raw`


[promtail-push-api]: /docs/loki/latest/clients/promtail/configuration/#loki_push_api

## Arguments

`loki.source.api` supports the following arguments:

 Name                     | Type                 | Description                                                | Default | Required 
--------------------------|----------------------|------------------------------------------------------------|---------|----------
 `forward_to`             | `list(LogsReceiver)` | List of receivers to send log entries to.                  |         | yes      
 `use_incoming_timestamp` | `bool`               | Whether or not to use the timestamp received from request. | `false` | no       
 `labels`                 | `map(string)`        | The labels to associate with each received logs record.    | `{}`    | no       
 `relabel_rules`          | `RelabelRules`       | Relabeling rules to apply on log entries.                  | `{}`    | no       

The `relabel_rules` field can make use of the `rules` export value from a
[`loki.relabel`][loki.relabel] component to apply one or more relabeling rules to log entries before they're forwarded to the list of receivers in `forward_to`.

[loki.relabel]: {{< relref "./loki.relabel.md" >}}

## Blocks

The following blocks are supported inside the definition of `loki.source.api`:

 Hierarchy | Name     | Description                                        | Required 
-----------|----------|----------------------------------------------------|----------
 `http`    | [http][] | Configures the HTTP server that receives requests. | no       

[http]: #http

### http

{{< docs/shared lookup="flow/reference/components/loki-server-http.md" source="agent" version="<AGENT VERSION>" >}}

## Exported fields

`loki.source.api` does not export any fields.

## Component health

`loki.source.api` is only reported as unhealthy if given an invalid configuration.

## Debug metrics

The following are some of the metrics that are exposed when this component is used. Note that the metrics include labels such as `status_code` where relevant, which can be used to measure request success rates.

* `loki_source_api_request_duration_seconds` (histogram): Time (in seconds) spent serving HTTP requests.
* `loki_source_api_request_message_bytes` (histogram): Size (in bytes) of messages received in the request.
* `loki_source_api_response_message_bytes` (histogram): Size (in bytes) of messages sent in response.
* `loki_source_api_tcp_connections` (gauge): Current number of accepted TCP connections.

## Example

This example starts an HTTP server on `0.0.0.0` address and port `9999`. The server receives log entries and forwards them to a `loki.write` component while adding a `forwarded="true"` label. The `loki.write` component will send the logs to the specified loki instance using basic auth credentials provided.

```river
loki.write "local" {
    endpoint {
        url = "http://loki:3100/api/v1/push"
        basic_auth {
            username = "<your username>"
            password_file = "<your password file>"
        }
    }
}

loki.source.api "loki_push_api" {
    http {
        listen_address = "0.0.0.0"
        listen_port = 9999
    }
    forward_to = [
        loki.write.local.receiver,
    ]
    labels = {
        forwarded = "true",
    }
}
```

