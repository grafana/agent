---
aliases:
  - /docs/grafana-cloud/agent/flow/reference/components/pyroscope.java/
  - /docs/grafana-cloud/monitor-infrastructure/agent/flow/reference/components/pyroscope.java/
  - /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/reference/components/pyroscope.java/
  - /docs/grafana-cloud/send-data/agent/flow/reference/components/pyroscope.java/
canonical: https://grafana.com/docs/agent/latest/flow/reference/components/pyroscope.java/
description: Learn about pyroscope.java
title: pyroscope.java
---

# pyroscope.java

{{< docs/shared lookup="flow/stability/beta.md" source="agent" version="<AGENT_VERSION>" >}}

`pyroscope.java` continuously profiles Java processes running on the local Linux OS
using [async-profiler](https://github.com/async-profiler/async-profiler).

{{% admonition type="note" %}}
To use the  `pyroscope.java` component you must run {{< param "PRODUCT_NAME" >}} as root and inside host PID namespace.
{{% /admonition %}}

## Usage

```river
pyroscope.java "LABEL" {
  targets    = TARGET_LIST
  forward_to = RECEIVER_LIST
}
```

## Arguments

The following arguments are supported:

| Name         | Type                     | Description                                      | Default | Required |
|--------------|--------------------------|--------------------------------------------------|---------|----------|
| `targets`    | `list(map(string))`      | List of java process targets to profile.         |         | yes      |
| `forward_to` | `list(ProfilesReceiver)` | List of receivers to send collected profiles to. |         | yes      |
| `tmp_dir`    | `string`                 | Temporary directory to store async-profiler.      | `/tmp`  | no       |

## Profiling behavior

The special label `__process_pid__` _must always_ be present in each target of `targets` and corresponds to the `PID` of
the process to profile.

After component startup, `pyroscope.java` creates a temporary directory under `tmp_dir` and extracts the
async-profiler binaries for both glibc and musl into the directory with the following layout.

```
/tmp/grafana-agent-asprof-glibc-{SHA1}/bin/asprof
/tmp/grafana-agent-asprof-glibc-{SHA1}/lib/libasyncProfiler.so
/tmp/grafana-agent-asprof-musl-{SHA1}/bin/asprof
/tmp/grafana-agent-asprof-musl-{SHA1}/lib/libasyncProfiler.so
```

After process profiling startup, the component detects libc type and copies according `libAsyncProfiler.so` into the
target process file system at the exact same path.

{{% admonition type="note" %}}
The `asprof` binary runs with root permissions.
If you change the `tmp_dir` configuration to something other than `/tmp`, then you must ensure that the
directory is only writable by root.
{{% /admonition %}}
## Blocks

The following blocks are supported inside the definition of
`pyroscope.java`:

| Hierarchy        | Block                | Description                            | Required |
|------------------|----------------------|----------------------------------------|----------|
| profiling_config | [profiling_config][] | Describes java profiling configuration. | no       |

[profiling_config]: #profiling_config-block

### profiling_config block

The `profiling_config` block describes how async-profiler is invoked.

The following arguments are supported:

| Name          | Type       | Description                                                                                             | Default | Required |
|---------------|------------|---------------------------------------------------------------------------------------------------------|---------|----------|
| `interval`    | `duration` | How frequently to collect profiles from the targets.                                                    | "60s"   | no       |
| `cpu`         | `bool`     | A flag to enable cpu profiling, using `itimer` async-profiler event.                                    | true    | no       |
| `sample_rate` | `int`      | CPU profiling sample rate. It is converted from Hz to interval and passed as `-i` arg to async-profiler. | 100     | no       |
| `alloc`       | `string`   | Allocation profiling sampling configuration  It is passed as `--alloc` arg to async-profiler.            | "512k"  | no       |
| `lock`        | `string`   | Lock profiling sampling configuration. It is passed as `--lock` arg to async-profiler.                   | "10ms"  | no       |

For more information on async-profiler configuration, see [profiler-options](https://github.com/async-profiler/async-profiler?tab=readme-ov-file#profiler-options)

## Exported fields

`pyroscope.java` does not export any fields that can be referenced by other
components.

## Component health

`pyroscope.java` is only reported as unhealthy when given an invalid
configuration. In those cases, exported fields retain their last healthy
values.

## Debug information

`pyroscope.java` does not expose any component-specific debug information.

## Debug metrics

`pyroscope.java` does not expose any component-specific debug metrics.

## Examples

### Profile every java process on the current host

```river
pyroscope.write "staging" {
  endpoint {
    url = "http://localhost:4040"
  }
}

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

discovery.relabel "java" {
  targets = discovery.process.all.targets
  rule {
    action = "keep"
    regex = ".*/java$"
    source_labels = ["__meta_process_exe"]
  }
}

pyroscope.java "java" {
  targets = discovery.relabel.java.output
  forward_to = [pyroscope.write.staging.receiver]
  profiling_config {
    interval = "60s"
    alloc = "512k"
    cpu = true
    sample_rate = 100
    lock = "1ms"
  }
}
```

<!-- START GENERATED COMPATIBLE COMPONENTS -->

## Compatible components

`pyroscope.java` can accept arguments from the following components:

- Components that export [Targets]({{< relref "../compatibility/#targets-exporters" >}})
- Components that export [Pyroscope `ProfilesReceiver`]({{< relref "../compatibility/#pyroscope-profilesreceiver-exporters" >}})


{{% admonition type="note" %}}

Connecting some components may not be sensible or components may require further configuration to make the 
connection work correctly. Refer to the linked documentation for more details.

{{% /admonition %}}

<!-- END GENERATED COMPATIBLE COMPONENTS -->