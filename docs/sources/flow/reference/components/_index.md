---
aliases:
- /docs/grafana-cloud/agent/flow/reference/components/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/reference/components/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/reference/components/
- /docs/grafana-cloud/send-data/agent/flow/reference/components/
canonical: https://grafana.com/docs/agent/latest/flow/reference/components/
description: Learn about the components in Grafana Agent Flow
title: Components reference
weight: 300
---

# Components reference

This section contains reference documentation for all recognized [components][].

## Which component do I need?

### Receiving telemetry

##### Logs

* In the Loki format: `loki.source` components such as `loki.source.file`.
  * If you need to convert from the Loki format to OTLP, you can use `otelcol.receiver.loki`.
* In the OTLP format: some `otelcol.receiver` components, such as `otelcol.receiver.otlp`.

##### Metrics

`prometheus.scrape` scrapes HTTP endpoints using the Prometheus exposition format.

`prometheus.receive_http` accepts metrics in the Prometheus format on a HTTP server.
  A separate process such as {{< param "PRODUCT_ROOT_NAME" >}}, Prometheus, or OpenTelemetry Collector could remote write ("push") metrics on this server.

##### Traces

* In the OTLP format: some `otelcol.receiver` components such as `otelcol.receiver.otlp`.

##### Profiles

`pyroscope.scrape` scrapes profiles from HTTP endpoints using the Pprof format.

### Processing telemetry

#### Logs

All `loki.source` components can receive logs, in the Loki format. If you need to convert from the Loki format to OTLP, you can use `otelcol.receiver.loki`.

##### Metrics

* In the Prometheus format: `prometheus.relabel`.
* In the OTLP format: `otelcol.processor` components such as `otelcol.processor.transform`.

##### Traces

* In the OTLP format: `otelcol.processor` components such as `otelcol.processor.transform`.

##### Profiles

At this time, profiles can only be received and sent by {{< param "PRODUCT_ROOT_NAME" >}}.

### Sending telemetry

##### Logs

All `loki.source` components can receive logs, in the Loki format. If you need to convert from the Loki format to OTLP, you can use `otelcol.receiver.loki`.

Some `otelcol.receiver` components such as `otelcol.receiver.otlp` can also receive logs, in the OTLP format.

##### Metrics

In the Prometheus format: `prometheus.remote_write`.

In the OTLP format: `otelcol.exporter` components such as `otelcol.exporter.otlp`.

##### Traces

* In the OTLP format: `otelcol.exporter` components such as `otelcol.exporter.otlp`.

##### Profiles

`pyroscope.write` sends profiles using Pyroscope's Push API.

### Converting one telemetry type to another

`loki.process` can create metrics out of logs using the `stage.metrics` block.

`otelcol.connector` components generally output a different type of telemetry from the one that they input.
For example, `otelcol.connector.spanmetrics` creates RED OTLP metrics from OTLP traces.

Converting metrics from the native Prometheus format to OTLP can be done using `otelcol.receiver.prometheus`.
To convert OTLP metrics to Prometheus, use `otelcol.exporter.prometheus`.

Converting logs from the native Loki format to OTLP can be done using `otelcol.receiver.loki`.
To convert OTLP metrics to Prometheus, use `otelcol.exporter.loki`.

### Configuring {{< param "PRODUCT_ROOT_NAME" >}} dynamically 

The Agent can receive part of all of its configuration dynamically, via `module` components such as `module.file`.
`module.git` pulls AGent configuration from a Git repository, whereas `module.http` pulls it from an HTTP endpoint.

### Other components

`mimir.rules.kubernetes` discovers `PrometheusRule` Kubernetes resources and loads them into a Mimir instance.

## All components

{{< section >}}

[components]: {{< relref "../../concepts/components.md" >}}
