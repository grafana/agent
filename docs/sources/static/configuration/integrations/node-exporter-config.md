---
aliases:
- ../../../configuration/integrations/node-exporter-config/
- /docs/grafana-cloud/monitor-infrastructure/agent/static/configuration/integrations/node-exporter-config/
- /docs/grafana-cloud/send-data/agent/static/configuration/integrations/node-exporter-config/
canonical: https://grafana.com/docs/agent/latest/static/configuration/integrations/node-exporter-config/
description: Learn about node_exporter_config
title: node_exporter_config
---

# node_exporter_config

The `node_exporter_config` block configures the `node_exporter` integration,
which is an embedded version of
[`node_exporter`](https://github.com/prometheus/node_exporter)
and allows for collecting metrics from the UNIX system that `node_exporter` is
running on. It provides a significant amount of collectors that are responsible
for monitoring various aspects of the host system.

Note that if running the Agent in a container, you will need to bind mount
folders from the host system so the integration can monitor them. You can use
the example below, making sure to replace `/path/to/config.yaml` with a path on
your host machine where an Agent configuration file is:

```
docker run \
  --net="host" \
  --pid="host" \
  --cap-add=SYS_TIME \
  -v "/:/host/root:ro,rslave" \
  -v "/sys:/host/sys:ro,rslave" \
  -v "/proc:/host/proc:ro,rslave" \
  -v /tmp/agent:/etc/agent \
  -v /path/to/config.yaml:/etc/agent-config/agent.yaml \
  grafana/agent:{{< param "AGENT_RELEASE" >}} \
  --config.file=/etc/agent-config/agent.yaml
```

Use this configuration file for testing out `node_exporter` support, replacing
the `remote_write` settings with settings appropriate for you:

```yaml
server:
  log_level: info

metrics:
  wal_directory: /tmp/agent
  global:
    scrape_interval: 60s
    remote_write:
    - url: https://prometheus-us-central1.grafana.net/api/prom/push
      basic_auth:
        username: user-id
        password: api-token

integrations:
  node_exporter:
    enabled: true
    rootfs_path: /host/root
    sysfs_path: /host/sys
    procfs_path: /host/proc
    udev_data_path: /host/root/run/udev/data
```

For running on Kubernetes, ensure to set the equivalent mounts and capabilities
there as well:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: agent
spec:
  containers:
  - image: {{< param "AGENT_RELEASE" >}}
    name: agent
    args:
    - --config.file=/etc/agent-config/agent.yaml
    securityContext:
      capabilities:
        add: ["SYS_TIME"]
      privileged: true
      runAsUser: 0
    volumeMounts:
    - name: rootfs
      mountPath: /host/root
      readOnly: true
    - name: sysfs
      mountPath: /host/sys
      readOnly: true
    - name: procfs
      mountPath: /host/proc
      readOnly: true
  hostPID: true
  hostNetwork: true
  dnsPolicy: ClusterFirstWithHostNet
  volumes:
  - name: rootfs
    hostPath:
      path: /
  - name: sysfs
    hostPath:
      path: /sys
  - name: procfs
    hostPath:
      path: /proc
```

The manifest and Tanka configs provided by this repository do not have the
mounts or capabilities required for running this integration.

Some collectors only work on specific operating systems, documented in the
table below. Enabling a collector that is not supported by the operating system
the Agent is running on is a no-op.

| Name             | Description | OS | Enabled by default |
| ---------------- | ----------- | -- | ------------------ |
| arp              | Exposes ARP statistics from /proc/net/arp. | Linux | yes |
| bcache           | Exposes bcache statistics from /sys/fs/bcache. | Linux | yes |
| bonding          | Exposes the number of configured and active slaves of Linux bonding interfaces. | Linux | yes |
| boottime         | Exposes system boot time derived from the kern.boottime sysctl. | Darwin, Dragonfly, FreeBSD, NetBSD, OpenBSD, Solaris | yes |
| btrfs            | Exposes statistics on btrfs. | Linux | yes |
| buddyinfo        | Exposes statistics of memory fragments as reported by /proc/buddyinfo. | Linux | no |
| cgroups          | Exposes number of active and enabled cgroups. | Linux | no |
| conntrack        | Shows conntrack statistics (does nothing if no /proc/sys/net/netfilter/ present). | Linux | yes |
| cpu              | Exposes CPU statistics. | Darwin, Dragonfly, FreeBSD, Linux, Solaris, NetBSD | yes |
| cpufreq          | Exposes CPU frequency statistics. | Linux, Solaris | yes |
| devstat          | Exposes device statistics. | Dragonfly, FreeBSD | no |
| diskstats        | Exposes disk I/O statistics. | Darwin, Linux, OpenBSD | yes |
| dmi              | Exposes DMI information. | Linux | yes |
| drbd             | Exposes Distributed Replicated Block Device statistics (to version 8.4). | Linux | no |
| drm              | Exposes GPU card info from /sys/class/drm/card?/device | Linux | no |
| edac             | Exposes error detection and correction statistics. | Linux | yes |
| entropy          | Exposes available entropy. | Linux | yes |
| ethtool          | Exposes ethtool stats | Linux | no |
| exec             | Exposes execution statistics. | Dragonfly, FreeBSD | yes |
| fibrechannel     | Exposes FibreChannel statistics. | Linux | yes |
| filefd           | Exposes file descriptor statistics from /proc/sys/fs/file-nr. | Linux | yes |
| filesystem       | Exposes filesystem statistics, such as disk space used. | Darwin, Dragonfly, FreeBSD, Linux, OpenBSD | yes |
| hwmon            | Exposes hardware monitoring and sensor data from /sys/class/hwmon. | Linux | yes |
| infiniband       | Exposes network statistics specific to InfiniBand and Intel OmniPath configurations. | Linux | yes |
| interrupts       | Exposes detailed interrupts statistics. | Linux, OpenBSD | no |
| ipvs             | Exposes IPVS status from /proc/net/ip_vs and stats from /proc/net/ip_vs_stats. | Linux | yes |
| ksmd             | Exposes kernel and system statistics from /sys/kernel/mm/ksm. | Linux | no |
| lnstat           | Exposes Linux network cache stats | Linux | no |
| loadavg          | Exposes load average. | Darwin, Dragonfly, FreeBSD, Linux, NetBSD, OpenBSD, Solaris | yes |
| logind           | Exposes session counts from logind. | Linux | no |
| mdadm            | Exposes statistics about devices in /proc/mdstat (does nothing if no /proc/mdstat present). | Linux | yes |
| meminfo          | Exposes memory statistics. | Darwin, Dragonfly, FreeBSD, Linux, OpenBSD, NetBSD | yes |
| meminfo_numa     | Exposes memory statistics from /proc/meminfo_numa. | Linux | no |
| mountstats       | Exposes filesystem statistics from /proc/self/mountstats. Exposes detailed NFS client statistics. | Linux | no |
| netclass         | Exposes network interface info from /sys/class/net. | Linux | yes |
| netisr           | Exposes netisr statistics. | FreeBSD | yes |
| netdev           | Exposes network interface statistics such as bytes transferred. | Darwin, Dragonfly, FreeBSD, Linux, OpenBSD | yes |
| netstat          | Exposes network statistics from /proc/net/netstat. This is the same information as netstat -s. | Linux | yes |
| network_route    | Exposes network route statistics. | Linux | no |
| nfs              | Exposes NFS client statistics from /proc/net/rpc/nfs. This is the same information as nfsstat -c. | Linux | yes |
| nfsd             | Exposes NFS kernel server statistics from /proc/net/rpc/nfsd. This is the same information as nfsstat -s. | Linux | yes |
| ntp              | Exposes local NTP daemon health to check time. | any | no |
| nvme             | Exposes NVMe statistics. | Linux | yes |
| os               | Exposes os-release information. | Linux | yes |
| perf             | Exposes perf based metrics (Warning: Metrics are dependent on kernel configuration and settings). | Linux | no |
| powersupplyclass | Collects information on power supplies. | any | yes |
| pressure         | Exposes pressure stall statistics from /proc/pressure/. | Linux (kernel 4.20+ and/or CONFIG_PSI) | yes |
| processes        | Exposes aggregate process statistics from /proc. | Linux | no |
| qdisc            | Exposes queuing discipline statistics. | Linux | no |
| rapl             | Exposes various statistics from /sys/class/powercap. | Linux | yes |
| runit            | Exposes service status from runit. | any | no |
| schedstat        | Exposes task scheduler statistics from /proc/schedstat. | Linux | yes |
| selinux          | Exposes SELinux statistics. | Linux | yes |
| slabinfo         | Exposes slab statistics from `/proc/slabinfo`. | Linux | no |
| softirqs         | Exposes detailed softirq statistics from `/proc/softirqs`. | Linux | no |
| sockstat         | Exposes various statistics from /proc/net/sockstat. | Linux | yes |
| softnet          | Exposes statistics from /proc/net/softnet_stat. | Linux | yes |
| stat             | Exposes various statistics from /proc/stat. This includes boot time, forks and interrupts. | Linux | yes |
| supervisord      | Exposes service status from supervisord. | any | no |
| sysctl           | Expose sysctl values from `/proc/sys`. | Linux | no |
| systemd          | Exposes service and system status from systemd. | Linux | no |
| tapestats        | Exposes tape device stats. | Linux | yes |
| tcpstat          | Exposes TCP connection status information from /proc/net/tcp and /proc/net/tcp6. (Warning: the current version has potential performance issues in high load situations). | Linux | no |
| textfile         | Collects metrics from files in a directory matching the filename pattern *.prom. The files must be using the text format defined here: https://prometheus.io/docs/instrumenting/exposition_formats/ | any | yes |
| thermal          | Exposes thermal statistics. | Darwin | yes |
| thermal_zone     | Exposes thermal zone & cooling device statistics from /sys/class/thermal. | Linux | yes |
| time             | Exposes the current system time. | any | yes |
| timex            | Exposes selected adjtimex(2) system call stats. | Linux | yes |
| udp_queues       | Exposes UDP total lengths of the rx_queue and tx_queue from /proc/net/udp and /proc/net/udp6. | Linux | yes |
| uname            | Exposes system information as provided by the uname system call. | Darwin, FreeBSD, Linux, OpenBSD, NetBSD | yes |
| vmstat           | Exposes statistics from /proc/vmstat. | Linux | yes |
| wifi             | Exposes WiFi device and station statistics. | Linux | no |
| xfs              | Exposes XFS runtime statistics. | Linux (kernel 4.4+) | yes |
| zfs              | Exposes ZFS performance statistics. | Linux, Solaris | yes |
| zoneinfo         | Exposes zone stats. | Linux | no |

```yaml
  # Enables the node_exporter integration, allowing the Agent to automatically
  # collect system metrics from the host UNIX system.
  [enabled: <boolean> | default = false]

  # Sets an explicit value for the instance label when the integration is
  # self-scraped. Overrides inferred values.
  #
  # The default value for this integration is inferred from the agent hostname
  # and HTTP listen port, delimited by a colon.
  [instance: <string>]

  # Automatically collect metrics from this integration. If disabled,
  # the node_exporter integration will be run but not scraped and thus not remote-written. Metrics for the
  # integration will be exposed at /integrations/node_exporter/metrics and can
  # be scraped by an external process.
  [scrape_integration: <boolean> | default = <integrations_config.scrape_integrations>]

  # How often should the metrics be collected? Defaults to
  # prometheus.global.scrape_interval.
  [scrape_interval: <duration> | default = <global_config.scrape_interval>]

  # The timtout before considering the scrape a failure. Defaults to
  # prometheus.global.scrape_timeout.
  [scrape_timeout: <duration> | default = <global_config.scrape_timeout>]

  # Allows for relabeling labels on the target.
  relabel_configs:
    [- <relabel_config> ... ]

  # Relabel metrics coming from the integration, allowing to drop series
  # from the integration that you don't care about.
  metric_relabel_configs:
    [ - <relabel_config> ... ]

  # How frequent to truncate the WAL for this integration.
  [wal_truncate_frequency: <duration> | default = "60m"]

  # Monitor the exporter itself and include those metrics in the results.
  [include_exporter_metrics: <boolean> | default = false]

  # Optionally defines the list of enabled-by-default collectors.
  # Anything not provided in the list below will be disabled by default,
  # but requires at least one element to be treated as defined.
  #
  # This is useful if you have a very explicit set of collectors you wish
  # to run.
  set_collectors:
    - [<string>]

  # Additional collectors to enable on top of the default set of enabled
  # collectors or on top of the list provided by set_collectors.
  #
  # This is useful if you have a few collectors you wish to run that are
  # not enabled by default, but do not want to explicitly provide an entire
  # list through set_collectors.
  enable_collectors:
    - [<string>]

  # Additional collectors to disable on top of the default set of disabled
  # collectors. Takes precedence over enable_collectors.
  #
  # This is useful if you have a few collectors you do not want to run that
  # are enabled by default, but do not want to explicitly provide an entire
  # list through set_collectors.
  disable_collectors:
    - [<string>]

  # procfs mountpoint.
  [procfs_path: <string> | default = "/proc"]

  # sysfs mountpoint.
  [sysfs_path: <string> | default = "/sys"]

  # rootfs mountpoint. If running in docker, the root filesystem of the host
  # machine should be mounted and this value should be changed to the mount
  # directory.
  [rootfs_path: <string> | default = "/"]

  # udev data path needed for diskstats from Node exporter. When running
  # in Kubernetes it should be set to /host/root/run/udev/data.
  [udev_data_path: <string> | default = "/run/udev/data"]

  # Expose expensive bcache priority stats.
  [enable_bcache_priority_stats: <boolean>]

  # Regexp of `bugs` field in cpu info to filter.
  [cpu_bugs_include: <string>]

  # Enable the node_cpu_guest_seconds_total metric.
  [enable_cpu_guest_seconds_metric: <boolean> | default = true]

  # Enable the cpu_info metric for the cpu collector.
  [enable_cpu_info_metric: <boolean> | default = true]

  # Regexp of `flags` field in cpu info to filter.
  [cpu_flags_include: <string>]

  # Regexp of devices to ignore for diskstats.
  [diskstats_device_exclude: <string> | default = "^(ram|loop|fd|(h|s|v|xv)d[a-z]|nvme\\d+n\\d+p)\\d+$"]

  # Regexp of devices to include for diskstats. If set, the diskstat_device_exclude field is ignored.
  [diskstats_device_include: <string>]

  # Regexp of ethtool devices to exclude (mutually exclusive with ethtool_device_include)
  [ethtool_device_exclude: <string>]

  # Regexp of ethtool devices to include (mutually exclusive with ethtool_device_exclude)
  [ethtool_device_include: <string>]

  # Regexp of ethtool stats to include.
  [ethtool_metrics_include: <string> | default = ".*"]

  # Regexp of mount points to ignore for filesystem collector.
  [filesystem_mount_points_exclude: <string> | default = "^/(dev|proc|sys|var/lib/docker/.+)($|/)"]

  # Regexp of filesystem types to ignore for filesystem collector.
  [filesystem_fs_types_exclude: <string> | default = "^(autofs|binfmt_misc|bpf|cgroup2?|configfs|debugfs|devpts|devtmpfs|fusectl|hugetlbfs|iso9660|mqueue|nsfs|overlay|proc|procfs|pstore|rpc_pipefs|securityfs|selinuxfs|squashfs|sysfs|tracefs)$"]

  # How long to wait for a mount to respond before marking it as stale.
  [filesystem_mount_timeout: <duration> | default = "5s"]

  # Array of IPVS backend stats labels.
  #
  # The default is [local_address, local_port, remote_address, remote_port, proto, local_mark].
  ipvs_backend_labels:
    [- <string>]

  # NTP server to use for ntp collector
  [ntp_server: <string> | default = "127.0.0.1"]

  # NTP protocol version
  [ntp_protocol_version: <int> | default = 4]

  # Certify that the server address is not a public ntp server.
  [ntp_server_is_local: <boolean> | default = false]

  # IP TTL to use wile sending NTP query.
  [ntp_ip_ttl: <int> | default = 1]

  # Max accumulated distance to the root.
  [ntp_max_distance: <duration> | default = "3466080us"]

  # Offset between local clock and local ntpd time to tolerate.
  [ntp_local_offset_tolerance: <duration> | default = "1ms"]

  # Regexp of net devices to ignore for netclass collector.
  [netclass_ignored_devices: <string> | default = "^$"]

  # Ignore net devices with invalid speed values. This will default to true in
  # node_exporter 2.0.
  [netclass_ignore_invalid_speed_device: <boolean> | default = false]

  # Enable collecting address-info for every device.
  [netdev_address_info: <boolean>]

  # Regexp of net devices to exclude (mutually exclusive with include)
  [netdev_device_exclude: <string> | default = ""]

  # Regexp of net devices to include (mutually exclusive with exclude)
  [netdev_device_include: <string> | default = ""]

  # Regexp of fields to return for netstat collector.
  [netstat_fields: <string> | default = "^(.*_(InErrors|InErrs)|Ip_Forwarding|Ip(6|Ext)_(InOctets|OutOctets)|Icmp6?_(InMsgs|OutMsgs)|TcpExt_(Listen.*|Syncookies.*|TCPSynRetrans|TCPTimeouts)|Tcp_(ActiveOpens|InSegs|OutSegs|OutRsts|PassiveOpens|RetransSegs|CurrEstab)|Udp6?_(InDatagrams|OutDatagrams|NoPorts|RcvbufErrors|SndbufErrors))$"]

  # List of CPUs from which perf metrics should be collected.
  [perf_cpus: <string> | default = ""]

  # Array of perf tracepoints that should be collected.
  perf_tracepoint:
    [- <string>]

  # Disable perf hardware profilers.
  [perf_disable_hardware_profilers: <boolean> | default = false]

  # Perf hardware profilers that should be collected.
  perf_hardware_profilers:
    [- <string>]

  # Disable perf software profilers.
  [perf_disable_software_profilers: <boolean> | default = false]

  # Perf software profilers that should be collected.
  perf_software_profilers:
    [- <string>]

  # Disable perf cache profilers.
  [perf_disable_cache_profilers: <boolean> | default = false]

  # Perf cache profilers that should be collected.
  perf_cache_profilers:
    [- <string>]

  # Regexp of power supplies to ignore for the powersupplyclass collector.
  [powersupply_ignored_supplies: <string> | default = "^$"]

  # Path to runit service directory.
  [runit_service_dir: <string> | default = "/etc/service"]

  # XML RPC endpoint for the supervisord collector.
  #
  # Setting SUPERVISORD_URL in the environment will override the default value.
  # An explicit value in the YAML config takes precedence over the environment
  # variable.
  [supervisord_url: <string> | default = "http://localhost:9001/RPC2"]

  # Numeric sysctl values to expose.
  # For sysctl with multiple numeric values,
  # an optional mapping can be given to expose each value as its own metric.
  sysctl_include:
    [- <string>]

  # String sysctl values to expose.
  sysctl_include_info:
    [- <string>]

  # Regexp of systemd units to include. Units must both match include and not
  # match exclude to be collected.
  [systemd_unit_include: <string> | default = ".+"]

  # Regexp of systemd units to exclude. Units must both match include and not
  # match exclude to be collected.
  [systemd_unit_exclude: <string> | default = ".+\\.(automount|device|mount|scope|slice)"]

  # Enables service unit tasks metrics unit_tasks_current and unit_tasks_max
  [systemd_enable_task_metrics: <boolean> | default = false]

  # Enables service unit metric service_restart_total
  [systemd_enable_restarts_metrics: <boolean> | default = false]

  # Enables service unit metric unit_start_time_seconds
  [systemd_enable_start_time_metrics: <boolean> | default = false]

  # Regexp of tapestats devices to ignore.
  [tapestats_ignored_devices: <string> | default = "^$"]

  # Directory to read *.prom files from for the textfile collector.
  [textfile_directory: <string> | default = ""]

  # Regexp of fields to return for the vmstat collector.
  [vmstat_fields: <string> | default = "^(oom_kill|pgpg|pswp|pg.*fault).*"]
```
