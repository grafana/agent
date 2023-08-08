---
aliases:
- /docs/grafana-cloud/agent/flow/reference/components/loki.echo/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/reference/components/loki.echo/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/reference/components/loki.echo/
- /docs/grafana-cloud/send-data/agent/flow/reference/components/loki.echo/
canonical: https://grafana.com/docs/agent/latest/flow/reference/components/loki.echo/
description: Learn about loki.echo
labels:
  stage: beta
title: loki.echo
---

# loki.echo

{{< docs/shared lookup="flow/stability/beta.md" source="agent" version="<AGENT VERSION>" >}}

`loki.echo` receives log entries from other `loki` components and prints them
to the process' standard output (stdout).

Multiple `loki.echo` components can be specified by giving them
different labels.

## Usage

```river
loki.echo "LABEL" {}
```

## Arguments

`loki.echo` accepts no arguments.

## Exported fields

The following fields are exported and can be referenced by other components:

Name | Type | Description
---- | ---- | -----------
`receiver` | `LogsReceiver` | A value that other components can use to send log entries to.

## Component health

`loki.echo` is only reported as unhealthy if given an invalid configuration.

## Debug information

`loki.echo` does not expose any component-specific debug information.

<!-- START GENERATED COMPATIBLE COMPONENTS -->

## Compatible components

`loki.echo` can accept data from the following components:

- Components that output Loki Logs:
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

## Example

This example creates a pipeline that reads log files from `/var/log` and
prints log lines to echo:

```river
local.file_match "varlog" {
  path_targets = [{
    __path__ = "/var/log/*log",
    job      = "varlog",
  }]
}

loki.source.file "logs" {
  targets    = local.file_match.varlog.targets
  forward_to = [loki.echo.example.receiver]
}

loki.echo "example" { }
```
