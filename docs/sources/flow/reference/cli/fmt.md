---
canonical: https://grafana.com/docs/agent/latest/flow/reference/cli/fmt/
title: grafana-agent fmt
weight: 100
---

# `grafana-agent fmt` command

The `grafana-agent fmt` command formats a given Grafana Agent Flow configuration file.

## Usage

Usage: `grafana-agent fmt [FLAG ...] FILE_NAME`

If the `FILE_NAME` argument is not provided or if the `FILE_NAME` argument is
equal to `-`, `grafana-agent fmt` formats the contents of standard input. Otherwise,
`grafana-agent fmt` reads and formats the file from disk specified by the argument.

The `--write` flag can be specified to replace the contents of the original
file on disk with the formatted results. `--write` can only be provided when
`grafana-agent fmt` is not reading from standard input.

The command fails if the file being formatted has syntactically incorrect River
configuration, but does not validate whether Flow components are configured
properly.

The following flags are supported:

* `--write`, `-w`: Write the formatted file back to disk when not reading from
  standard input.