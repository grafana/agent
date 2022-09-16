---
aliases:
- /docs/agent/latest/flow/reference/components/prometheus.integrations.node_exporter
title: prometheus.integrations.node_exporter
---

# prometheus.integrations.node_exporter
The component embeds the popular
[node_exporter](https://github.com/prometheus/node_exporter) which exposes a 
wide variety of hardware and OS metrics.

The `node_exporter` itself is comprised of various _collectors_, which can be enabled and disabled at will.

## Example
```river
prometheus.integrations.node_exporter {
}

// Configure a prometheus.scrape component to collect node_exporter metrics.
prometheus.scrape "demo" {
  targets    = prometheus.integrations.node_exporter.targets
  forward_to = [ /* ... */ ]
}
```

## Arguments
The following arguments can be used to configure the exporter's behavior.
All arguments are optional. Omitted fields take their default values.

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`include_exporter_metrics` | boolean      | Whether metrics about the exporter itself should be reported | false | no 
`procfs_path`              | string       | The procfs mountpoint. | `/proc` | no
`sysfs_path`               | string       | The sysfs mountpoint.  | `sys`   | no
`rootfs_path`              | string       | Specify a prefix for accessing the host filesystem. | `/` | no
`enable_collectors`        | list(string) | Collectors to mark as enabled.  | | no
`disable_collectors`       | list(string) | Collectors to mark as disabled. | | no
`set_collectors`           | list(string) | Overrides the default set of enabled collectors with the collectors listed. | | no

If running in Docker, the root filesystem of the host machine should be mounted and `rootfs_path` should be changed to the mount directory.

Additionally, the following subblocks are supported for configuring collector-specific options.

Name | Description | Required
---- | ----------- | --------
[`relabel_config`](#relabel_config-block) | Relabeling steps to apply to targets | no

[`bcache`](#bcache-block) | Configures the bcache collector | no
[`cpu`](#cpu-block) | Configures the cpu collector | no
[`disk`](#disk-block) | Configures the diskstats collector | no
[`ethtool`](#ethtool-block) | Configures the ethtool collector | no
[`filesystem`](#filesystem-block) | Configures the filesystem collector | no
[`ipvs`](#ipvs-block) | Configures the ipvs collector | no
[`ntp`](#ntp-block) | Configures the ntp collector | no
[`netclass`](#netclass-block) | Configures the netclass collector | no
[`netdev`](#netdev-block) | Configures the netdev collector | no
[`netstat`](#netstat-block) | Configures the netstat collector | no
[`perf`](#perf-block) | Configures the perf collector | no
[`powersupply`](#powersupply-block) | Configures the powersupply collector | no
[`runit`](#runit-block) | Configures the runit collector | no
[`supervisord`](#supervisord-block) | Configures the supervisord collector | no
[`systemd`](#systemd-block) | Configures the systemd collector | no
[`tapestats`](#tapestats-block) | Configures the tapestats collector | no
[`textfile`](#textfile-block) | Configures the textfile collector | no
[`vmstat`](#vmstat-block) | Configures the vmstat collector | no






### `bcache` block
Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`priority_stats` | boolean |  Enable exposing of expensive bcache priority stats | false | no

### `cpu` block
Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`guest`         | boolean | Enable the node_cpu_guest_seconds_total metric. | true | no
`info`          | boolean | Enable the cpu_info metric for the cpu collector. | true | no
`bugs_include`  | string  | Regexp of `bugs` field in cpu info to filter. | | no
`flags_include` | string  | Regexp of `flags` field in cpu info to filter. | no

### `disk` block
Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`ignored_devices` | string | Regexp of devices to ignore for diskstats | `"^(ram|loop|fd|(h|s|v|xv)d[a-z]|nvme\\d+n\\d+p)\\d+$"` | no

### `ethtool` block
Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`device_exclude` | string | Regexp of ethtool devices to exclude (mutually exclusive with ethtool_device_include). | | no
`device_include` | string | Regexp of ethtool devices to include (mutually exclusive with ethtool_device_exclude). | | no
`metrics_include`| string | Regexp of ethtool stats to include. | `.*` | no

### `filesystem` block
Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`fs_types_exclude`     | string   | Regexp of filesystem types to ignore for filesystem collector.| `"^(autofs|binfmt_misc|bpf|cgroup2?|configfs|debugfs|devpts|devtmpfs|fusectl|hugetlbfs|iso9660|mqueue|nsfs|overlay|proc|procfs|pstore|rpc_pipefs|securityfs|selinuxfs|squashfs|sysfs|tracefs)$"` | no
`mount_points_exclude` | string   | Regexp of mount points to ignore for filesystem collector. | `"^/(dev|proc|sys|var/lib/docker/.+)($|/)"` | no
`mount_timeout`        | duration | How long to wait for a mount to respond before marking it as stale. | "5s" | no

### `ipvs` block
Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`backend_labels` | list(string) | Array of IPVS backend stats labels. | `[local_address, local_port, remote_address, remote_port, proto, local_mark]` | no
### `ntp` block
### `netclass` block
### `netdev` block
### `netstat` block
### `perf` block
### `powersupply` block
### `runit` block
### `supervisord` block
### `systemd` block
### `tapestats` block
### `textfile` block
### `vmstat` block


## Exported fields
The following fields are exported and can be referenced by other components.

Name      | Type                | Description 
--------- | ------------------- | ----------- 
`targets` | `list(map(string))` | The targets where the `node_exporter` metrics will be exposed to.

For example, the `targets` could either be passed to a `prometheus.relabel`
component to rewrite the metrics' label set, or to a `prometheus.scrape`
component that collects the exposed metrics.

## Component health

`prometheus.integrations.node_exporter` is only reported as unhealthy if given
an invalid configuration. In those cases, exported fields are kept at their
last healthy values.

## Debug information

`prometheus.relabel` does not expose any component-specific debug information.

## Debug metrics

`prometheus.relabel` does not expose any component-specific debug metrics.

