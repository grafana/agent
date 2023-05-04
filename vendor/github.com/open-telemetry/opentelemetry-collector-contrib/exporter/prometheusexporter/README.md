# Prometheus Exporter

| Status                   |                   |
| ------------------------ |-------------------|
| Stability                | [beta]            |
| Supported pipeline types | metrics           |
| Distributions            | [core], [contrib] |

Exports data in the [Prometheus format](https://prometheus.io/docs/concepts/data_model/), which allows it to be scraped by a [Prometheus](https://prometheus.io/) server.

## Getting Started

The following settings are required:

- `endpoint` (no default): the address on which metrics will be exposed by the Prometheus scrape handler. For full list of `HTTPServerSettings` refer [here](https://github.com/open-telemetry/opentelemetry-collector/tree/main/config/confighttp).

The following settings can be optionally configured:

- `const_labels` (no default): key/values that are applied for every exported metric.
- `namespace` (no default): if set, exports metrics under the provided value.
- `send_timestamps` (default = `false`): if true, sends the timestamp of the underlying metric sample in the response.
- `metric_expiration` (default = `5m`): defines how long metrics are exposed without updates
- `resource_to_telemetry_conversion`
  - `enabled` (default = false): If `enabled` is `true`, all the resource attributes will be converted to metric labels by default.
- `enable_open_metrics`: (default = `false`): If true, metrics will be exported using the OpenMetrics format. Exemplars are only exported in the OpenMetrics format.

Example:

```yaml
exporters:
  prometheus:
    endpoint: "1.2.3.4:1234"
    tls:
      ca_file: "/path/to/ca.pem"
      cert_file: "/path/to/cert.pem"
      key_file: "/path/to/key.pem"
    namespace: test-space
    const_labels:
      label1: value1
      "another label": spaced value
    send_timestamps: true
    metric_expiration: 180m
    enable_open_metrics: true
    resource_to_telemetry_conversion:
      enabled: true
```

## Metric names and labels normalization

OpenTelemetry metric names and attributes are normalized to be compliant with Prometheus naming rules. [Details on this normalization process are described in the Prometheus translator module](../../pkg/translator/prometheus/).

[beta]:https://github.com/open-telemetry/opentelemetry-collector#beta
[contrib]:https://github.com/open-telemetry/opentelemetry-collector-releases/tree/main/distributions/otelcol-contrib
[core]:https://github.com/open-telemetry/opentelemetry-collector-releases/tree/main/distributions/otelcol