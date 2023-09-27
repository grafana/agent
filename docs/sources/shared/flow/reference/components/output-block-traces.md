---
aliases:
- ../../otelcol/output-block-traces/
- /docs/grafana-cloud/agent/shared/flow/reference/components/output-block-traces/
- /docs/grafana-cloud/monitor-infrastructure/agent/shared/flow/reference/components/output-block-traces/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/shared/flow/reference/components/output-block-traces/
canonical: https://grafana.com/docs/agent/latest/shared/flow/reference/components/output-block-traces/
description: Shared content, output block traces
headless: true
---

The `output` block configures a set of components to forward resulting
telemetry data to.

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`traces` | `list(otelcol.Consumer)` | List of consumers to send traces to. | `[]` | no

The `output` block must be specified, but all of its arguments are optional. By
default, telemetry data is dropped. To send telemetry data to other components,
configure the `traces` argument accordingly.
