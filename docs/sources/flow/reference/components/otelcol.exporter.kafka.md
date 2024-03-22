---
canonical: https://grafana.com/docs/agent/latest/flow/reference/components/otelcol.exporter.kafka/
description: Learn about otelcol.exporter.kafka
title: otelcol.exporter.kafka
---

# otelcol.exporter.kafka

`otelcol.exporter.kafka` accepts telemetry data from other `otelcol` components
and writes them over to a Kafka topic.

> **NOTE**: `otelcol.exporter.kafka` is a wrapper over the upstream
> OpenTelemetry Collector `kafka` exporter. Bug reports or feature requests will
> be redirected to the upstream repository, if necessary.

This component uses a synchronous producer that blocks and does not batch
messages, therefore it should be used with batch and queued retry processors
prepended to it, for higher throughput and resiliency.

Multiple `otelcol.exporter.kafka` components can be specified by giving them
different labels.

## Usage

```river
otelcol.exporter.kafka "LABEL" {
  protocol_version = "2.0.0"
}
```

## Arguments

`otelcol.exporter.kafka` supports the following arguments:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`protocol_version` | `duration` | Kafka protocol version to use. | | yes
`brokers` | list(string) | Kafka brokers to connect to. | `["localhost:9092"]` | no
`client_id` | string | The client ID to configure the Sarama Kafka client with for all produce requests. | `"sarama"` | no
`topic` | string | The name of the Kafka topic to export to. | see below | no
`encoding` | string | The encoding of data sent to Kafka. | see below | no
`timeout`  | duration | The timeout to use for sending data. | `"5s"` | no
`partition_traces_by_id` | bool | Whether to include trace IDs as the message key in messages sent to Kafka. | false | no
`resolve_canonical_bootstrap_servers_only` | bool | Whether to resolve then reverse-lookup broker IPs during startup. | false | no

The `encoding` argument determines how to export messages to Kafka.
`encoding` must be one of the following strings:

* `"otlp_proto"`: Encode messages as OTLP protobuf.
* `"otlp_json"`: Encode messages as OTLP JSON.
* `"jaeger_proto"`: Encode messages as a single Jaeger protobuf span.
* `"jaeger_json"`: Encode messages as a single Jaeger JSON span.
* `"zipkin_proto"`: Encode messages as a list of Zipkin protobuf spans.
* `"zipkin_json"`: Encode messages as a list of Zipkin JSON spans.
* `"raw"`: Copy the log message bytes into the body of a log record.

`"otlp_proto"` or `"otlp_json"` must be used to send all telemetry types to
Kafka; other encodings are signal-specific.

## Blocks

The following blocks are supported inside the definition of
`otelcol.exporter.otlp`:

Hierarchy | Block | Description | Required
--------- | ----- | ----------- | --------
authentication | [authentication][] | Configures authentication for connecting to Kafka brokers. | no
authentication > plaintext | [plaintext][] | Authenticates against Kafka brokers with plaintext. | no
authentication > sasl | [sasl][] | Authenticates against Kafka brokers with SASL. | no
authentication > tls | [tls][] | Configures TLS for connecting to the Kafka brokers. | no
authentication > kerberos | [kerberos][] | Authenticates against Kafka brokers with Kerberos. | no
metadata | [metadata][] | Configures how to retrieve metadata from Kafka brokers. | no
metadata > retry | [retry][] | Configures how to retry metadata retrieval. | no
sending_queue | [sending_queue][] | Configures batching of data before sending. | no
retry_on_failure | [retry_on_failure][] | Configures retry mechanism for failed requests. | no
debug_metrics | [debug_metrics][] | Configures the metrics that this component generates to monitor its state. | no

The `>` symbol indicates deeper levels of nesting. For example, `client > tls`
refers to a `tls` block defined inside a `client` block.

[authentication]: #authentication-block
[plaintext]: #plaintext-block
[sasl]: #sasl-block
[tls]: #tls-block
[kerberos]: #kerberos-block
[metadata]: #metadata-block
[sending_queue]: #sending_queue-block
[retry_on_failure]: #retry_on_failure-block
[debug_metrics]: #debug_metrics-block

### authentication block

The `authentication` block holds the definition of different authentication
mechanisms to use when connecting to Kafka brokers. It doesn't support any
arguments and is configured fully through inner blocks.

### plaintext block

The `plaintext` block configures `PLAIN` authentication against Kafka brokers.

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`username` | `string` | Username to use for `PLAIN` authentication. | | yes
`password` | `secret` | Password to use for `PLAIN` authentication. | | yes

### sasl block

The `sasl` block configures SASL authentication against Kafka brokers.

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`username` | `string` | Username to use for SASL authentication. | | yes
`password` | `secret` | Password to use for SASL authentication. | | yes
`mechanism` | `string` | SASL mechanism to use when authenticating. | | yes
`version` | `number` | Version of the SASL Protocol to use when authenticating. | `0` | no

The `mechanism` argument can be set to one of the following strings:

* `"PLAIN"`
* `"AWS_MSK_IAM"`
* `"SCRAM-SHA-256"`
* `"SCRAM-SHA-512"`

When `mechanism` is set to `"AWS_MSK_IAM"`, the [`aws_msk` child block][aws_msk] must also be provided.

The `version` argument can be set to either `0` or `1`.

### tls block

The `tls` block configures TLS settings used for connecting to the Kafka
brokers. If the `tls` block isn't provided, TLS won't be used for
communication.

{{< docs/shared lookup="flow/reference/components/otelcol-tls-config-block.md" source="agent" version="<AGENT_VERSION>" >}}

### kerberos block

