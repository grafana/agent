---
aliases:
- /docs/grafana-cloud/agent/flow/reference/components/otelcol.connector.spanmetrics/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/reference/components/otelcol.connector.spanmetrics/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/reference/components/otelcol.connector.spanmetrics/
canonical: https://grafana.com/docs/agent/latest/flow/reference/components/otelcol.connector.spanmetrics/
labels:
  stage: experimental
title: otelcol.connector.spanmetrics
description: Learn about otelcol.connector.spanmetrics
---

# otelcol.connector.spanmetrics

{{< docs/shared lookup="flow/stability/experimental.md" source="agent" version="<AGENT VERSION>" >}}

`otelcol.connector.spanmetrics` accepts span data from other `otelcol` components and
aggregates Request, Error and Duration (R.E.D) OpenTelemetry metrics from the spans:

- **Request** counts are computed as the number of spans seen per unique set of dimensions,
  including Errors. Multiple metrics can be aggregated if, for instance, a user wishes to
  view call counts just on `service.name`` and `span.name``.

- **Error** counts are computed from the Request counts which have an `Error` status code metric dimension.

- **Duration** is computed from the difference between the span start and end times and inserted
  into the relevant duration histogram time bucket for each unique set dimensions.

> **NOTE**: `otelcol.connector.spanmetrics` is a wrapper over the upstream
> OpenTelemetry Collector `spanmetrics` connector. Bug reports or feature requests
> will be redirected to the upstream repository, if necessary.

Multiple `otelcol.connector.spanmetrics` components can be specified by giving them
different labels.

## Usage

```river
otelcol.connector.spanmetrics "LABEL" {
  histogram {
    ...
  }

  output {
    metrics = [...]
  }
}
```

## Arguments

`otelcol.connector.spanmetrics` supports the following arguments:

| Name                      | Type       | Description                                             | Default        | Required |
| ------------------------- | ---------- | ------------------------------------------------------- | -------------- | -------- |
| `dimensions_cache_size`   | `number`   | How many dimensions to cache.                           | `1000`         | no       |
| `aggregation_temporality` | `string`   | Configures whether to reset the metrics after flushing. | `"CUMULATIVE"` | no       |
| `metrics_flush_interval`  | `duration` | How often to flush generated metrics.                   | `"15s"`        | no       |
| `namespace`               | `string`   | Metric namespace.                                       | `""`           | no       |
| `exclude_dimensions`      | `list(string)` | List of dimensions to be excluded from the default set of dimensions. | `false` | no |

Adjusting `dimensions_cache_size` can improve the Agent process' memory usage.

The supported values for `aggregation_temporality` are:

- `"CUMULATIVE"`: The metrics will **not** be reset after they are flushed.
- `"DELTA"`: The metrics will be reset after they are flushed.

If `namespace` is set, the generated metric name will be added a `namespace.` prefix.

## Blocks

The following blocks are supported inside the definition of
`otelcol.connector.spanmetrics`:

| Hierarchy               | Block           | Description                                             | Required |
| ----------------------- | --------------- | ------------------------------------------------------- | -------- |
| dimension               | [dimension][]   | Dimensions to be added in addition to the default ones. | no       |
| histogram               | [histogram][]   | Configures the histogram derived from spans durations.  | yes      |
| histogram > exponential | [exponential][] | Configuration for a histogram with exponential buckets. | no       |
| histogram > explicit    | [explicit][]    | Configuration for a histogram with explicit buckets.    | no       |
| exemplars               | [exemplars][]   | Configures how to attach exemplars to histograms.       | no       |
| output                  | [output][]      | Configures where to send telemetry data.                | yes      |

It is necessary to specify either a "[exponential][]" or an "[explicit][]" block:

- Specifying both an "[exponential][]" and an "[explicit][]" block is not allowed.
- Specifying neither an "[exponential][]" nor an "[explicit][]" block is not allowed.

[dimension]: #dimension-block
[histogram]: #histogram-block
[exponential]: #exponential-block
[explicit]: #explicit-block
[exemplars]: #exemplars-block
[output]: #output-block

### dimension block

The `dimension` block configures dimensions to be added in addition to the default ones.

The default dimensions are:

- `service.name`
- `span.name`
- `span.kind`
- `status.code`

The default dimensions are always added. If no additional dimensions are specified,
only the default ones will be added.

The following attributes are supported:

| Name      | Type     | Description                                      | Default | Required |
| --------- | -------- | ------------------------------------------------ | ------- | -------- |
| `name`    | `string` | Span attribute or resource attribute to look up. |         | yes      |
| `default` | `string` | Value to use if the attribute is missing.        | null    | no       |

`otelcol.connector.spanmetrics` will look for the `name` attribute in the span's
collection of attributes. If it is not found, the resource attributes will be checked.

