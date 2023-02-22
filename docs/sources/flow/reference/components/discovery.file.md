---
title: discovery.file
---

# discovery.file

`discovery.file` discovers files on the local filesystem using glob patterns and the [doublestar][] library.

[doublestar]: https://github.com/bmatcuk/doublestar

## Usage

```river
discovery.file "LABEL" {
  path_targets = [{"__path__" = "DOUBLESTAR_PATH"}]
}
```

## Arguments

The following arguments are supported:

Name            | Type                | Description                                                                                | Default | Required
--------------- | ------------------- | ------------------------------------------------------------------------------------------ |---------| --------
`path_targets`  | `list(map(string))` | Targets to expand; looks for glob patterns on the  `__path__` and `__path_exclude__` keys. |         | yes
`sync_period`   | `duration`          | How often to sync filesystem and targets.                                                  | `"10s"` | no

`path_targets` uses [doublestar][] style paths.
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
discovery.file "tmp" {
    path_targets = [{"__path__" = "/tmp/logs/**/*.log"}]
}

loki.source.file "files" {
    targets    = discovery.file.tmp.targets
    forward_to = [ /* ... */ ]
}
```

### Kubernetes

This example finds all the logs on pods and monitors them.

```river
discovery.kubernetes "k8s" {
  role = "pod"
}

discovery.relabel "k8s" {
  targets = discovery.kubernetes.k8s.targets
 
  rule {
    source_labels = ["__meta_kubernetes_namespace", "__meta_kubernetes_pod_label_name"]
    target_label  = "job"
    separator     = "/"
  }

  rule {
    source_labels = ["__meta_kubernetes_pod_uid", "__meta_kubernetes_pod_container_name"]
    target_label  = "__path__"
    separator     = "/" 
    replacement   = "/var/log/pods/*$1/*.log"
  } 
}

discovery.file "pods" {
    path_targets = discovery.relabel.k8s.output
}

loki.source.file "pods" {
    targets = discovery.file.pods.targets
    forward_to = [loki.write.endpoint.receiver]
}

loki.write "endpoint" {
    endpoint {
        url = "LOKI_PATH"
        basic_auth {
            username = USERNAME
            password = "PASSWORD"
        }
    }
}
```
