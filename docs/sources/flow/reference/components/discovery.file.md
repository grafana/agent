---
aliases:
- /docs/agent/latest/flow/reference/components/discovery.file
title: discovery.file
---

# discovery.file

`discovery.file` discovers files on the local filesystem using the [doublestar][] library.

[doublestar]: https://github.com/bmatcuk/doublestar

## Usage

```river
discovery.file "LABEL" {
  paths = ["DOUBLESTAR_PATH"]
}
```

## Arguments

The following arguments are supported:

Name | Type           | Description                                                                                                      | Default | Required
---- |----------------|------------------------------------------------------------------------------------------------------------------|-----| --------
`paths` | `list(map(string))` | Doublestar-compatible paths to search, looks for keys with `__path__`.                                           |     | yes
`exclude_paths` | `list(map(string))` | Doublestar-compatible paths to exclude, exclude_paths supercedes paths, looks for keys with `__path_exclude__` . |     | no
`update_period` | `duration`     | How often to sync filesystem and targets.                                                                        | `"10s"` | no

`paths` and `exclude_paths` use [doublestar][] style paths.
* `/tmp/**/*.log` will match all subfolders of `tmp` and include any files that end in `*.log`.
* `/tmp/apache/*.log` will match only files in `/tmp/apache/` that end in `*.log`.
* `/tmp/**` will match all subfolders of `tmp`, `tmp` itself, and all files.


## Exported fields

The following fields are exported and can be referenced by other components:

Name | Type | Description
---- | ---- | -----------
`targets` | `list(map(string))` | The set of targets discovered from the filesystem.

Each target includes the following labels:

* `__path__`: Absolute path to the file.

## Component health

`discovery.file` is only reported as unhealthy when given an invalid
configuration. In those cases, exported fields retain their last healthy
values.

## Debug information

`discovery.file` does not expose any component-specific debug information.

### Debug metrics

`discovery.file` does not expose any component-specific debug metrics.

## Examples

This example discovers all files and folders under `/tmp/logs`. The absolute paths are 
used by `loki.source.file.files` targets.

```river
discovery.file "files" {
    paths = [{"__path__" = "/tmp/logs/**/*.log"}]
}
loki.source.file "files" {
    targets = discovery.file.files.targets
    forward_to = []
}
```