The `kerberos` block configures Kerberos authentication against the Kafka
broker.

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`service_name` | `string` | Kerberos service name. | | no
`realm` | `string` | Kerberos realm. | | no
`use_keytab` | `string` | Enables using keytab instead of password. | | no
`username` | `string` | Kerberos username to authenticate as. | | yes
`password` | `secret` | Kerberos password to authenticate with. | | no
`config_file` | `string` | Path to Kerberos location (for example, `/etc/krb5.conf`). | | no
`keytab_file` | `string` | Path to keytab file (for example, `/etc/security/kafka.keytab`). | | no

When `use_keytab` is `false`, the `password` argument is required. When
`use_keytab` is `true`, the file pointed to by the `keytab_file` argument is
used for authentication instead. At most one of `password` or `keytab_file`
must be provided.

### metadata block

The `metadata` block configures how to retrieve and store metadata from the
Kafka broker.

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`include_all_topics` | `bool` | When true, maintains metadata for all topics. | `true` | no

If the `include_all_topics` argument is `true`, `otelcol.receiver.kafka`
maintains a full set of metadata for all topics rather than the minimal set
that has been necessary so far. Including the full set of metadata is more
convenient for users but can consume a substantial amount of memory if you have
many topics and partitions.

Retrieving metadata may fail if the Kafka broker is starting up at the same
time as the `otelcol.receiver.kafka` component. The [`retry` child
block][retry] can be provided to customize retry behavior.

### retry block

The `retry` block configures how to retry retrieving metadata when retrieval
fails.

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`max_retries` | `number` | How many times to reattempt retrieving metadata. | `3` | no
`backoff` | `duration` | Time to wait between retries. | `"250ms"` | no

### sending_queue block

The `sending_queue` block configures an in-memory buffer of batches before data is sent
to the gRPC server.

{{< docs/shared lookup="flow/reference/components/otelcol-queue-block.md" source="agent" version="<AGENT_VERSION>" >}}

### retry_on_failure block

The `retry_on_failure` block configures how failed requests to the gRPC server are
retried.

{{< docs/shared lookup="flow/reference/components/otelcol-retry-block.md" source="agent" version="<AGENT_VERSION>" >}}

### debug_metrics block

{{< docs/shared lookup="flow/reference/components/otelcol-debug-metrics-block.md" source="agent" version="<AGENT_VERSION>" >}}

## Exported fields

The following fields are exported and can be referenced by other components:

Name | Type | Description
---- | ---- | -----------
`input` | `otelcol.Consumer` | A value that other components can use to send telemetry data to.

`input` accepts `otelcol.Consumer` data for any telemetry signal (metrics,
logs, or traces).

## Component health

`otelcol.exporter.kafka` is only reported as unhealthy if given an invalid
configuration.

## Debug information

`otelcol.exporter.kafka` does not expose any component-specific debug
information.

## Debug metrics

* `exporter_sent_spans_ratio_total` (counter): Number of spans successfully sent to destination.
* `exporter_send_failed_spans_ratio_total` (counter): Number of spans in failed attempts to send to destination.
* `exporter_queue_capacity_ratio` (gauge): Fixed capacity of the retry queue (in batches)
* `exporter_queue_size_ratio` (gauge): Current size of the retry queue (in batches)
* `rpc_client_duration_milliseconds` (histogram): Measures the duration of inbound RPC.
* `rpc_client_request_size_bytes` (histogram): Measures size of RPC request messages (uncompressed).
* `rpc_client_requests_per_rpc` (histogram): Measures the number of messages received per RPC. Should be 1 for all non-streaming RPCs.
* `rpc_client_response_size_bytes` (histogram): Measures size of RPC response messages (uncompressed).
* `rpc_client_responses_per_rpc` (histogram): Measures the number of messages received per RPC. Should be 1 for all non-streaming RPCs.

## Examples

The following examples show you how to create an exporter to send data to different destinations.

### Send data to a local Kafka instance

You can create an exporter that sends your data to a local Kafka instance with no authentication:

```river
otelcol.exporter.kafka "default" {
  protocol_version = "2.0.0"
  brokers          = ["localhost:9092"]
}
```

### Batch OTLP data and send to a Kafka instance.

You can create forwards received telemetry data through a batch processor
before finally exporting it to a remote Kafka instance.

```river
otelcol.receiver.otlp "default" {
  http {}
  grpc {}

  output {
    metrics = [otelcol.processor.batch.default.input]
    logs    = [otelcol.processor.batch.default.input]
    traces  = [otelcol.processor.batch.default.input]
  }
}

otelcol.processor.batch "default" {
  output {
    metrics = [otelcol.exporter.kafka.default.input]
    logs    = [otelcol.exporter.kafka.default.input]
    traces  = [otelcol.exporter.kafka.default.input]
  }
}

otelcol.exporter.kafka "default" {
  protocol_version = "2.0.0"
  brokers = [
    "test1.foo.bar.kafka.us-east-1.amazonaws.com:9094",
    "test2.foo.bar.kafka.us-east-1.amazonaws.com:9094",
  ]

  authentication {
    plaintext {
      username = env("KAFKA_USERNAME")
      password = env("KAFKA_PASSWORD")
    }
  }
}
```

<!-- START GENERATED COMPATIBLE COMPONENTS -->

## Compatible components

`otelcol.exporter.kafka` has exports that can be consumed by the following components:

- Components that consume [OpenTelemetry `otelcol.Consumer`](../../compatibility/#opentelemetry-otelcolconsumer-consumers)

{{< admonition type="note" >}}
Connecting some components may not be sensible or components may require further configuration to make the connection work correctly.
Refer to the linked documentation for more details.
{{< /admonition >}}

<!-- END GENERATED COMPATIBLE COMPONENTS -->
