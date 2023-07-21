---
canonical: https://grafana.com/docs/agent/latest/flow/reference/components/loki.write/
title: loki.write
---

# loki.write

`loki.write` receives log entries from other loki components and sends them
over the network using Loki's `logproto` format.

Multiple `loki.write` components can be specified by giving them
different labels.

## Usage

```river
loki.write "LABEL" {
  endpoint {
    url = REMOTE_WRITE_URL
  }
}
```

## Arguments

`loki.write` supports the following arguments:

Name              | Type          | Description                                      | Default | Required
----------------- | ------------- | ------------------------------------------------ | ------- | --------
`max_streams`     | `int`         | Time to wait before marking a request as failed. | `"5s"`  | no
`external_labels` | `map(string)` | Labels to add to logs sent over the network.     |         | no

## Blocks

The following blocks are supported inside the definition of
`loki.write`:

Hierarchy | Block | Description | Required
--------- | ----- | ----------- | --------
endpoint | [endpoint][] | Location to send logs to. | no
endpoint > basic_auth | [basic_auth][] | Configure basic_auth for authenticating to the endpoint. | no
endpoint > authorization | [authorization][] | Configure generic authorization to the endpoint. | no
endpoint > oauth2 | [oauth2][] | Configure OAuth2 for authenticating to the endpoint. | no
endpoint > oauth2 > tls_config | [tls_config][] | Configure TLS settings for connecting to the endpoint. | no
endpoint > tls_config | [tls_config][] | Configure TLS settings for connecting to the endpoint. | no

The `>` symbol indicates deeper levels of nesting. For example, `endpoint >
basic_auth` refers to a `basic_auth` block defined inside an
`endpoint` block.

[endpoint]: #endpoint-block
[basic_auth]: #basic_auth-block
[authorization]: #authorization-block
[oauth2]: #oauth2-block
[tls_config]: #tls_config-block

### endpoint block

The `endpoint` block describes a single location to send logs to. Multiple
`endpoint` blocks can be provided to send logs to multiple locations.

The following arguments are supported:

Name                  | Type          | Description                           | Default        | Required
--------------------- | ------------- | ------------------------------------- | -------------- | --------
`url`                 | `string`      | Full URL to send logs to. | | yes
`name`                | `string`      | Optional name to identify this endpoint with. | | no
`headers`             | `map(string)` | Extra headers to deliver with the request. | | no
`batch_wait`          | `duration`    | Maximum amount of time to wait before sending a batch. | `"1s"` | no
`batch_size`          | `string`      | Maximum batch size of logs to accumulate before sending. | `"1MiB"` | no
`remote_timeout`      | `duration`    | Timeout for requests made to the URL. | `"10s"` | no
`tenant_id`           | `string`      | The tenant ID used by default to push logs. | | no
`min_backoff_period`  | `duration`    | Initial backoff time between retries. | `"500ms"` | no
`max_backoff_period`  | `duration`    | Maximum backoff time between retries. | `"5m"` | no
`max_backoff_retries` | `int`         | Maximum number of retries. | 10 | no
`bearer_token`        | `secret`      | Bearer token to authenticate with. | | no
`bearer_token_file`   | `string`      | File containing a bearer token to authenticate with. | | no
`proxy_url`           | `string`      | HTTP proxy to proxy requests through. | | no
`follow_redirects`    | `bool`        | Whether redirects returned by the server should be followed. | `true` | no
`enable_http2`        | `bool`        | Whether HTTP2 is supported for requests. | `true` | no

 At most one of the following can be provided:
 - [`bearer_token` argument](#endpoint-block).
 - [`bearer_token_file` argument](#endpoint-block). 
 - [`basic_auth` block][basic_auth].
 - [`authorization` block][authorization].
 - [`oauth2` block][oauth2].

If no `tenant_id` is provided, the component assumes that the Loki instance at
`endpoint` is running in single-tenant mode and no X-Scope-OrgID header is
sent.

When multiple `endpoint` blocks are provided, the `loki.write` component 
creates a client for each. Received log entries are fanned-out to these clients
in succession. That means that if one client is bottlenecked, it may impact
the rest.

Endpoints can be named for easier identification in debug metrics by using the
`name` argument. If the `name` argument isn't provided, a name is generated
based on a hash of the endpoint settings.

### basic_auth block

{{< docs/shared lookup="flow/reference/components/basic-auth-block.md" source="agent" >}}

### authorization block

{{< docs/shared lookup="flow/reference/components/authorization-block.md" source="agent" >}}

### oauth2 block

{{< docs/shared lookup="flow/reference/components/oauth2-block.md" source="agent" >}}

### tls_config block

{{< docs/shared lookup="flow/reference/components/tls-config-block.md" source="agent" >}}

## Exported fields

The following fields are exported and can be referenced by other components:

Name | Type | Description
---- | ---- | -----------
`receiver` | `LogsReceiver` | A value that other components can use to send log entries to.

## Component health

`loki.write` is only reported as unhealthy if given an invalid
configuration.

## Debug information

`loki.write` does not expose any component-specific debug
information.

## Debug metrics
* `loki_write_encoded_bytes_total` (counter): Number of bytes encoded and ready to send.
* `loki_write_sent_bytes_total` (counter): Number of bytes sent.
* `loki_write_dropped_bytes_total` (counter): Number of bytes dropped because failed to be sent to the ingester after all retries.
* `loki_write_sent_entries_total` (counter): Number of log entries sent to the ingester.
* `loki_write_dropped_entries_total` (counter): Number of log entries dropped because they failed to be sent to the ingester after all retries.
* `loki_write_request_duration_seconds` (histogram): Duration of sent requests.
* `loki_write_batch_retries_total` (counter): Number of times batches have had to be retried.
* `loki_write_stream_lag_seconds` (gauge): Difference between current time and last batch timestamp for successful sends.

## Example

This example creates a `loki.write` component that sends received entries to a
local Loki instance:

```river
loki.write "local" {
    endpoint {
        url = "http://loki:3100/loki/api/v1/push"
    }
}
```

## Compression

`loki.write` uses [snappy](https://en.wikipedia.org/wiki/Snappy_(compression)) for compression.
