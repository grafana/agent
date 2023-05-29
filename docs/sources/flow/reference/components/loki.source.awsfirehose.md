---
title: loki.source.awsfirehose
---

# loki.source.awsfirehose

`loki.source.awsfirehose` receives log entries over HTTP from [AWS Firehose](https://docs.aws.amazon.com/firehose/latest/dev/what-is-this-service.html)
and forwards them to other `loki.*` components.

The HTTP API exposed is compatible with [Firehose HTTP Delivery API](https://docs.aws.amazon.com/firehose/latest/dev/httpdeliveryrequestresponse.html).
Since the API model that AWS Firehose uses to deliver data over HTTP, the same delivery stream can be used to ship data
from different origins such as:

- [AWS CloudWatch logs](https://docs.aws.amazon.com/firehose/latest/dev/writing-with-cloudwatch-logs.html)
- [AWS CloudWatch events](https://docs.aws.amazon.com/firehose/latest/dev/writing-with-cloudwatch-events.html)
- Custom data through [DirectPUT requests](https://docs.aws.amazon.com/firehose/latest/dev/writing-with-sdk.html)

## Usage

```river
loki.source.awsfirehose "LABEL" {
    http {
        listen_address = "LISTEN_ADDRESS"
        listen_port = PORT 
    }
    forward_to = RECEIVER_LIST
}
```

The component will start HTTP server on the configured port and address with the following endpoints:

- `/awsfirehose/api/v1/push` - accepting `POST` requests compatible with [AWS Firehose HTTP Specifications](https://docs.aws.amazon.com/firehose/latest/dev/httpdeliveryrequestresponse.html).

## Arguments

`loki.source.awsfirehose` supports the following arguments:

 Name                     | Type                 | Description                                                | Default | Required
--------------------------|----------------------|------------------------------------------------------------|---------|----------
 `forward_to`             | `list(LogsReceiver)` | List of receivers to send log entries to.                  |         | yes
 `use_incoming_timestamp` | `bool`               | Whether or not to use the timestamp received from request. | `false` | no
 `relabel_rules`          | `RelabelRules`       | Relabeling rules to apply on log entries.                  | `{}`    | no

The `relabel_rules` field can make use of the `rules` export value from a
[`loki.relabel`][loki.relabel] component to apply one or more relabeling rules to log entries before they're forwarded
to the list of receivers in `forward_to`.

[loki.relabel]: {{< relref "./loki.relabel.md" >}}

## Blocks

The following blocks are supported inside the definition of `loki.source.awsfirehose`:

 Hierarchy | Name     | Description                                        | Required
-----------|----------|----------------------------------------------------|----------
 `http`    | [http][] | Configures the HTTP server that receives requests. | no
 `grpc`    | [grpc][] | Configures the gRPC server that receives requests. | no

[http]: #http

[grpc]: #grpc

### http

{{< docs/shared lookup="flow/reference/components/loki-server-http.md" source="agent" >}}

### grpc

{{< docs/shared lookup="flow/reference/components/loki-server-grpc.md" source="agent" >}}

## Exported fields

`loki.source.awsfirehose` does not export any fields.

## Component health

`loki.source.awsfirehose` is only reported as unhealthy if given an invalid configuration.

## Debug metrics

The following are some of the metrics that are exposed when this component is used. Note that the metrics include labels
such as `status_code` where relevant, which can be used to measure request success rates.

- `loki_source_awsfirehose_request_errors` (counter): Count of errors while receiving a request.
- `loki_source_awsfirehose_record_errors` (counter): Count of errors while decoding an individual record.
- `loki_source_awsfirehose_records_received` (counter): Count of records received.
- `loki_source_awsfirehose_batch_size` (histogram): Size (in units) of number of records received per request.

## Example

This example starts an HTTP server on `0.0.0.0` address and port `9999`. The server receives log entries and forwards
them to a `loki.write` component. The `loki.write` component will send the logs to the specified loki instance using
basic auth credentials provided.

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

loki.source.awsfirehose "loki_fh_receiver" {
    http {
        listen_address = "0.0.0.0"
        listen_port = 9999
    }
    forward_to = [
        loki.write.local.receiver,
    ]
}
```
