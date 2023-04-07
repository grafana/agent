---
title: otelcol.exporter.logging
---

# otelcol.exporter.logging

`otelcol.exporter.logging` accepts telemetry data from other `otelcol` components
and writes them to the console.

This component writes logs at the info level. The [logging config block][] must be
configured to write logs at the info level.

> **NOTE**: `otelcol.exporter.logging` is a wrapper over the upstream
> OpenTelemetry Collector `logging` exporter. Bug reports or feature requests will
> be redirected to the upstream repository, if necessary.

Multiple `otelcol.exporter.logging` components can be specified by giving them
different labels.

[logging config block]: {{< relref "../config-blocks/logging.md" >}}

## Usage

```river
otelcol.exporter.logging "LABEL" { }
```

## Arguments

`otelcol.exporter.logging` supports the following arguments:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`verbosity`           | `string` | Verbosity of the generated logs. | `"normal"` | no
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

`otelcol.exporter.logging` is only reported as unhealthy if given an invalid
configuration.

## Debug information

`otelcol.exporter.logging` does not expose any component-specific debug
information.

## Example

This example creates an exporter to write traces directly to the console:

```river
prometheus.exporter.unix { }

prometheus.scrape "default" {
	targets    = prometheus.exporter.unix.targets
	forward_to = [otelcol.receiver.prometheus.default.receiver]
}

otelcol.receiver.prometheus "default" {
	output {
		metrics = [otelcol.exporter.logging.default.input]
	}
}

otelcol.exporter.logging "default" {
	verbosity           = "detailed"
	sampling_initial    = 1
	sampling_thereafter = 1
}
```
