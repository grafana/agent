---
canonical: https://grafana.com/docs/agent/latest/flow/reference/components/otelcol.processor.batch/
title: otelcol.processor.batch
---

# otelcol.processor.batch

`otelcol.processor.batch` accepts telemetry data from other `otelcol`
components and places them into batches. Batching improves the compression of
data and reduces the number of outgoing network requests required to transmit
data. This processor supports both size and time based batching.

We strongly recommend that you configure the batch processor on every Agent that
uses OpenTelemetry (otelcol) Flow components.  The batch processor should be 
defined in the pipeline after the `otelcol.processor.memory_limiter` as well 
as any sampling processors. This is because batching should happen after any 
data drops such as sampling.

> **NOTE**: `otelcol.processor.batch` is a wrapper over the upstream
> OpenTelemetry Collector `batch` processor. Bug reports or feature requests
> will be redirected to the upstream repository, if necessary.

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
`metadata_keys` | `list(string)` | Creates a different batcher for each key/value combination of metadata. | `[]` | no
`metadata_cardinality_limit` | `number` | Limit of the unique metadata key/value combinations. | `1000` | no

`otelcol.processor.batch` accumulates data into a batch until one of the
following events happens:

* The duration specified by `timeout` elapses since the time the last batch was
  sent.

* The number of spans, log lines, or metric samples processed exceeds the
  number specified by `send_batch_size`.

Use `send_batch_max_size` to limit the amount of data contained in a single
batch. When set to `0`, batches can be any size.

For example, assume `send_batch_size` is set to the default `8192` and there
are currently 8000 batched spans. If the batch processor receives 8000 more
spans at once, the total batch size would be 16,192 which would then be flushed
as a single batch. `send_batch_max_size` constrains how big a batch can get.
When set to a non-zero value, `send_batch_max_size` must be greater or equal to
`send_batch_size`.

`metadata_cardinality_limit` applies for the lifetime of the process.

Receivers should be configured with `include_metadata = true` so that metadata 
keys are available to the processor.

Each distinct combination of metadata triggers the allocation of a new 
background task in the Agent that runs for the lifetime of the process, and each 
background task holds one pending batch of up to `send_batch_size` records. Batching 
by metadata can therefore substantially increase the amount of memory dedicated to batching.

The maximum number of distinct combinations is limited to the configured `metadata_cardinality_limit`, 
which defaults to 1000 to limit memory impact.

## Blocks

The following blocks are supported inside the definition of
`otelcol.processor.batch`:

Hierarchy | Block | Description | Required
--------- | ----- | ----------- | --------
output | [output][] | Configures where to send received telemetry data. | yes

[output]: #output-block

### output block

{{< docs/shared lookup="flow/reference/components/output-block.md" source="agent" >}}

## Exported fields

The following fields are exported and can be referenced by other components:

Name | Type | Description
---- | ---- | -----------
`input` | `otelcol.Consumer` | A value that other components can use to send telemetry data to.

`input` accepts `otelcol.Consumer` data for any telemetry signal (metrics,
logs, or traces).

## Component health

`otelcol.processor.batch` is only reported as unhealthy if given an invalid
configuration.

## Debug information

`otelcol.processor.batch` does not expose any component-specific debug
information.

## Examples

### Basic usage

This example batches telemetry data before sending it to
[otelcol.exporter.otlp][] for further processing:

```river
otelcol.processor.batch "default" {
  output {
    metrics = [otelcol.exporter.otlp.production.input]
    logs    = [otelcol.exporter.otlp.production.input]
    traces  = [otelcol.exporter.otlp.production.input]
  }
}

otelcol.exporter.otlp "production" {
  client {
    endpoint = env("OTLP_SERVER_ENDPOINT")
  }
}
```

### Batching based on metadata

Batching by metadata enables support for multi-tenant OpenTelemetry pipelines 
with batching over groups of data having the same authorization metadata.

```river
otelcol.receiver.jaeger "default" {
  protocols {
    grpc {
      include_metadata = true
    }
    thrift_http {}
    thrift_binary {}
    thrift_compact {}
  }

  output {
    traces = [otelcol.processor.batch.default.input]
  }
}

otelcol.processor.batch "default" {
  // batch data by tenant id
  metadata_keys = ["tenant_id"]
  // limit to 10 batcher processes before raising errors
  metadata_cardinality_limit = 123

  output {
    traces  = [otelcol.exporter.otlp.production.input]
  }
}

otelcol.exporter.otlp "production" {
  client {
    endpoint = env("OTLP_SERVER_ENDPOINT")
  }
}
```

[otelcol.exporter.otlp]: {{< relref "./otelcol.exporter.otlp.md" >}}
