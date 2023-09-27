---
aliases:
- ../../otelcol/output-block-metrics/
- /docs/grafana-cloud/agent/shared/flow/reference/components/output-block-metrics/
- /docs/grafana-cloud/monitor-infrastructure/agent/shared/flow/reference/components/output-block-metrics/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/shared/flow/reference/components/output-block-metrics/
description: Shared content, output block metrics
headless: true
---

The `output` block configures a set of components to forward resulting
telemetry data to.

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`metrics` | `list(otelcol.Consumer)` | List of consumers to send metrics to. | `[]` | no

The `output` block must be specified, but all of its arguments are optional. By
default, telemetry data is dropped. To send telemetry data to other components,
configure the `metrics` argument accordingly.
