---
description: Learn how to Flow compares to other offerings
menuTitle: When to use Flow
title: When to use Flow
weight: 25
---

# When to use Flow

Grafana Agent Flow allows you to process telemetry in a variety of formats, natively:
* [OTLP][] (`otelcol` components)
* Prometheus (`prometheus` components)
* Loki (`loki` components)
* Pyroscope (`pyroscope` components)

Native processing of telemetry could lead to higher efficiency and ease of use.

It is also possible from switch to one format to another. For example:
* `otelcol.exporter.prometheus` converts OTLP metrics to Prometheus logs.
* `otelcol.receiver.prometheus` converts Prometheus metrics to OTLP metrics.
* `otelcol.exporter.loki` converts OTLP logs to Loki logs.
* `otelcol.receiver.loki` converts Loki logs to OTLP logs.
* `otelcol.connector.spanlogs` converts OTLP spans to OTLP logs.
* `otelcol.connector.spanmetrics` converts OTLP spans to OTLP metrics.

There are even Flow components which do not deal with telemetry. For example, `mimir.rules.kubernetes` 
can be used to configure a Mimir instance.

The articles in this section describe in more detail how Flow compares to similar products:

{{< section >}}

[OTLP]: https://grafana.com/docs/grafana-cloud/send-data/otlp/