---
aliases:
- /docs/grafana-cloud/agent/flow/reference/components/otelcol.receiver.vcenter/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/reference/components/otelcol.receiver.vcenter/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/reference/components/otelcol.receiver.vcenter/
canonical: https://grafana.com/docs/agent/latest/flow/reference/components/otelcol.receiver.vcenter/
title: otelcol.receiver.vcenter
description: Learn about otelcol.receiver.vcenter
---

# otelcol.receiver.vcenter

`otelcol.receiver.vcenter` accepts telemetry data from a 
vCenter or ESXi host running VMware vSphere APIs and
forwards it to other `otelcol.*` components.

> **NOTE**: `otelcol.receiver.vcenter` is a wrapper over the upstream
> OpenTelemetry Collector `vcenter` receiver from the `otelcol-contrib`
> distribution. Bug reports or feature requests will be redirected to the
> upstream repository, if necessary.

Multiple `otelcol.receiver.vcenter` components can be specified by giving them
different labels.

## Usage

```river
otelcol.receiver.vcenter "LABEL" {
  endpoint = "VCENTER_ENDPOINT"
  username = "VCENTER_USERNAME"
  password = "VCENTER_PASSWORD"

  output {
    metrics = [...]
  }
}
```

## Arguments

`otelcol.receiver.vcenter` supports the following arguments:


Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`endpoint` | `string` | Endpoint to a vCenter Server or ESXi host which has the SDK path enabled. | | yes
`username` | `string` | Username to use for authentication. | | yes
`password` | `string` | Password to use for authentication. | | yes
`collection_interval` | `duration` | This receiver collects metrics on an interval. | `"1m"` | no
`initial_delay` | `duration` | Defines how long this receiver waits before starting.. | `"1s"` | no
`timeout` | `duration` | Defines the timeout for the underlying HTTP client. | `"0s"` | no

`endpoint` has the following format: `<protocol>://<hostname>` (e.g. `https://vcsa.hostname.localnet`)

## Blocks

The following blocks are supported inside the definition of
`otelcol.receiver.vcenter`:

Hierarchy | Block | Description | Required
--------- | ----- | ----------- | --------
tls | [tls][] | Configures TLS for the server. | no
debug_metrics | [debug_metrics][] | Configures the metrics that this component generates to monitor its state. | no
output | [output][] | Configures where to send received telemetry data. | yes

[tls]: #tls-block
[debug_metrics]: #debug_metrics-block
[output]: #output-block

### tls block

The `tls` block configures TLS settings used for a server. If the `tls` block
isn't provided, TLS won't be used for connections to the server.

{{< docs/shared lookup="flow/reference/components/otelcol-tls-config-block.md" source="agent" version="<AGENT VERSION>" >}}

### debug_metrics block

{{< docs/shared lookup="flow/reference/components/otelcol-debug-metrics-block.md" source="agent" version="<AGENT VERSION>" >}}

### output block

{{< docs/shared lookup="flow/reference/components/output-block.md" source="agent" version="<AGENT VERSION>" >}}

## Exported fields

`otelcol.receiver.vcenter` does not export any fields.

## Component health

`otelcol.receiver.vcenter` is only reported as unhealthy if given an invalid
configuration.

## Debug information

`otelcol.receiver.vcenter` does not expose any component-specific debug
information.

## Example

This example forwards received telemetry data through a batch processor before
finally sending it to an OTLP-capable endpoint:

```river
otelcol.receiver.vcenter "default" {
  endpoint         = "http://localhost:15672"
  username		   = "otelu"
  password         = "password"

  output {
    metrics = [otelcol.processor.batch.default.input]
    logs    = [otelcol.processor.batch.default.input]
    traces  = [otelcol.processor.batch.default.input]
  }
}

otelcol.processor.batch "default" {
  output {
    metrics = [otelcol.exporter.otlp.default.input]
    logs    = [otelcol.exporter.otlp.default.input]
    traces  = [otelcol.exporter.otlp.default.input]
  }
}

otelcol.exporter.otlp "default" {
  client {
    endpoint = env("OTLP_ENDPOINT")
  }
}
