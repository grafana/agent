---
title: Command-line interface
weight: 100
---

# Command-line interface

When in Flow mode, the `agent` binary exposes a command-line interface with
subcommands to perform various operations.

The most common subcommand is [`run`][run] which accepts a config file and
starts Grafana Agent Flow.

Available commands:

* [`agent run`][run]: Start Grafana Agent Flow, given a config file.
* [`agent fmt`][fmt]: Format a Grafana Agent Flow config file.
* `agent completion`: Generate shell completion for the `agent` CLI.
* `agent help`: Print help for supported commands.

[run]: {{< relref "./run.md" >}}
[fmt]: {{< relref "./fmt.md" >}}
