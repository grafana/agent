---
aliases:
  - /docs/grafana-cloud/agent/flow/reference/components/discovery.process/
  - /docs/grafana-cloud/monitor-infrastructure/agent/flow/reference/components/discovery.process/
  - /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/reference/components/discovery.process/
  - /docs/grafana-cloud/send-data/agent/flow/reference/components/discovery.process/
canonical: https://grafana.com/docs/agent/latest/flow/reference/components/discovery.process/
description: Learn about discovery.process
title: discovery.process
---

# discovery.process

{{< docs/shared lookup="flow/stability/beta.md" source="agent" version="<AGENT_VERSION>" >}}

`discovery.process` discovers processes running on the local Linux OS.

{{< admonition type="note" >}}
To use the `discovery.process` component you must run {{< param "PRODUCT_NAME" >}} as root and inside host PID namespace.
{{< /admonition >}}

## Usage

```river
discovery.process "LABEL" {

}
```

## Arguments

The following arguments are supported:

| Name               | Type                | Description                                                                              | Default | Required |
|--------------------|---------------------|------------------------------------------------------------------------------------------|---------|----------|
| `join`             | `list(map(string))` | Join external targets to discovered processes targets based on `__container_id__` label. |         | no       |
| `refresh_interval` | `duration`          | How often to sync targets.                                                               | "60s"   | no       |

### Targets joining

If `join` is specified, `discovery.process` will join the discovered processes based on the `__container_id__` label.

For example, if `join` is specified as follows:

```river
discovery.process "all" {
  join = discovery.kubernetes.<label>.targets
}
```

And the discovered Kubernetes targets are:


```json
[
  {
    "pod": "pod-1",
    "__container_id__": "container-1"
  },
  {
    "pod": "pod-2",
    "__container_id__": "container-2"
  }
]
```

And the discovered processes are:

```json
[
  {
    "__process_pid__": "1",
    "__container_id__": "container-1"
  },
  {
    "__process_pid__": "2"
  }
]
```

The resulting targets are:

```json
[
  {
    "__container_id__": "container-1",
    "__process_pid__": "1",
    "pod": "pod-1"
  },
  {
    "__process_pid__": "2"
  },
  {
    "__container_id__": "container-1",
    "pod": "pod-1"
  },
  {
    "__container_id__": "container-2",
    "pod": "pod-2"
  }
]
```

## Blocks

The following blocks are supported inside the definition of `discovery.process`:

| Hierarchy       | Block               | Description                                    | Required |
|-----------------|---------------------|------------------------------------------------|----------|
| discover_config | [discover_config][] | Configures which process metadata to discover. | no       |

[discover_config]: #discover_config-block

### discover_config block

The `discover_config` block describes which process metadata to discover.

The following arguments are supported:

