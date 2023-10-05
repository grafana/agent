---
aliases:
- /docs/grafana-cloud/agent/flow/reference/cli/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/reference/cli/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/reference/cli/
canonical: https://grafana.com/docs/agent/latest/flow/reference/cli/
description: The Grafana Agent command line interface provides subcommands to perform
  various operations.
menuTitle: Command-line interface
title: The Grafana Agent command-line interface
description: Learn about the Grafana Agent command line interface
weight: 100
---

# The Grafana Agent command-line interface

When in Flow mode, the `grafana-agent` binary exposes a command-line interface with
subcommands to perform various operations.

The most common subcommand is [`run`][run] which accepts a config file and
starts Grafana Agent Flow.

Available commands:

* [`convert`][convert]: Convert a Grafana Agent configuration file.
* [`fmt`][fmt]: Format a Grafana Agent Flow configuration file.
* [`run`][run]: Start Grafana Agent Flow, given a configuration file.
* [`tools`][tools]: Read the WAL and provide statistical information.
* `completion`: Generate shell completion for the `grafana-agent-flow` CLI.
* `help`: Print help for supported commands.

[run]: {{< relref "./run.md" >}}
[fmt]: {{< relref "./fmt.md" >}}
[convert]: {{< relref "./convert.md" >}}
[tools]: {{< relref "./tools.md" >}}
