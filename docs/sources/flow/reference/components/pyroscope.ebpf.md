---
title: pyroscope.ebpf
labels:
  stage: beta
---

# pyroscope.ebpf

{{< docs/shared lookup="flow/stability/beta.md" source="agent" >}}

`pyroscope.ebpf` configures an ebpf profiling job for the current host. The collected performance profiles are forwarded
to
the list of receivers passed in
`forward_to`.

Multiple `pyroscope.ebpf` components can be specified by giving them different labels, however it is not recommended.

## Usage

```river
pyroscope.ebpf "LABEL" {
  forward_to = RECEIVER_LIST
}
```

## Arguments

The component configures and starts a new ebpf profiling job to collect performance profiles from the current host.

The following arguments can be used to configure a `pyroscope.ebpf`. Only the
`forward_to` field is required and any omitted fields take their default
values.

| Name                      | Type                     | Description                                                            | Default | Required |
|---------------------------|--------------------------|------------------------------------------------------------------------|---------|----------|
| `forward_to`              | `list(ProfilesReceiver)` | List of receivers to send collected profiles to.                       |         | yes      |  
| `targets`                 | `list(map(string))`      | List of targets to group profiles by container id                      |         | no       |   
| `default_target`          | `map(string)`            | Default target to use when a PID is not associated with a container id |         | no       |    
| `targets_only`            | `bool`                   | A flag to ignore profiles not associated with a container id           | true    | no       |      
| `collect_interval`        | `duration`               | How frequently to collect profiles                                     | `15s`   | no       |       
| `sample_rate`             | `int`                    | How many times per second to collect profile samples                   | 97      | no       |     
| `pid_cache_size`          | `int`                    | The size of the pid -> proc symbol table LRU cache                     | 32      | no       |      
| `elf_cache_size`          | `int`                    | The size of the elf file -> symbols LRU cache                          | 128     | no       |       
| `container_id_cache_size` | `int`                    | The size of the pid -> container ID  LRU cache                         | 64      | no       |       
| `collect_user_profile`    | `bool`                   | A flag to enable/disable collection of userspace profiles              | true    | no       |       
| `collect_kernel_profile`  | `bool`                   | A flag to enable/disable collection of kernelspace profiles            | true    | no       |       

## Exported fields

`pyroscope.ebpf` does not export any fields that can be referenced by other
components.

## Component health

`pyroscope.ebpf` is only reported as unhealthy if given an invalid
configuration.

## Debug information

`pyroscope.ebpf` does not expose any component-specific debug information.

## Debug metrics

* `pyroscope_fanout_latency` (histogram): Write latency for sending to direct and indirect components.
* `pyroscope_ebpf_pid_cache_hit_total` (counter): Total number of ebpf symbolizer pid cache hit.
* `pyroscope_ebpf_pid_cache_miss_total` (counter): Total number of ebpf symbolizer pid cache miss.
* `pyroscope_ebpf_elf_cache_build_id_hit_total` (counter): Total number of ebpf symbolizer elf cache (build-id) hit.
* `pyroscope_ebpf_elf_cache_build_id_miss_total` (counter): Total number of ebpf symbolizer elf cache (build-id) miss.
* `pyroscope_ebpf_elf_cache_stat_hit_total` (counter): Total number of ebpf symbolizer elf cache (stat) hit.
* `pyroscope_ebpf_elf_cache_stat_miss_total` (counter): Total number of ebpf symbolizer elf cache (stat) miss.
* `pyroscope_ebpf_container_id_cache_hit_total` (counter): Total number of ebpf target finder container id cache hit.
* `pyroscope_ebpf_container_id_cache_miss_total` (counter): Total number of ebpf target finder container id cache miss.

## Profile collecting behavior

The `pyroscope.ebpf` component is designed to collect stacktraces associated with a process running on the current host.
Stack traces are collected according to the defined `sample_rate`, meaning traces are gathered this many times
per second.

The following labels are automatically injected to the collected profiles if not found and can help pin down a
profiling target.

| Label              | Description                                                                                                                                                                                            |
|--------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `service_name`     | Pyroscope service name. It's automatically selected from discovery meta labels if possible. Otherwise defaults to the fully formed component name or a value from `pyroscope.ebpf.service_name` field. |
| `__name__`         | pyroscope metric name. Defaults to `process_cpu`.                                                                                                                                                      |
| `__container_id__` | The container ID derived from target.                                                                                                                                                                  |

### Container ID

Each collected stack trace is then associated with a specified target from the targets list, determined by a
container ID. This association process involves checking the `__container_id__`, `__meta_docker_container_id`,
and `__meta_kubernetes_pod_container_id` labels of a target against the `/proc/{pid}/cgroup` of a process.

If a corresponding container ID is found, the stack traces are aggregated per target based on the container ID.
However, if a container ID is not identified, the stack trace is then associated with a `default_target`.

If the `targets_only` is set to true, any stacktrace not associated with a listed target
(i.e., those that would typically be associated with the default_target) are disregarded.

### Service name

The special label `service_name` is required and must always be present. If it's not provided, it is
sourced from `__meta_kubernetes_pod_annotation_pyroscope_io_service_name` which is a
`pyroscope.io/service_name` pod annotation. If `service_name` is not present, the target is considered as misconfigured
and discarded.

## Example

### Kubernetes discovery

In the following example, performance profiles are collected from pods on the same node, discovered using
`discovery.kubernetes`. Pod selection relies on the `HOSTNAME` environment variable, which is a pod name if the agent is
used as a Grafana agent helm chart. Service name is set to `{namespace}/{container_name}` from kubernetes meta labels.

```river
discovery.kubernetes "all_pods" {
  role = "pod"
}

discovery.relabel "local_pods" {
  targets = discovery.kubernetes.all_pods.targets
  rule {
    action = "keep"
    regex = env("HOSTNAME")
    source_labels = ["__meta_kubernetes_pod_node_name"]
  }
  rule {
    action = "replace"
    replacement = "${1}/${2}"
    separator = "/"
    source_labels = ["__meta_kubernetes_namespace", "__meta_kubernetes_pod_container_name"]
    target_label = "service_name"
  }
}
pyroscope.ebpf "local_pods" {
  forward_to = [ pyroscope.write.endpoint.receiver ]
  targets = discovery.relabel.local_pods.output
}

pyroscope.write "endpoint" {
  endpoint {
    url = "http://pyroscope:4100"
  }
}
```

### Docker discovery

The following example collects performance profiles from containers discovered by `discovery.docker` and ignores all
other profiles collected from outside any docker container. `service_name` label is set to
`__meta_docker_container_name` label.

```river
discovery.docker "linux" {
  host = "unix:///var/run/docker.sock"
}

discovery.relabel "local_containers" {
  targets = discovery.docker.linux.targets
  rule {
    action = "replace"
    source_labels = ["__meta_docker_container_name"]
    target_label = "service_name"
  }
}

pyroscope.write "staging" {
  endpoint {
    url = "http://pyroscope:4100"
  }
}

pyroscope.ebpf "default" {  
  forward_to   = [ pyroscope.write.staging.receiver ]
  targets      = discovery.relabel.local_containers.output
}
```
