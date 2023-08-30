---
aliases:
- ../../otelcol/output-block-logs/
canonical: https://grafana.com/docs/agent/latest/shared/flow/reference/components/output-block-logs/
headless: true
---

The `output` block configures a set of components to forward resulting
telemetry data to.

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`logs` | `list(otelcol.Consumer)` | List of consumers to send logs to. | `[]` | no

The `output` block must be specified, but all of its arguments are optional. By
default, telemetry data is dropped. To send telemetry data to other components,
configure the `logs` argument accordingly.
