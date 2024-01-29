---
aliases:
- /docs/grafana-cloud/agent/flow/reference/components/otelcol.receiver.awsfirehose/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/reference/components/otelcol.receiver.awsfirehose/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/reference/components/otelcol.receiver.awsfirehose/
- /docs/grafana-cloud/send-data/agent/flow/reference/components/otelcol.receiver.awsfirehose/
canonical: https://grafana.com/docs/agent/latest/flow/reference/components/otelcol.receiver.awsfirehose/
description: Learn about otelcol.receiver.awsfirehose
title: otelcol.receiver.awsfirehose
---

# otelcol.receiver.awsfirehose

`otelcol.receiver.awsfirehose` receives metrics from [Cloudwatch Metrics Streams](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/CloudWatch-Metric-Streams.html) using AWS Kinesis Data Firehose and forwards it to other `otelcol.*` components.

Set the output format of the Metrics Stream to JSON. Make sure the receiver is accessible by AWS on port 443. You can set the output format with a load balancer.

{{% admonition type="note" %}}
`otelcol.receiver.awsfirehose` is a wrapper over the upstream OpenTelemetry Collector `awsfirehose` receiver. If necessary, bug reports or feature requests are redirected to the upstream repository.
{{% /admonition %}}

You can specify multiple `otelcol.receiver.awsfirehose` components by giving them different labels.

## Usage

```river
otelcol.receiver.awsfirehose "LABEL" {
  endpoint = "HOST:PORT"
  output {
    metrics = [...]
  }
}
```

## Arguments

`otelcol.receiver.awsfirehose` supports the following arguments:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`record_type` | `string` | The type of record received from the delivery stream. | `cwmetrics` | no
`access_key` | `string` | The access key to be checked on each request received. | | no
`endpoint` | `string` | `host:port` to listen for traffic on. | `"0.0.0.0:4433"` | no
`max_request_body_size` | `string` | Maximum request body size the HTTP server will allow. No limit when unset. | | no
`include_metadata` | `boolean` | Propagate incoming connection metadata to downstream consumers. | | no


The supported values for `record_type`  are:
* `cwmetrics`: The record type for the CloudWatch metric stream. Expects the format for the records to be JSON. See the [CloudWatch documentation][cloudwatch-metric-streams] for details.
[cloudwatch-metric-streams]: https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/CloudWatch-Metric-Streams.html

## Blocks

The following blocks are supported inside the definition of
`otelcol.receiver.awsfirehose`:

Hierarchy | Block | Description | Required
--------- | ----- | ----------- | --------
tls | [tls][] | Configures TLS for the HTTP server. | no
cors | [cors][] | Configures CORS for the HTTP server. | no
debug_metrics | [debug_metrics][] | Configures the metrics that this component generates to monitor its state. | no
output | [output][] | Configures where to send received traces. | yes

The `>` symbol indicates deeper levels of nesting. For example, `grpc > tls`
refers to a `tls` block defined inside a `grpc` block.

[tls]: #tls-block
[cors]: #cors-block
[debug_metrics]: #debug_metrics-block
[output]: #output-block

### tls block

The `tls` block configures TLS settings used for a server. If the `tls` block
isn't provided, TLS won't be used for connections to the server.

{{< docs/shared lookup="flow/reference/components/otelcol-tls-config-block.md" source="agent" version="<AGENT_VERSION>" >}}

### cors block

The `cors` block configures CORS settings for an HTTP server.

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`allowed_origins` | `list(string)` | Allowed values for the `Origin` header. | | no
`allowed_headers` | `list(string)` | Accepted headers from CORS requests. | `["X-Requested-With"]` | no
`max_age` | `number` | Configures the `Access-Control-Max-Age` response header. | | no

The `allowed_headers` argument specifies which headers are acceptable from a
CORS request. The following headers are always implicitly allowed:

* `Accept`
* `Accept-Language`
* `Content-Type`
* `Content-Language`

If `allowed_headers` includes `"*"`, all headers are permitted.

### debug_metrics block

{{< docs/shared lookup="flow/reference/components/otelcol-debug-metrics-block.md" source="agent" version="<AGENT_VERSION>" >}}

### output block

{{< docs/shared lookup="flow/reference/components/output-block.md" source="agent" version="<AGENT_VERSION>" >}}

## Exported fields

`otelcol.receiver.awsfirehose` does not export any fields.

## Component health

`otelcol.receiver.awsfirehose` is only reported as unhealthy if given an invalid
configuration.

## Debug information

`otelcol.receiver.awsfirehose` does not expose any component-specific debug
information.

## Example

This example forwards received metrics through a batch processor before finally
sending it to an OTLP-capable endpoint:

```river
otelcol.receiver.awsfirehose "default" {
  output {
    metrics = [otelcol.processor.batch.default.input]
  }
}

otelcol.processor.batch "default" {
  output {
    metrics = [otelcol.exporter.otlp.default.input]
    logs    = [otelcol.exporter.otlp.default.input]
    traces  = [otelcol.exporter.otlp.default.input]
  }
}

otelcol.exporter.otlp "default" {
  client {
    endpoint = env("OTLP_ENDPOINT")
  }
}
```
<!-- START GENERATED COMPATIBLE COMPONENTS -->

## Compatible components

`otelcol.receiver.awsfirehose` can accept arguments from the following components:

- Components that export [OpenTelemetry `otelcol.Consumer`]({{< relref "../compatibility/#opentelemetry-otelcolconsumer-exporters" >}})


{{% admonition type="note" %}}

Connecting some components may not be sensible or components may require further configuration to make the 
connection work correctly. Refer to the linked documentation for more details.

{{% /admonition %}}

<!-- END GENERATED COMPATIBLE COMPONENTS -->