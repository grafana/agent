---
canonical: https://grafana.com/docs/agent/latest/flow/reference/cli/
title: Command-line interface
weight: 100
---

# Command-line interface

When in Flow mode, the `grafana-agent` binary exposes a command-line interface with
subcommands to perform various operations.

The most common subcommand is [`run`][run] which accepts a config file and
starts Grafana Agent Flow.

Available commands:

* [`grafana-agent-flow run`][run]: Start Grafana Agent Flow, given a configuration file.
* [`grafana-agent-flow fmt`][fmt]: Format a Grafana Agent Flow configuration file.
* [`grafana-agent-flow convert`][convert]: Convert a Grafana Agent configuration file.
* `grafana-agent-flow completion`: Generate shell completion for the `grafana-agent-flow` CLI.
* `grafana-agent-flow help`: Print help for supported commands.

[run]: {{< relref "./run.md" >}}
[fmt]: {{< relref "./fmt.md" >}}
[convert]: {{< relref "./convert.md" >}}
