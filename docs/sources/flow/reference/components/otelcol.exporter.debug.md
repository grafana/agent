---
aliases:
  - /docs/grafana-cloud/agent/flow/reference/components/otelcol.exporter.debug/
  - /docs/grafana-cloud/monitor-infrastructure/agent/flow/reference/components/otelcol.exporter.debug/
  - /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/reference/components/otelcol.exporter.debug/
  - /docs/grafana-cloud/send-data/agent/flow/reference/components/otelcol.exporter.debug/
canonical: https://grafana.com/docs/agent/latest/flow/reference/components/otelcol.exporter.debug/
description: Learn about otelcol.exporter.debug
labels:
  stage: experimental
title: otelcol.exporter.debug
---

# otelcol.exporter.debug

`otelcol.exporter.debug` accepts telemetry data from other `otelcol` components and writes them to the console (stderr).
You can control the verbosity of the logs.

{{< admonition type="note" >}}
`otelcol.exporter.debug` is a wrapper over the upstream OpenTelemetry Collector `debug` exporter.
If necessary, bug reports or feature requests are redirected to the upstream repository.
{{< /admonition >}}

Multiple `otelcol.exporter.debug` components can be specified by giving them different labels.

## Usage

```river
otelcol.exporter.debug "LABEL" { }
```

## Arguments

`otelcol.exporter.debug` supports the following arguments:

| Name                  | Type     | Description                                          | Default    | Required |
| --------------------- | -------- | ---------------------------------------------------- | ---------- | -------- |
| `verbosity`           | `string` | Verbosity of the generated logs.                     | `"normal"` | no       |
| `sampling_initial`    | `int`    | Number of messages initially logged each second.     | `2`        | no       |
| `sampling_thereafter` | `int`    | Sampling rate after the initial messages are logged. | `500`      | no       |

The `verbosity` argument must be one of `"basic"`, `"normal"`, or `"detailed"`.

## Exported fields

The following fields are exported and can be referenced by other components:

| Name    | Type               | Description                                                      |
| ------- | ------------------ | ---------------------------------------------------------------- |
| `input` | `otelcol.Consumer` | A value that other components can use to send telemetry data to. |

`input` accepts `otelcol.Consumer` data for any telemetry signal (metrics,
logs, or traces).

## Component health

`otelcol.exporter.debug` is only reported as unhealthy if given an invalid
configuration.

## Debug information

`otelcol.exporter.debug` does not expose any component-specific debug
information.

## Example

This example scrapes Prometheus UNIX metrics and writes them to the console:

```river
prometheus.exporter.unix "default" { }

prometheus.scrape "default" {
    targets    = prometheus.exporter.unix.default.targets
    forward_to = [otelcol.receiver.prometheus.default.receiver]
}

otelcol.receiver.prometheus "default" {
    output {
        metrics = [otelcol.exporter.debug.default.input]
    }
}

otelcol.exporter.debug "default" {
    verbosity           = "detailed"
    sampling_initial    = 1
    sampling_thereafter = 1
}
```

<!-- START GENERATED COMPATIBLE COMPONENTS -->

## Compatible components

`otelcol.exporter.debug` has exports that can be consumed by the following components:

- Components that consume [OpenTelemetry `otelcol.Consumer`](../../compatibility/#opentelemetry-otelcolconsumer-consumers)

{{< admonition type="note" >}}
Connecting some components may not be sensible or components may require further configuration to make the connection work correctly.
Refer to the linked documentation for more details.
{{< /admonition >}}

<!-- END GENERATED COMPATIBLE COMPONENTS -->
