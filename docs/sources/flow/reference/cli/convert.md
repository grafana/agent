---
title: grafana-agent convert
weight: 100
---

# `grafana-agent convert` command

The `grafana-agent convert` command converts a supported configuration file format
to a Grafana Agent Flow river file.

## Usage

Usage: `grafana-agent convert [FLAG ...] FILE_NAME`

If the `FILE_NAME` argument is not provided or if the `FILE_NAME` argument is
equal to `-`, `grafana-agent convert` converts the contents of standard input. Otherwise,
`grafana-agent convert` reads and converts the file from disk specified by the argument.

The `--output` flag can be specified to write the contents of the converted
file on disk to the specified path. `--output` can only be provided when
`grafana-agent convert` is not reading from standard input.

The command fails if the file being converted has syntactically incorrect
configuration or cannot be converted to a Grafana Agent Flow river file.

The following flags are supported:

* `--output`, `-o`: The filepath where the output is written.

* `--source-format`, `-f`: The format of the source file. Supported formats: 'prometheus'.

* `--bypass-warnings`, `-b`: Enable bypassing warnings when converting.