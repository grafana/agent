---
aliases:
- /docs/agent/latest/flow/reference/components/loki.source.file
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
The list of arguments that can be used to configure the block is presented
below.

Name         | Type                   | Description          | Default | Required
------------ | ---------------------- | -------------------- | ------- | --------
`targets`    | `list(map(string))`    | List of files to read from. | | yes
`forward_to` | `list(chan loki.Entry)` | List of receivers to send log entries to. | | yes

## Blocks

The `loki.source.file` component doesn't support any inner blocks and is
configured fully through arguments.

## Exported fields

`loki.source.file` does not export any fields.

## Component health

`loki.source.file` is only reported as unhealthy if given an invalid
configuration.

## Debug information

`loki.source.file` does not expose any component-specific debug information. ????

## Component behavior
Each element in the list of `targets` as a set of key-value pairs called
_labels_.
The set of targets can either be _static_, or dynamically provided periodically
by a service discovery component such as `discovery.fileglob`. The special
label `__path__` _must always_ be present and must point to the absolute path
of the file to read from.

The `__path__` value will be available as the `filename` label to each log
entry the component will read. All other labels starting with a double
underscore are considered _internal_ and will be removed from the log entries
after they've been read.

The component will use its data path (a directory named after the domain's
fully qualified name) to store its _positions file_. The positions file is used
to store read offsets, so that in case of a component or Agent restart,
`loki.source.file` can pick up tailing from the same spot. 

In case a file is removed from the `targets` list, its positions file entry
is also removed; that means that when it's added back on, `loki.source.file`
will start reading from the beginning.

## Example

This example collects log entries from the files specified in the targets
argument and forwards them to a `loki.write` component so they are can be 
written to Loki.

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
