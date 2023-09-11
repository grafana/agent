---
canonical: https://grafana.com/docs/agent/latest/flow/reference/components/otelcol.processor.probabilistic_sampler/
labels:
  stage: opensource
title: otelcol.processor.probabilistic_sampler
---

# otelcol.processor.probabilistic_sampler

`otelcol.processor.probabilistic_sampler` accepts logs and traces data from other otelcol components and apply probabilistic sampling based on configuration options.

> **Note**: `otelcol.processor.probabilistic_sampler` is a wrapper over the upstream
> OpenTelemetry Collector Contrib `probabilistic_sampler` processor. Bug reports or feature
> requests will be redirected to the upstream repository, if necessary.

You can specify multiple `otelcol.processor.probabilistic_sampler` components by giving them
different labels.

## Usage

```river
otelcol.processor.probabilistic_sampler "LABEL" {
  output {
    logs    = [...]
    traces  = [...]
  }
}
```

## Usage for traces

The `probabilistic_sampler` supports two types of sampling for traces: 
1. `sampling.priority` [semantic
   convention](https://github.com/opentracing/specification/blob/master/semantic_conventions.md#span-tags-table) as defined by OpenTracing
2. Trace ID hashing

The `sampling.priority` semantic convention takes priority over trace ID hashing. 
Trace ID hashing samples based on hash values determined by trace IDs.

### Arguments

For traces `otelcol.processor.probabilistic_sampler` supports the following arguments:

Name | Type     | Description | Default | Required
---- |----------| ----------- |---------| --------
`hash_seed`               | `uint32` | An integer used to compute the hash algorithm. Note that all collectors for a given tier (e.g. behind the same load balancer) should have the same hash_seed. |         | no
`sampling_percentage`     | `float32`| Percentage at which traces are sampled; >= 100 samples all traces | `0`     | no

## Usage for logs

The probabilistic sampler supports sampling logs according to their trace ID, or by a specific log record attribute.

The probabilistic sampler optionally may use a `hash_seed` to compute the hash of a log record.
This sampler samples based on hash values determined by log records. In order for hashing to work, all collectors for a given tier (e.g. behind the same load balancer) must have the same `hash_seed`. It is also possible to leverage a different `hash_seed` at different collector tiers to support additional sampling requirements.

### Arguments

For logs `otelcol.processor.probabilistic_sampler` supports the following arguments:

Name | Type      | Description | Default | Required
---- |-----------| ----------- |---------| --------
`hash_seed`               | `uint32`  | An integer used to compute the hash algorithm. Note that all collectors for a given tier (e.g. behind the same load balancer) should have the same hash_seed. |         | no
`sampling_percentage`     | `float32` | Percentage at which traces are sampled; >= 100 samples all traces | `0`     | no
`attribute_source`        | `string`  | Defines where to look for the attribute in from_attribute. The allowed values are `traceID` or `record` | `"traceID"`  | no
`from_attribute`          | `string`  | The optional name of a log record attribute used for sampling purposes, such as a unique log record ID. The value of the attribute is only used if the trace ID is absent or if `attribute_source` is set to `record` | | no
`sampling_priority`       | `string`  | The optional name of a log record attribute used to set a different sampling priority from the `sampling_percentage` setting. 0 means to never sample the log record, and >= 100 means to always sample the log record | | no

## Exported fields

The following fields are exported and can be referenced by other components:

Name | Type | Description
---- | ---- | -----------
`input` | `otelcol.Consumer` | A value that other components can use to send telemetry data to.

`input` accepts `otelcol.Consumer` data for any telemetry signal (metrics,
logs, or traces).

## Component health

`otelcol.processor.probabilistic_sampler` is only reported as unhealthy if given an invalid
configuration.

## Debug information

`otelcol.processor.probabilistic_sampler` does not expose any component-specific debug
information.

## Example

```river
otelcol.processor.probabilistic_sampler "default" {
  hash_seed           = 123
  sampling_percentage = 15.3

  output {
    logs = [otelcol.exporter.otlp.default.input]
  }
}
```

Sample 15% of the logs:

```river
otelcol.processor.probabilistic_sampler "default" {
  sampling_percentage = 15

  output {
    logs = [otelcol.exporter.otlp.default.input]
  }
}
```

Sample logs according to their logID attribute:

```river
otelcol.processor.probabilistic_sampler "default" {
  sampling_percentage = 15
  attribute_source    = "record"
  from_attribute      = "logID"

  output {
    logs = [otelcol.exporter.otlp.default.input]
  }
}
```

Sample logs according to the attribute `priority`:

```river
otelcol.processor.probabilistic_sampler "default" {
  sampling_percentage = 15
  sampling_priority   = "priority"

  output {
    logs = [otelcol.exporter.otlp.default.input]
  }
}
```
