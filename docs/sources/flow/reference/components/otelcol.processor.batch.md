---
aliases:
- /docs/agent/latest/flow/reference/components/otelcol.processor.batch
title: otelcol.processor.batch
---

# otelcol.processor.batch

`otelcol.processor.batch` accepts telemetry data from other `otelcol`
components and places them into batches. Batching improves the compression of
data and reduces the number of outgoing network requests required to transmit
data.

> **NOTE**: `otelcol.processor.batch` is a wrapper over the upstream
> OpenTelemetry Collector `batch` processor. Bug reports or feature requests
> may be redirected to the upstream repository.

Multiple `otelcol.processor.batch` components can be specified by giving them
different labels.

## Usage

```river
otelcol.processor.batch "LABEL" {
  output {
    metrics = [...]
    logs    = [...]
    traces  = [...]
  }
}
```

## Arguments

`otelcol.processor.batch` supports the following arguments:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`timeout` | `duration` | How long to wait before flushing the batch. | `"200ms"` | no
`send_batch_size` | `number` | Amount of data to buffer before flushing the batch. | `8192` | no
`send_batch_max_size` | `number` | Upper limit of a batch size. | `0` | no

`otelcol.processor.batch` accumulates data into a batch until one of the
following events happens:

* The duration specified by `timeout` elapses since the time the last batch was
  sent.

* The number of spans, log lines, or metric samples processed goes above the
  number specified by `send_batch_size`.

`send_batch_max_size` can be used to limit the amount of data contained in a
single batch. When set to `0`, batches are allowed to be any size.

For example, assume `send_batch_size` is set to the default `8192` and there
are currently 8000 batched spans. If 8000 more spans are received at once, it
would bring the total batch size to 16,192, which would then be flushed as a
single batch. `send_batch_max_size` allows to constrain how big a batch can
get. When set to a non-zero value, `send_batch_max_size` must be greater or
equal to `send_batch_size`.

## Blocks

The following blocks are supported inside the definition of
`otelcol.processor.batch`:

Hierarchy | Block | Description | Required
--------- | ----- | ----------- | --------
output | [output][] | Configures where to send received telemetry data. | **yes**

[output]: #output-block

### output block

The `output` block configures a set of components to send batched telemetry
data to.

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`metrics` | `list(otelcol.Consumer)` | List of consumers to send metrics to. | `[]` | no
`logs` | `list(otelcol.Consumer)` | List of consumers to send logs to. | `[]` | no
`traces` | `list(otelcol.Consumer)` | List of consumers to send traces to. | `[]` | no

The `output` block must be specified, but all of its arguments are optional. By
default, telemetry data will be dropped. To send telemetry data to other
components, configure the `metrics`, `logs`, and `traces` arguments
accordingly.

## Exported fields

The following fields are exported and can be referenced by other components:

Name | Type | Description
---- | ---- | -----------
`input` | `otelcol.Consumer` | A value which other components can use to send telemetry data to.

`input` accepts `otelcol.Consumer` data for any telemetry signal (metrics,
logs, or traces).

## Component health

`otelcol.processor.batch` is only reported as unhealthy if given an invalid
configuration.

## Debug information

`otelcol.processor.batch` does not expose any component-specific debug
information.
