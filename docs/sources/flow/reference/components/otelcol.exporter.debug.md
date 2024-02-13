---
aliases:
- /docs/grafana-cloud/agent/flow/reference/components/otelcol.exporter.debug/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/reference/components/otelcol.exporter.debug/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/reference/components/otelcol.exporter.debug/
- /docs/grafana-cloud/send-data/agent/flow/reference/components/otelcol.exporter.debug/
canonical: https://grafana.com/docs/agent/latest/flow/reference/components/otelcol.exporter.debug/
description: Learn about otelcol.exporter.debug
title: otelcol.exporter.debug
---

# otelcol.exporter.debug

`otelcol.exporter.debug` accepts telemetry data from other `otelcol` components
and writes them to the console (stderr). The verbosity of the logs can also be controlled.

> **NOTE**: `otelcol.exporter.debug` is a wrapper over the upstream
> OpenTelemetry Collector `debug` exporter. Bug reports or feature requests will
> be redirected to the upstream repository, if necessary.

Multiple `otelcol.exporter.debug` components can be specified by giving them
different labels.

## Usage

```river
otelcol.exporter.debug "LABEL" { }
```

## Arguments

`otelcol.exporter.debug` supports the following arguments:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`verbosity`           | `string` | Verbosity of the generated logs. | `"basic"` | no
`sampling_initial`    | `int`    | Number of messages initially logged each second. | `2` | no
`sampling_thereafter` | `int`    | Sampling rate after the initial messages are logged. | `500` | no

The `verbosity` argument must be one of `"basic"`, `"normal"`, or `"detailed"`.

## Exported fields

The following fields are exported and can be referenced by other components:

Name | Type | Description
---- | ---- | -----------
`input` | `otelcol.Consumer` | A value that other components can use to send telemetry data to.

`input` accepts `otelcol.Consumer` data for any telemetry signal (metrics,
logs, or traces).

## Component health

`otelcol.exporter.debug` is only reported as unhealthy if given an invalid
configuration.

## Debug information

`otelcol.exporter.debug` does not expose any component-specific debug
information.

## Example

This example scrapes prometheus unix metrics and writes them to the console:

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
