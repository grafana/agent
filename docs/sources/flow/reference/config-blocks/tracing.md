---
aliases:
- /docs/agent/latest/flow/reference/config-blocks/tracing
title: tracing
weight: 100
---

# tracing block

`tracing` is an optional configuration block used to customize how Grafana Agent
produces traces. `tracing` is specified without a label and can only be provided
once per configuration file.

## Example

```river
tracing {
  sampling_fraction = 0.1

  write_to = [otelcol.exporter.otlp.tempo.input]
}

otelcol.exporter.otlp "tempo" {
  // Send traces to a locally running Tempo without TLS enabled.
  client {
    endpoint = env("TEMPO_OTLP_ENDPOINT")

    tls {
      insecure = true
    }
  }
}
```

## Arguments

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`sampling_fraction` | `number` | Fraction of traces to keep. | `0.1` | no
`write_to` | `list(otelcol.Consumer)` | Inputs from `otelcol` components to send traces to. | `[]` | no

The `write_to` argument controls which components to send traces to for
processing. The elements in the array can be any `otelcol` component which
accept traces, including processors and exporters. When `write_to` is set
to an empty array `[]`, all traces are dropped.

> **NOTE**: Any traces generated before the `tracing` block has been evaluated,
> such as at the early start of the process' lifetime, are dropped.

The `sampling_fraction` argument controls what percentage of generated traces
should be sent to the consumers specified by `write_to`. When set to `1` or
greater, 100% of traces are kept. When set to `0` or lower, 0% of traces are
kept.

