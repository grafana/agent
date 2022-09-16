---
aliases:
- /docs/agent/latest/flow/reference/cli/run
title: agent run
weight: 100
---

# `agent fmt` command

The `agent fmt` command formats a given Grafana Agent Flow configuration file.

## Usage

Usage: `agent fmt [flags] file`

If the `file` argument is not provided or if the `file` argument is equal to
`-`, `agent fmt` formats the contents of standard input. Otherwise, `agent fmt`
reads and formats the file from disk specified by the argument.

The `--write` flag can be specified to write the formatted result back to disk.
`--write` can only be provided when `agent fmt` is not reading from standard
input.

The following flags are supported:

* `--write`, `-w`: Write the formatted file back to disk when not reading from
  standard input.
