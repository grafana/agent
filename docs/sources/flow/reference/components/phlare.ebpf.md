---
title: phlare.ebpf
labels:
  stage: beta
---

# phlare.ebpf

{{< docs/shared lookup="flow/stability/beta.md" source="agent" >}}

`phlare.ebpf` configures an ebpf profiling job for the current host. The collected performance profiles are forwarded to
the list of receivers passed in
`forward_to`.

Multiple `phlare.ebpf` components can be specified by giving them different labels, however it is not recommended.

## Usage

```river
phlare.ebpf "LABEL" {
  forward_to = RECEIVER_LIST
}
```

## Arguments

The component configures and starts a new ebpf profiling job to collect performance profiles from the current host.

The following arguments can be used to configure a `phlare.ebpf`. Only the
`forward_to` field is required and any omitted fields take their default
values.

| Name               | Type                     | Description                                                            | Default        | Required |
|--------------------|--------------------------|------------------------------------------------------------------------|----------------|----------|
| `forward_to`       | `list(ProfilesReceiver)` | List of receivers to send collected profiles to.                       |                | yes      |  
| `targets`          | `list(map(string))`      | List of targets to group profiles by container id                      |                | no       |   
| `default_target`   | `map(string)`            | Default target to use when a PID is not associated with a container id |                | no       |    
| `targets_only`     | `bool`                   | A flag to ignore profiles not associated with a container id           | false          | no       |     
| `job_name`         | `string`                 | The job name to override the job label with.                           | component name | no       |      
| `collect_interval` | `duration`               | How frequently to collect profiles                                     | `10s`          | no       |       
| `sample_rate`      | `int`                    | How many times per second to collect samples                           | 100            | no       |     
| `pid_cache_size`   | `int`                    | The size of the symbolizer's per pid LRU cache                         | 64             | no       |      
| `elf_cache_size`   | `int`                    | The size of the symbolizer's per elf LRU cache                         | 128            | no       |       

## Exported fields

`phlare.ebpf` does not export any fields that can be referenced by other
components.

## Component health

`phlare.ebpf` is only reported as unhealthy if given an invalid
configuration.

## Debug information

`phlare.ebpf` does not expose any component-specific debug information.

## Debug metrics

* `phlare_fanout_latency` (histogram): Write latency for sending to direct and indirect components.

## Profile collecting behavior

The `phlare.ebpf` component is designed to collect stacktraces associated with a process running on the current host.
Stack traces are collected according to the defined `sample_rate`, meaning traces are gathered this many times
per second.

The following labels are automatically injected to the collected profiles if not found and can help pin down a
profiling target.

| Label              | Description                                                                                                                                                                                         |
|--------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `service_name`     | Pyroscope service name. It's automatically selected from discovery meta labels if possible. Otherwise defaults to the fully formed component name or a value from `phlare.ebpf.service_name` field. |
| `__name__`         | Phlare metric name. Defaults to `process_cpu`.                                                                                                                                                      |
| `__container_id__` | The container ID derived from target.                                                                                                                                                               |

### Container ID

Each collected stack trace is then associated with a specified target from the targets list, determined by a
container ID. This association process involves checking the `__container_id__`, `__meta_docker_container_id`,
and `__meta_kubernetes_pod_container_id` labels of a target against the `/proc/{pid}/cgroup` of a process.

If a corresponding container ID is found, the stack traces are aggregated per target based on the container ID.
However, if a container ID is not identified, the stack trace is then associated with a `default_target`.

If the `targets_only` is set to true, any stacktrace not associated with a listed target
(i.e., those that would typically be associated with the default_target) are disregarded.

### Service name

Every target must have a `service_name` label. If this label is absent from a target, the system will attempt to derive
it from discovery meta labels.

For targets discovered via `discovery.kubernetes`, the `service_name` is constructed by combining
`__meta_kubernetes_namespace` and `__meta_kubernetes_pod_name`, separated by a "/".

For targets discovered through `discovery.docker`, the `service_name` is set as `__meta_docker_container_name`.

If it's not possible to derive a service name from these methods, the system will check if a `phlare.ebpf.service_name`
field. If the field is non-empty, it will be used.

If all else fails, the service name defaults to the fully formed component name.

## Example

### Kubernetes discovery

In the following example, performance profiles are collected from containers on the same node, discovered using
`discovery.kubernetes`. Container selection relies on the `HOSTNAME` environmental variable, which, when the agent is
used as a Grafana agent helm chart, contains the pod name.

Since no explicit `service_name` is provided, it is automatically derived from the `__meta_kubernetes_namespace` and
`__meta_kubernetes_pod_name` labels. Any profiles that can't be linked to a specific container are associated with a
default target. The service_name for default target is set to `phlare_default_service_name`.

```river
discovery.kubernetes "pods" {
  role = "pod"
}

discovery.relabel "local_pods" {
  targets = discovery.kubernetes.pods.targets

  rule {
    source_labels = ["__meta_kubernetes_pod_node_name"]
    regex = regex_quote(env("HOSTNAME"))  
    action = "keep"
  }
}

phlare.write "staging" {
  endpoint {
    url = "http://64.176.81.68:4100"
  }
}

phlare.ebpf "default" {
  targets      = discovery.relabel.local_pods.output
  forward_to   = [ phlare.write.staging.receiver ]
  service_name = "phlare_default_service_name"
}

```

### Docker discovery

The following example collects performance profiles from containers discovered by `discovery.docker` and ignores all
other profiles collected from outside any docker container. `service_name` label is automatically selected
from `__meta_docker_container_name` label.

```river
discovery.docker "linux" {
  host = "unix:///var/run/docker.sock"
}

phlare.write "staging" {
  endpoint {
    url = "http://phlare:4100"
  }
}

phlare.ebpf "default" {  
  forward_to   = [ phlare.write.staging.receiver ]
  targets      = discovery.docker.linux.targets
  targets_only = true
}
```

### Static targets

In this example, performance profiles are collected from a single, statically defined container. As no explicit
service_name is provided, it defaults to `phlare.ebpf.default`.

```river
phlare.write "staging" {
  endpoint {
    url = "http://phlare:4100"
  }
}

phlare.ebpf "default" {
  forward_to   = [ phlare.write.staging.receiver ]
  targets      = [ {"__container_id__" = "495b22a24cf6a71b3d048d900074d0f44968217d919ef370aff46bdede79415f"} ]
  targets_only = true
}
```
