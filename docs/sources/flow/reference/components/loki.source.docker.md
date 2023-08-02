---
aliases:
- /docs/agent/latest/flow/reference/components/loki.source.docker
canonical: https://grafana.com/docs/agent/latest/flow/reference/components/loki.source.docker/
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
`labels`        | `map(string)`        | The default set of labels to apply on entries. | `"{}"` | no
`relabel_rules` | `RelabelRules`       | Relabeling rules to apply on log entries. | `"{}"` | no

## Blocks

The `loki.source.docker` component doesn't support any inner blocks and is
configured fully through arguments.

## Exported fields

`loki.source.docker` does not export any fields.

## Component health

`loki.source.docker` is only reported as unhealthy if given an invalid
configuration.

## Debug information

`loki.source.docker` exposes some debug information per target:
* Whether the target is ready to tail entries.
* The labels associated with the target.
* The most recent time a log line was read.

## Debug metrics

* `loki_source_docker_target_entries_total` (gauge): Total number of successful entries sent to the Docker target.
* `loki_source_docker_target_parsing_errors_total` (gauge): Total number of parsing errors while receiving Docker messages.

## Component behavior
The component uses its data path (a directory named after the domain's
fully qualified name) to store its _positions file_. The positions file is used
to store read offsets, so that in case of a component or Agent restart,
`loki.source.docker` can pick up tailing from the same spot.

## Example

This example collects log entries from the files specified in the `targets`
argument and forwards them to a `loki.write` component to be written to Loki.

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