If the attribute is missing in both the span and resource attributes:

- If `default` is not set, the dimension will be omitted.
- If `default` is set, the dimension will be added and its value will be set to the value of `default`.

### histogram block

The `histogram` block configures the histogram derived from spans' durations.

The following attributes are supported:

| Name   | Type     | Description                     | Default | Required |
| ------ | -------- | ------------------------------- | ------- | -------- |
| `unit` | `string` | Configures the histogram units. | `"ms"`  | no       |
| `disable`| `bool` | Disable all histogram metrics. | `false` | no |

The supported values for `aggregation_temporality` are:

- `"ms"`: milliseconds
- `"s"`: seconds

### exponential block

The `exponential` block configures a histogram with exponential buckets.

The following attributes are supported:

| Name       | Type     | Description                                                      | Default | Required |
| ---------- | -------- | ---------------------------------------------------------------- | ------- | -------- |
| `max_size` | `number` | Maximum number of buckets per positive or negative number range. | `160`   | no       |

### explicit block

The `explicit` block configures a histogram with explicit buckets.

The following attributes are supported:

| Name      | Type             | Description                | Default                                                                                                                      | Required |
| --------- | ---------------- | -------------------------- | ---------------------------------------------------------------------------------------------------------------------------- | -------- |
| `buckets` | `list(duration)` | List of histogram buckets. | `["2ms", "4ms", "6ms", "8ms", "10ms", "50ms", "100ms", "200ms", "400ms", "800ms", "1s", "1400ms", "2s", "5s", "10s", "15s"]` | no       |

### exemplars block

The `exemplars` block configures how to attach exemplars to histograms.

The following attributes are supported:

| Name       | Type     | Description                                                      | Default | Required |
| ---------- | -------- | ---------------------------------------------------------------- | ------- | -------- |
| `enabled`  | `bool`   | Configures whether to add exemplars to histograms.               | `false` | no       |

### output block

{{< docs/shared lookup="flow/reference/components/output-block-metrics.md" source="agent" version="<AGENT VERSION>" >}}

## Exported fields

The following fields are exported and can be referenced by other components:

| Name    | Type               | Description                                                      |
| ------- | ------------------ | ---------------------------------------------------------------- |
| `input` | `otelcol.Consumer` | A value that other components can use to send telemetry data to. |

`input` accepts `otelcol.Consumer` traces telemetry data. It does not accept metrics and logs.

## Component health

`otelcol.connector.spanmetrics` is only reported as unhealthy if given an invalid
configuration.

## Debug information

`otelcol.connector.spanmetrics` does not expose any component-specific debug
information.

## Examples

### Explicit histogram and extra dimensions

In the example below, `http.status_code` and `http.method` are additional dimensions on top of:

- `service.name`
- `span.name`
- `span.kind`
- `status.code`

```river
otelcol.receiver.otlp "default" {
  http {}
  grpc {}

  output {
    traces  = [otelcol.connector.spanmetrics.default.input]
  }
}

otelcol.connector.spanmetrics "default" {
  // Since a default is not provided, the http.status_code dimension will be omitted
  // if the span does not contain http.status_code.
  dimension {
    name = "http.status_code"
  }

  // If the span is missing http.method, the connector will insert
  // the http.method dimension with value 'GET'.
  dimension {
    name = "http.method"
    default = "GET"
  }

  dimensions_cache_size = 333

  aggregation_temporality = "DELTA"

  histogram {
    unit = "s"
    explicit {
      buckets = ["333ms", "777s", "999h"]
    }
  }

  // The period on which all metrics (whose dimension keys remain in cache) will be emitted.
  metrics_flush_interval = "33s"

  namespace = "test.namespace"

  output {
    metrics = [otelcol.exporter.otlp.production.input]
  }
}

otelcol.exporter.otlp "production" {
  client {
    endpoint = env("OTLP_SERVER_ENDPOINT")
  }
}
```

### Sending metrics via a Prometheus remote write

In order for a `target_info` metric to be generated, the incoming spans resource scope
attributes must contain `service.name` and `service.instance.id` attributes.

The `target_info` metric will be generated for each resource scope, while OpenTelemetry
metric names and attributes will be normalized to be compliant with Prometheus naming rules.

```river
otelcol.receiver.otlp "default" {
  http {}
  grpc {}

  output {
    traces  = [otelcol.connector.spanmetrics.default.input]
  }
}

otelcol.connector.spanmetrics "default" {
  histogram {
    exponential {}
  }

  output {
    metrics = [otelcol.exporter.prometheus.default.input]
  }
}

otelcol.exporter.prometheus "default" {
  forward_to = [prometheus.remote_write.mimir.receiver]
}

prometheus.remote_write "mimir" {
  endpoint {
    url = "http://mimir:9009/api/v1/push"
  }
}
```
