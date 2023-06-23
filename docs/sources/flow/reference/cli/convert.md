---
title: grafana-agent convert
weight: 100
---

# `agent convert` command

The `agent convert` command converts a supported configuration file format
to a Grafana Agent Flow configuration file.

## Usage

Usage: `agent convert [FLAG ...] FILE_NAME`

If the `FILE_NAME` argument is not provided or if the `FILE_NAME` argument is
equal to `-`, `agent convert` converts the contents of standard input. Otherwise,
`agent convert` reads and converts the file from disk specified by the argument.

The `--write` flag can be specified to write the contents of the converted
file on disk with the converted results. `--write` can only be provided when
`agent convert` is not reading from standard input.

The command fails if the file being converted has syntactically incorrect
configuration.

The following flags are supported:

* `--write`, `-w`: Write the converted file back to disk when not reading from
  standard input.

* `--source-format`, `-f`: The format of the source file. Supported formats: 'prometheus'.

* `--bypass-warnings`, `-b`: Enable bypassing warnings when converting.