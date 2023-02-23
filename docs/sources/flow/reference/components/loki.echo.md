---
title: loki.echo
labels:
  stage: beta
---

# loki.echo

{{< docs/shared lookup="flow/stability/beta.md" source="agent" >}}

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

## Example

This example creates a pipeline that reads log files from `/var/log` and
prints log lines to echo:

```river
discovery.file "varlog" {
  path_targets = [{
    __path__ = "/var/log/*log",
    job      = "varlog",
  }]
}

loki.source.file "logs" {
  targets    = discovery.file.varlog.targets
  forward_to = [loki.echo.example.receiver]
}

loki.echo "example" { }
```
