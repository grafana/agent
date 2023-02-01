---
aliases:
- /docs/agent/latest/flow/reference/components/loki.source.docker
title: loki.source.docker
---

# loki.source.docker

`loki.source.docker` reads log entries from Docker containers and forwards them
to other `loki.*` components. Each component can read from a single Docker
daemon.

Multiple `loki.source.docker` components can be specified by giving them
different labels.

## Usage

```river
loki.source.docker "LABEL" {
  host       = HOST
  targets    = TARGET_LIST
  forward_to = RECEIVER_LIST
}
```

## Arguments
The component starts a new reader for each of the given `targets` and fans out
log entries to the list of receivers passed in `forward_to`.

`loki.source.file` supports the following arguments:

Name            | Type                 | Description          | Default | Required
--------------- | -------------------- | -------------------- | ------- | --------
`host`          | `string`             | Address of the Docker daemon. | | yes
`targets`       | `list(map(string))`  | List of containers to read logs from. | | yes
`forward_to`    | `list(LogsReceiver)` | List of receivers to send log entries to. | | yes
`relabel_rules` | `RelabelRules`       | Relabeling rules to apply on log entries. | "{}" | no

## Blocks

The `loki.source.docker` component doesn't support any inner blocks and is
configured fully through arguments.

## Exported fields

`loki.source.docker` does not export any fields.

## Component health

`loki.source.docker` is only reported as unhealthy if given an invalid
configuration.

## Debug information

`loki.source.docker` exposes some target-level debug information per reader:
 ????
* A
* B
* C

## Debug metrics
* `loki_source_docker_target_entries_total` (gauge): Total number of successful entries sent to the Docker target.
* `loki_source_docker_target_parsing_errors_total` (gauge): Total number of parsing errors while receiving Docker messages.

## Component behavior
Each element in the list of `targets` as a set of key-value pairs called
_labels_.
The set of targets can either be _static_, or dynamically provided periodically
by a service discovery component. The special label `__path__` _must always_ be
present and must point to the absolute path of the file to read from.
<!-- TODO(@tpaschalis) refer to discovery.fileglob -->

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
argument and forwards them to a `loki.write` component so they are can be 
written to Loki.

```river
discovery.docker "linux" {
  host = "unix:///var/run/docker.sock"
}

loki.source.docker "default" {
  host       = "unix:///var/run/docker.sock"
  targets    = discovery.docker.linux.targets 
  forward_to = [loki.write.local.receiver]
}

loki.write "local" {
  endpoint {
    url = "loki:3100/api/v1/push"
  }
}
```
