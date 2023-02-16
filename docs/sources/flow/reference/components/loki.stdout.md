---
title: loki.stdout
---

# loki.stdout

`loki.stdout` receives log entries from other `loki` components and prints them
to the process' standard output (stdout).

Multiple `loki.stdout` components can be specified by giving them
different labels.

## Usage

```river
loki.stdout "LABEL" {}
```

## Arguments

`loki.stdout` accepts no arguments.

## Exported fields

The following fields are exported and can be referenced by other components:

Name | Type | Description
---- | ---- | -----------
`receiver` | `LogsReceiver` | A value that other components can use to send log entries to.

## Component health

`loki.stdout` is only reported as unhealthy if given an invalid configuration.

## Debug information

`loki.stdout` does not expose any component-specific debug information.

## Example

This example creates a pipeline which reads log files from `/var/log` and
prints log lines to stdout:

```river
discovery.file "varlog" {
  path_targets = [{
    __path__ = "/var/log/*log",
    job      = "varlog",
  }]
}

loki.source.file "logs" {
  targets    = discovery.file.varlog.targets
  forward_to = [loki.stdout.example.receiver]
}

loki.stdout "example" { }
```
