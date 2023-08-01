---
canonical: https://grafana.com/docs/agent/latest/flow/reference/components/loki.source.file/
title: loki.source.file
---

# loki.source.file

`loki.source.file` reads log entries from files and forwards them to other
`loki.*` components.

Multiple `loki.source.file` components can be specified by giving them
different labels.

## Usage

```river
loki.source.file "LABEL" {
  targets    = TARGET_LIST
  forward_to = RECEIVER_LIST
}
```

## Arguments
The component starts a new reader for each of the given `targets` and fans out
log entries to the list of receivers passed in `forward_to`.

`loki.source.file` supports the following arguments:

Name         | Type                   | Description          | Default | Required
------------ | ---------------------- | -------------------- | ------- | --------
`targets`    | `list(map(string))`    | List of files to read from. | | yes
`forward_to` | `list(LogsReceiver)`   | List of receivers to send log entries to. | | yes
`encoding`   | `string`               | The encoding to convert from when reading files. | `""` | no

## Blocks

The `loki.source.file` component doesn't support any inner blocks and is
configured fully through arguments.

The `encoding` argument must be a valid [IANA encoding][] name. If not set, it
defaults to UTF-8. 

## Exported fields

`loki.source.file` does not export any fields.

## Component health

`loki.source.file` is only reported as unhealthy if given an invalid
configuration.

## Debug information

`loki.source.file` exposes some target-level debug information per reader:
* The tailed path.
* Whether the reader is currently running.
* What is the last recorded read offset in the positions file.

## Debug metrics
* `loki_source_file_read_bytes_total` (gauge): Number of bytes read.
* `loki_source_file_file_bytes_total` (gauge): Number of bytes total.
* `loki_source_file_read_lines_total` (counter): Number of lines read.
* `loki_source_file_encoding_failures_total` (counter): Number of encoding failures.
* `loki_source_file_files_active_total` (gauge): Number of active files.

## Component behavior
Each element in the list of `targets` as a set of key-value pairs called
_labels_.
The set of targets can either be _static_, or dynamically provided periodically
by a service discovery component. The special label `__path__` _must always_ be
present and must point to the absolute path of the file to read from.
<!-- TODO(@tpaschalis) refer to local.file_match -->

The `__path__` value is  available as the `filename` label to each log entry
the component reads. All other labels starting with a double underscore are
considered _internal_ and are removed from the log entries before they're
passed to other `loki.*` components.

The component uses its data path (a directory named after the domain's
fully qualified name) to store its _positions file_. The positions file is used
to store read offsets, so that in case of a component or Agent restart,
`loki.source.file` can pick up tailing from the same spot.

If a file is removed from the `targets` list, its positions file entry is also
removed. When it's added back on, `loki.source.file` starts reading it from the
beginning.

## Example

This example collects log entries from the files specified in the targets
argument and forwards them to a `loki.write` component to be written to Loki.

```river
loki.source.file "tmpfiles" {
  targets    = [
    {__path__ = "/tmp/foo.txt", "color" = "pink"},
    {__path__ = "/tmp/bar.txt", "color" = "blue"},
    {__path__ = "/tmp/baz.txt", "color" = "grey"},
  ]
  forward_to = [loki.write.local.receiver]
}

loki.write "local" {
  endpoint {
    url = "loki:3100/api/v1/push"
  }
}
```

[IANA encoding]: https://www.iana.org/assignments/character-sets/character-sets.xhtml
