---
canonical: https://grafana.com/docs/agent/latest/flow/reference/components/otelcol.processor.spanlogs/
title: otelcol.processor.spanlogs
---

# otelcol.processor.spanlogs

`otelcol.processor.spanlogs` accepts telemetry data from other `otelcol`
components.

> **NOTE**: `otelcol.processor.spanlogs` is a custom component unrelated 
> to any processors from the OpenTelemetry Collector.

Multiple `otelcol.processor.spanlogs` components can be specified by giving them
different labels.

## Usage

```river
otelcol.processor.spanlogs "LABEL" {
  output {
    metrics = [...]
    logs    = [...]
    traces  = [...]
  }
}
```

## Arguments

`otelcol.processor.spanlogs` supports the following arguments:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`spans` | `bool` |  | `false` | no
`roots` | `bool` |  | `false` | no
`processes` | `bool` |  | `false` | no
`span_attributes` | `list(string)` |  | `[]` | no
`process_attributes` | `list(string)` |  | `[]` | no
`labels` | `list(string)` |  | `[]` | no

## Blocks

The following blocks are supported inside the definition of
`otelcol.processor.spanlogs`:

Hierarchy | Block | Description | Required
--------- | ----- | ----------- | --------
overrides | `list(string)` |  | no
output | [output][] | Configures where to send received telemetry data. | yes

[output]: #output-block

### output block

{{< docs/shared lookup="flow/reference/components/output-block-logs.md" source="agent" >}}

## Exported fields

The following fields are exported and can be referenced by other components:

Name | Type | Description
---- | ---- | -----------
`input` | `otelcol.Consumer` | A value that other components can use to send telemetry data to.

`input` accepts `otelcol.Consumer` data for any telemetry signal (metrics,
logs, or traces).

## Component health

`otelcol.processor.spanlogs` is only reported as unhealthy if given an invalid
configuration.

## Debug information

`otelcol.processor.spanlogs` does not expose any component-specific debug
information.

## Examples


[otelcol.exporter.otlp]: {{< relref "./otelcol.exporter.otlp.md" >}}
