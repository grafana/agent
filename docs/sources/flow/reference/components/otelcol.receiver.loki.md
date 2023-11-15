---
aliases:
- /docs/grafana-cloud/agent/flow/reference/components/otelcol.receiver.loki/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/reference/components/otelcol.receiver.loki/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/reference/components/otelcol.receiver.loki/
- /docs/grafana-cloud/send-data/agent/flow/reference/components/otelcol.receiver.loki/
canonical: https://grafana.com/docs/agent/latest/flow/reference/components/otelcol.receiver.loki/
description: Learn about otelcol.receiver.loki
labels:
  stage: beta
title: otelcol.receiver.loki
---

# otelcol.receiver.loki

{{< docs/shared lookup="flow/stability/beta.md" source="agent" version="<AGENT VERSION>" >}}

`otelcol.receiver.loki` receives Loki log entries, converts them to the
OpenTelemetry logs format, and forwards them to other `otelcol.*` components.

Multiple `otelcol.receiver.loki` components can be specified by giving them
different labels.

## Usage

```river
otelcol.receiver.loki "LABEL" {
  output {
    logs = [...]
  }
}
```

## Arguments

`otelcol.receiver.loki` doesn't support any arguments and is configured fully
through inner blocks.

## Blocks

The following blocks are supported inside the definition of
`otelcol.receiver.loki`:

Hierarchy | Block | Description | Required
--------- | ----- | ----------- | --------
output | [output][] | Configures where to send converted telemetry data. | yes

[output]: #output-block

### output block

{{< docs/shared lookup="flow/reference/components/output-block.md" source="agent" version="<AGENT VERSION>" >}}

## Exported fields

The following fields are exported and can be referenced by other components:

Name | Type | Description
---- | ---- | -----------
`receiver` | `LogsReceiver` | A value that other components can use to send Loki logs to.

## Component health

`otelcol.receiver.loki` is only reported as unhealthy if given an invalid
configuration.

## Debug information

`otelcol.receiver.loki` does not expose any component-specific debug
information.

## Example

This example uses the `otelcol.receiver.loki` component as a bridge
between the Loki and OpenTelemetry ecosystems. The component exposes a
receiver which the `loki.source.file` component uses to send Loki log entries
to. The logs are converted to the OTLP format before they are forwarded
to the `otelcol.exporter.otlp` component to be sent to an OTLP-capable
endpoint:

```river
loki.source.file "default" {
  targets = [
    {__path__ = "/tmp/foo.txt", "loki.format" = "logfmt"},
    {__path__ = "/tmp/bar.txt", "loki.format" = "json"},
  ]
  forward_to = [otelcol.receiver.loki.default.receiver]
}

otelcol.receiver.loki "default" {
  output {
    logs = [otelcol.exporter.otlp.default.input]
  }
}

otelcol.exporter.otlp "default" {
  client {
    endpoint = env("OTLP_ENDPOINT")
  }
}
```

<!-- START GENERATED COMPATIBLE COMPONENTS -->

## Compatible components

`otelcol.receiver.loki` can accept data from the following components:

- Components that output Loki Logs:
  - [`faro.receiver`]({{< relref "../components/faro.receiver.md" >}})
  - [`loki.process`]({{< relref "../components/loki.process.md" >}})
  - [`loki.relabel`]({{< relref "../components/loki.relabel.md" >}})
  - [`loki.source.api`]({{< relref "../components/loki.source.api.md" >}})
  - [`loki.source.awsfirehose`]({{< relref "../components/loki.source.awsfirehose.md" >}})
  - [`loki.source.azure_event_hubs`]({{< relref "../components/loki.source.azure_event_hubs.md" >}})
  - [`loki.source.cloudflare`]({{< relref "../components/loki.source.cloudflare.md" >}})
  - [`loki.source.docker`]({{< relref "../components/loki.source.docker.md" >}})
  - [`loki.source.file`]({{< relref "../components/loki.source.file.md" >}})
  - [`loki.source.gcplog`]({{< relref "../components/loki.source.gcplog.md" >}})
  - [`loki.source.gelf`]({{< relref "../components/loki.source.gelf.md" >}})
  - [`loki.source.heroku`]({{< relref "../components/loki.source.heroku.md" >}})
  - [`loki.source.journal`]({{< relref "../components/loki.source.journal.md" >}})
  - [`loki.source.kafka`]({{< relref "../components/loki.source.kafka.md" >}})
  - [`loki.source.kubernetes`]({{< relref "../components/loki.source.kubernetes.md" >}})
  - [`loki.source.kubernetes_events`]({{< relref "../components/loki.source.kubernetes_events.md" >}})
  - [`loki.source.podlogs`]({{< relref "../components/loki.source.podlogs.md" >}})
  - [`loki.source.syslog`]({{< relref "../components/loki.source.syslog.md" >}})
  - [`loki.source.windowsevent`]({{< relref "../components/loki.source.windowsevent.md" >}})
  - [`otelcol.exporter.loki`]({{< relref "../components/otelcol.exporter.loki.md" >}})


Note that connecting some components may not be feasible or components may require further configuration to make the connection work correctly. Please refer to the linked documentation for more details.

<!-- END GENERATED COMPATIBLE COMPONENTS -->
