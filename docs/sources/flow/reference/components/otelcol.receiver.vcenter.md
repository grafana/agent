---
aliases:
- /docs/grafana-cloud/agent/flow/reference/components/otelcol.receiver.vcenter/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/reference/components/otelcol.receiver.vcenter/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/reference/components/otelcol.receiver.vcenter/
canonical: https://grafana.com/docs/agent/latest/flow/reference/components/otelcol.receiver.vcenter/
title: otelcol.receiver.vcenter
description: Learn about otelcol.receiver.vcenter
labels:
  stage: experimental
---

# otelcol.receiver.vcenter

{{< docs/shared lookup="flow/stability/experimental.md" source="agent" version="<AGENT VERSION>" >}}

`otelcol.receiver.vcenter` accepts metrics from a 
vCenter or ESXi host running VMware vSphere APIs and
forwards it to other `otelcol.*` components.

> **NOTE**: `otelcol.receiver.vcenter` is a wrapper over the upstream
> OpenTelemetry Collector `vcenter` receiver from the `otelcol-contrib`
> distribution. Bug reports or feature requests will be redirected to the
> upstream repository, if necessary.

Multiple `otelcol.receiver.vcenter` components can be specified by giving them
different labels.

The full list of metrics that can be collected can be found in [vcenter receiver documentation][vcenter metrics].

[vcenter metrics]: https://github.com/open-telemetry/opentelemetry-collector/blob/{{< param "OTEL_VERSION" >}}/receiver/vcenterreceiver/documentation.md

## Prerequisites

This receiver has been built to support ESXi and vCenter versions:

- 7.5
- 7.0
- 6.7

A “Read Only” user assigned to a vSphere with permissions to the vCenter server, cluster and all subsequent resources being monitored must be specified in order for the receiver to retrieve information about them.

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
`collection_interval` | `duration` | Defines how often to collect metrics. | `"1m"` | no
`initial_delay` | `duration` | Defines how long this receiver waits before starting. | `"1s"` | no
`timeout` | `duration` | Defines the timeout for the underlying HTTP client. | `"0s"` | no

`endpoint` has the format `<protocol>://<hostname>`. For example, `https://vcsa.hostname.localnet`.

## Blocks

The following blocks are supported inside the definition of
`otelcol.receiver.vcenter`:

Hierarchy | Block | Description | Required
--------- | ----- | ----------- | --------
tls | [tls][] | Configures TLS for the HTTP client. | no
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
  endpoint = "http://localhost:15672"
  username = "otelu"
  password = "password"

  output {
    metrics = [otelcol.processor.batch.default.input]
  }
}

otelcol.processor.batch "default" {
  output {
    metrics = [otelcol.exporter.otlp.default.input]
  }
}

otelcol.exporter.otlp "default" {
  client {
    endpoint = env("OTLP_ENDPOINT")
  }
}
