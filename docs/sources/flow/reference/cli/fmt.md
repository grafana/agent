---
canonical: https://grafana.com/docs/agent/latest/flow/reference/cli/fmt/
description: The `fmt` command formats a Grafana Agent configuration file.
title: fmt command
menuTitle: fmt
weight: 200
---

# `fmt` command

The `fmt` command formats a given Grafana Agent Flow configuration file.

## Usage

Usage:

* `AGENT_MODE=flow grafana-agent fmt [FLAG ...] FILE_NAME`
* `grafana-agent-flow fmt [FLAG ...] FILE_NAME`

   Replace the following:

   * `FLAG`: One or more flags that define the input and output of the command.
   * `FILE_NAME`: The Grafana Agent configuration file.

If the `FILE_NAME` argument is not provided or if the `FILE_NAME` argument is
equal to `-`, `fmt` formats the contents of standard input. Otherwise,
`fmt` reads and formats the file from disk specified by the argument.

The `--write` flag can be specified to replace the contents of the original
file on disk with the formatted results. `--write` can only be provided when
`fmt` is not reading from standard input.

The command fails if the file being formatted has syntactically incorrect River
configuration, but does not validate whether Flow components are configured
properly.

The following flags are supported:

* `--write`, `-w`: Write the formatted file back to disk when not reading from
  standard input.