| Name                 | Type   | Description                                                                   | Default | Required |
|----------------------|--------|-------------------------------------------------------------------------------|---------|----------|
| `exe`                | `bool` | A flag to enable discovering `__meta_process_exe` label.                      | true    | no       |
| `cwd`                | `bool` | A flag to enable discovering `__meta_process_cwd` label.                      | true    | no       |
| `commandline`        | `bool` | A flag to enable discovering `__meta_process_commandline` label.              | true    | no       |
| `uid`                | `bool` | A flag to enable discovering `__meta_process_uid`: label.                     | true    | no       |
| `username`           | `bool` | A flag to enable discovering `__meta_process_username`: label.                | true    | no       |
| `container_id`       | `bool` | A flag to enable discovering `__container_id__` label.                        | true    | no       |
| `analyze_executable` | `bool` | A flag to enable [additional process analysis](#additional-process-analysis). | false   | no       |

## Exported fields

The following fields are exported and can be referenced by other components:

| Name      | Type                | Description                                            |
|-----------|---------------------|--------------------------------------------------------|
| `targets` | `list(map(string))` | The set of processes discovered on the local Linux OS. |

Each target includes the following labels:

* `__process_pid__`: The process PID.
* `__meta_process_exe`: The process executable path. Taken from `/proc/<pid>/exe`.
* `__meta_process_cwd`: The process current working directory. Taken from `/proc/<pid>/cwd`.
* `__meta_process_commandline`: The process command line. Taken from `/proc/<pid>/cmdline`.
* `__meta_process_uid`: The process UID. Taken from `/proc/<pid>/status`.
* `__meta_process_username`: The process username. Taken from `__meta_process_uid` and `os/user/LookupID`.
* `__container_id__`: The container ID. Taken from `/proc/<pid>/cgroup`. If the process is not running in a container,
  this label is not set.

### Additional process analysis

If `analyze_executable` is set to true within #discover_config-block, additional labels will be added for each target.
The exact labels depend on the language and toolchain used to build the executable, as well as the context in which it executes.
The agent currently supports language detection for Go, .NET, Java, Python and C++.

#### Example: Additional labels for a Go process

```json
{
  "__meta_process_go__": "true",
  "__meta_process_go_build_setting__buildmode": "exe",
  "__meta_process_go_build_setting__compiler": "gc",
  "__meta_process_go_build_setting_cgo_cflags": "",
  "__meta_process_go_build_setting_cgo_cppflags": "",
  "__meta_process_go_build_setting_cgo_cxxflags": "",
  "__meta_process_go_build_setting_cgo_enabled": "1",
  "__meta_process_go_build_setting_cgo_ldflags": "",
  "__meta_process_go_build_setting_goarch": "arm64",
  "__meta_process_go_build_setting_goos": "linux",
  "__meta_process_go_build_setting_vcs": "git",
  "__meta_process_go_build_setting_vcs_modified": "true",
  "__meta_process_go_build_setting_vcs_revision": "2c80d78e7c602e6beb34aa5b6144fd771fc0f476",
  "__meta_process_go_build_setting_vcs_time": "2024-02-09T20:18:03Z",
  "__meta_process_go_godeltaprof__": "true",
  "__meta_process_go_godeltaprof_version__": "v0.1.7",
  "__meta_process_go_module_path__": "github.com/grafana/agent",
  "__meta_process_go_module_version__": "(devel)",
  "__meta_process_go_version__": "go1.21.6"
}
```

#### Example: Additional labels for a Java process

```json
{
  "__meta_process_java__": "true",
  "__meta_process_java_classpath__": "build/libs/simple.jar",
  "__meta_process_java_home__": "/usr/lib/jvm/java-17-openjdk-arm64",
  "__meta_process_java_os_arch__": "aarch64",
  "__meta_process_java_os_name__": "Linux",
  "__meta_process_java_version__": "17.0.9",
  "__meta_process_java_version_date__": "2023-10-17",
  "__meta_process_java_vm_flags__": "-XX:CICompilerCount=4 -XX:+UseG1GC ...",
  "__meta_process_java_vm_type__": "jdk",
}
```

## Component health

`discovery.process` is only reported as unhealthy when given an invalid
configuration. In those cases, exported fields retain their last healthy
values.

## Debug information

`discovery.process` does not expose any component-specific debug information.

## Debug metrics

`discovery.process` does not expose any component-specific debug metrics.

## Examples

### Example discovering processes on the local host

```river
discovery.process "all" {
  refresh_interval = "60s"
  discover_config {
    cwd = true
    exe = true
    commandline = true
    username = true
    uid = true
    container_id = true
  }
}

```

### Example discovering processes on the local host and joining with `discovery.kubernetes`

```river
discovery.kubernetes "pyroscope_kubernetes" {
  selectors {
    field = "spec.nodeName=" + env("HOSTNAME")
    role = "pod"
  }
  role = "pod"
}

discovery.process "all" {
  join = discovery.kubernetes.pyroscope_kubernetes.targets
  refresh_interval = "60s"
  discover_config {
    cwd = true
    exe = true
    commandline = true
    username = true
    uid = true
    container_id = true
  }
}

```
<!-- START GENERATED COMPATIBLE COMPONENTS -->

## Compatible components

`discovery.process` can accept arguments from the following components:

- Components that export [Targets]({{< relref "../compatibility/#targets-exporters" >}})

`discovery.process` has exports that can be consumed by the following components:

- Components that consume [Targets]({{< relref "../compatibility/#targets-consumers" >}})

{{< admonition type="note" >}}
Connecting some components may not be sensible or components may require further configuration to make the connection work correctly.
Refer to the linked documentation for more details.
{{< /admonition >}}

<!-- END GENERATED COMPATIBLE COMPONENTS -->