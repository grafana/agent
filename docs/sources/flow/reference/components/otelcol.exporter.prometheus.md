---
aliases:
- /docs/grafana-cloud/agent/flow/reference/components/otelcol.exporter.prometheus/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/reference/components/otelcol.exporter.prometheus/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/reference/components/otelcol.exporter.prometheus/
- /docs/grafana-cloud/send-data/agent/flow/reference/components/otelcol.exporter.prometheus/
canonical: https://grafana.com/docs/agent/latest/flow/reference/components/otelcol.exporter.prometheus/
description: Learn about otelcol.exporter.prometheus
title: otelcol.exporter.prometheus
---

# otelcol.exporter.prometheus

`otelcol.exporter.prometheus` accepts OTLP-formatted metrics from other
`otelcol` components, converts metrics to Prometheus-formatted metrics,
and forwards the resulting metrics to `prometheus` components.

> **NOTE**: `otelcol.exporter.prometheus` is a custom component unrelated to the
> `prometheus` exporter from OpenTelemetry Collector.
>
> Conversion of metrics are done according to the OpenTelemetry
> [Metrics Data Model][] specification.

Multiple `otelcol.exporter.prometheus` components can be specified by giving them
different labels.

[Metrics Data Model]: https://opentelemetry.io/docs/reference/specification/metrics/data-model/

## Usage

```river
otelcol.exporter.prometheus "LABEL" {
  forward_to = [...]
}
```

## Arguments

`otelcol.exporter.prometheus` supports the following arguments:

Name | Type | Description                                               | Default | Required
---- | ---- |-----------------------------------------------------------| ------- | --------
`include_target_info` | `boolean` | Whether to include `target_info` metrics.                 | `true` | no
`include_scope_info` | `boolean` | Whether to include `otel_scope_info` metrics.             | `false` | no
`include_scope_labels` | `boolean` | Whether to include additional OTLP labels in all metrics. | `true` | no
`add_metric_suffixes` | `boolean` | Whether to add type and unit suffixes to metrics names.   | `true` | no
`gc_frequency` | `duration` | How often to clean up stale metrics from memory.          | `"5m"` | no
`forward_to` | `list(MetricsReceiver)` | Where to forward converted Prometheus metrics.            | | yes
`resource_to_telemetry_conversion` | `boolean` | Whether to convert OTel resource attributes to Prometheus labels. | `false` | no

By default, OpenTelemetry resources are converted into `target_info` metrics. 
OpenTelemetry instrumentation scopes are converted into `otel_scope_info`
metrics. Set the `include_scope_info` and `include_target_info` arguments to
`false`, respectively, to disable the custom metrics.

When `include_scope_labels` is `true`  the `otel_scope_name` and
`otel_scope_version` labels are added to every converted metric sample.

When `include_target_info` is true, OpenTelemetry Collector resources are converted into `target_info` metrics.

## Exported fields

The following fields are exported and can be referenced by other components:

Name | Type | Description
---- | ---- | -----------
`input` | `otelcol.Consumer` | A value that other components can use to send telemetry data to.

`input` accepts `otelcol.Consumer` data for metrics. Other telemetry signals are ignored.

Metrics sent to the `input` are converted to Prometheus-compatible metrics and
are forwarded to the `forward_to` argument.

The following are dropped during the conversion process:

* Metrics that use the delta aggregation temporality
* ExponentialHistogram data points

## Component health

`otelcol.exporter.prometheus` is only reported as unhealthy if given an invalid
configuration.

## Debug information

`otelcol.exporter.prometheus` does not expose any component-specific debug
information.

## Example

This example accepts metrics over OTLP and forwards it using
`prometheus.remote_write`:

```river
otelcol.receiver.otlp "default" {
  grpc {}

  output {
    metrics = [otelcol.exporter.prometheus.default.input]
  }
}

otelcol.exporter.prometheus "default" {
  forward_to = [prometheus.remote_write.mimir.receiver]
}

prometheus.remote_write "mimir" {
  endpoint {
    url = "http://mimir:9009/api/v1/push"
  }
}
```
<!-- START GENERATED COMPATIBLE COMPONENTS -->

## Compatible components

`otelcol.exporter.prometheus` can accept arguments from the following components:

- Components that export [Prometheus `MetricsReceiver`]({{< relref "../compatibility/#prometheus-metricsreceiver-exporters" >}})

`otelcol.exporter.prometheus` has exports that can be consumed by the following components:

- Components that consume [OpenTelemetry `otelcol.Consumer`]({{< relref "../compatibility/#opentelemetry-otelcolconsumer-consumers" >}})

{{% admonition type="note" %}}

Connecting some components may not be sensible or components may require further configuration to make the 
connection work correctly. Refer to the linked documentation for more details.

{{% /admonition %}}

<!-- END GENERATED COMPATIBLE COMPONENTS -->