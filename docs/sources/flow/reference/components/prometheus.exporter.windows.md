---
# NOTE(rfratto): the title below has zero-width spaces injected into it to
# prevent it from overflowing the sidebar on the rendered site. Be careful when
# modifying this section to retain the spaces.
#
# Ideally, in the future, we can fix the overflow issue with css rather than
# injecting special characters.

title: prometheus.exporter.windows
---

# prometheus.exporter.windows
The `prometheus.exporter.windows` component embeds
[windows_exporter](https://github.com/prometheus-community/windows_exporter) which exposes a
wide variety of hardware and OS metrics for Windows-based systems.

The `windows_exporter` itself comprises various _collectors_, which can be
enabled and disabled at will. For more information on collectors, refer to the
[`collectors-list`](#collectors-list) section.

## Usage

```river
prometheus.exporter.windows "LABEL" { 
}
```

## Arguments
The following arguments can be used to configure the exporter's behavior.
All arguments are optional. Omitted fields take their default values.

| Name                 | Type             | Description                               | Default | Required |
|----------------------|------------------|-------------------------------------------|---------|----------|
| `enabled_collectors` | `list(string)`   | List of collectors to enable.             | `["cpu","cs","logical_disk","net","os","service","system"]` | no       |
| `timeout`            | `duration`       | Configure timeout for collecting metrics. | `4m`    | no       |

`enabled_collectors` defines a hand-picked list of enabled-by-default
collectors. If set, anything not provided in that list is disabled by
default. See the [Collectors list](#collectors-list) for the default set.

## Blocks

The following blocks are supported inside the definition of
`prometheus.exporter.windows` to configure collector-specific options:

Hierarchy      | Name               | Description                              | Required
---------------|--------------------|------------------------------------------|----------
dfsr           | [dfsr][]           | Configures the iis collector.            | no       
exchange       | [exchange][]       | Configures the exchange collector.       | no
iis            | [iis][]            | Configures the iis collector.            | no
logical_disk   | [logical_disk][]   | Configures the logical_disk collector.   | no       
msmq           | [msmq][]           | Configures the msmq collector.           | no
mssql          | [mssql][]          | Configures the mssql collector.          | no
network        | [network][]        | Configures the network collector.        | no
process        | [process][]        | Configures the process collector.        | no
scheduled_task | [scheduled_task][] | Configures the scheduled_task collector. | no
service        | [service][]        | Configures the service collector.        | no
smtp           | [smtp][]           | Configures the smtp collector.           | no
text_file      | [text_file][]      | Configures the text_file collector.      | no

[exchange]: #exchange-block
[iis]: #iis-block
[text_file]: #textfile-block
[smtp]: #smtp-block
[service]: #service-block
[process]: #process-block
[network]: #network-block
[mssql]: #mssql-block
[msmq]: #msmq-block
[logical_disk]: #logicaldisk-block

### dfsr block
Name | Type     | Description | Default | Required
---- |----------| ----------- | ------- | --------
`source_enabled` | `list(string)` | Comma-separated list of DFSR Perflib sources to use. | `["connection","folder","volume"]` | no


### exchange block
Name | Type     | Description | Default | Required
---- |----------| ----------- | ------- | --------
`enabled_list` | `string` | Comma-separated list of collectors to use. | `""` | no

The collectors specified by `enabled_list` can include the following:

- `ADAccessProcesses`
- `TransportQueues`
- `HttpProxy`
- `ActiveSync`
- `AvailabilityService`
- `OutlookWebAccess`
- `Autodiscover`
- `WorkloadManagement`
- `RpcClientAccess` 

For example, `enabled_list` may be set to `"AvailabilityService,OutlookWebAccess"`. 


### iis block
Name | Type     | Description | Default | Required
---- |----------| ----------- | ------- | --------
`app_blacklist` | `string` | Regular expression of applications to ignore. |  | no
`app_whitelist` | `string` | Regular expression of applications to report on. |  | no
`site_blacklist` | `string` | Regular expression of sites to ignore. |  | no
`site_whitelist` | `string` | Regular expression of sites to report on. |  | no

### text_file block
Name | Type     | Description | Default | Required
---- |----------| ----------- | ------- | --------
`text_file_directory` | `string` | The directory containing the files to be ingested. | `C:\Program Files\windows_exporter\textfile_inputs` | no

When `text_file_directory` is set, only files with the extension `.prom` inside the specified directory are read. Each `.prom` file found must end with an empty line feed to work properly.  


### smtp block
Name | Type     | Description | Default | Required
---- |----------| ----------- | ------- | --------
`blacklist` | `string` | Regexp of virtual servers to ignore. |  | no
`whitelist` | `string` | Regexp of virtual servers to include. | `".+"` | no

For a server name to be included, it must match the regular expression specified by `whitelist` and must _not_ match the regular expression specified by `blacklist`. 

### service block
Name | Type     | Description | Default | Required
---- |----------| ----------- | ------- | --------
`where_clause` | `string` | WQL 'where' clause to use in WMI metrics query. |  | no

The `where_clause` argument can be used to limit the response to the services you specify, reducing the size of the response.


### process block
Name | Type     | Description | Default | Required
---- |----------| ----------- | ------- | --------
`blacklist` | `string` | Regular expression of processes to exclude. |  | no
`whitelist` | `string` | Regular expression of processes to include. | `".*"` | no

Processes must match the regular expression specified by `whitelist` and must _not_ match the regular expression specified by `blacklist` to be included.

### network block
Name | Type     | Description | Default | Required
---- |----------| ----------- | ------- | --------
`blacklist` | `string` | Regular expression of NIC:s to exclude. |  | no
`whitelist` | `string` | Regular expression of NIC:s to include. | `".*"` | no

NIC names must match the regular expression specified by `whitelist` and must _not_ match the regular expression specified by `blacklist` to be included.

### mssql block
Name | Type     | Description | Default | Required
---- |----------| ----------- | ------- | --------
`enabled_classes` | `list(string)` | Comma-separated list of MSSQL WMI classes to use. | `["accessmethods", "availreplica", "bufman", "databases", "dbreplica", "genstats", "locks", "memmgr", "sqlstats", "sqlerrorstransactions"]` | no

### msmq block
Name | Type     | Description | Default | Required
---- |----------| ----------- | ------- | --------
`where_clause` | `string` | WQL 'where' clause to use in WMI metrics query. |  | no

Specifying `enabled_classes` is useful to limit the response to the MSMQs you specify, reducing the size of the response. 


### logical_disk block
Name | Type     | Description | Default | Required
---- |----------| ----------- | ------- | --------
`blacklist` | `string` | Regular expression of volumes to exclude. |  | no
`whitelist` | `string` | Regular expression of volumes to include. | `".+"` | no

Volume names must match the regular expression specified by `whitelist` and must _not_ match the regular expression specified by `blacklist` to be included.

## Exported fields
The following fields are exported and can be referenced by other components:

Name      | Type                | Description
--------- | ------------------- | -----------
`targets` | `list(map(string))` | The targets that can be used to collect `windows` metrics.

For example, the `targets` could either be passed to a `prometheus.relabel`
component to rewrite the metrics' label set, or to a `prometheus.scrape`
component that collects the exposed metrics.

## Component health

`prometheus.exporter.windows` is only reported as unhealthy if given
an invalid configuration. In those cases, exported fields retain their last
healthy values.

## Debug information

`prometheus.exporter.windows` does not expose any component-specific
debug information.

## Debug metrics

`prometheus.exporter.windows` does not expose any component-specific
debug metrics.

## Collectors list
The following table lists the available collectors that `windows_exporter` brings
bundled in. Some collectors only work on specific operating systems; enabling a
collector that is not supported by the host OS where Flow is running
is a no-op.

Users can choose to enable a subset of collectors to limit the amount of
metrics exposed by the `prometheus.exporter.windows` component,
or disable collectors that are expensive to run.


Name     | Description | Enabled by default
---------|-------------|--------------------
[ad](https://github.com/grafana/windows_exporter/blob/871715ba0b43c640257fb5ff6491b7420f23dcdd/docs/collector.ad.md) | Active Directory Domain Services |
[adfs](https://github.com/grafana/windows_exporter/blob/871715ba0b43c640257fb5ff6491b7420f23dcdd/docs/collector.adfs.md) | Active Directory Federation Services |
[cache](https://github.com/grafana/windows_exporter/blob/871715ba0b43c640257fb5ff6491b7420f23dcdd/docs/collector.cache.md) | Cache metrics |
[cpu](https://github.com/grafana/windows_exporter/blob/871715ba0b43c640257fb5ff6491b7420f23dcdd/docs/collector.cpu.md) | CPU usage | &#10003;
[cpu_info](https://github.com/grafana/windows_exporter/blob/871715ba0b43c640257fb5ff6491b7420f23dcdd/docs/collector.cpu_info.md) | CPU Information |
[cs](https://github.com/grafana/windows_exporter/blob/871715ba0b43c640257fb5ff6491b7420f23dcdd/docs/collector.cs.md) | "Computer System" metrics (system properties, num cpus/total memory) | &#10003;
[container](https://github.com/grafana/windows_exporter/blob/871715ba0b43c640257fb5ff6491b7420f23dcdd/docs/collector.container.md) | Container metrics |
[dfsr](https://github.com/grafana/windows_exporter/blob/871715ba0b43c640257fb5ff6491b7420f23dcdd/docs/collector.dfsr.md) | DFSR metrics |
[dhcp](https://github.com/grafana/windows_exporter/blob/871715ba0b43c640257fb5ff6491b7420f23dcdd/docs/collector.dhcp.md) | DHCP Server |
[dns](https://github.com/grafana/windows_exporter/blob/871715ba0b43c640257fb5ff6491b7420f23dcdd/docs/collector.dns.md) | DNS Server |
[exchange](https://github.com/grafana/windows_exporter/blob/871715ba0b43c640257fb5ff6491b7420f23dcdd/docs/collector.exchange.md) | Exchange metrics |
[fsrmquota](https://github.com/grafana/windows_exporter/blob/871715ba0b43c640257fb5ff6491b7420f23dcdd/docs/collector.fsrmquota.md) | Microsoft File Server Resource Manager (FSRM) Quotas collector |
[hyperv](https://github.com/grafana/windows_exporter/blob/871715ba0b43c640257fb5ff6491b7420f23dcdd/docs/collector.hyperv.md) | Hyper-V hosts |
[iis](https://github.com/grafana/windows_exporter/blob/871715ba0b43c640257fb5ff6491b7420f23dcdd/docs/collector.iis.md) | IIS sites and applications |
[logical_disk](https://github.com/grafana/windows_exporter/blob/871715ba0b43c640257fb5ff6491b7420f23dcdd/docs/collector.logical_disk.md) | Logical disks, disk I/O | &#10003;
[logon](https://github.com/grafana/windows_exporter/blob/871715ba0b43c640257fb5ff6491b7420f23dcdd/docs/collector.logon.md) | User logon sessions |
[memory](https://github.com/grafana/windows_exporter/blob/871715ba0b43c640257fb5ff6491b7420f23dcdd/docs/collector.memory.md) | Memory usage metrics |
[msmq](https://github.com/grafana/windows_exporter/blob/871715ba0b43c640257fb5ff6491b7420f23dcdd/docs/collector.msmq.md) | MSMQ queues |
[mssql](https://github.com/grafana/windows_exporter/blob/871715ba0b43c640257fb5ff6491b7420f23dcdd/docs/collector.mssql.md) | [SQL Server Performance Objects](https://docs.microsoft.com/en-us/sql/relational-databases/performance-monitor/use-sql-server-objects#SQLServerPOs) metrics  |
[netframework_clrexceptions](https://github.com/grafana/windows_exporter/blob/871715ba0b43c640257fb5ff6491b7420f23dcdd/docs/collector.netframework_clrexceptions.md) | .NET Framework CLR Exceptions |
[netframework_clrinterop](https://github.com/grafana/windows_exporter/blob/871715ba0b43c640257fb5ff6491b7420f23dcdd/docs/collector.netframework_clrinterop.md) | .NET Framework Interop Metrics |
[netframework_clrjit](https://github.com/grafana/windows_exporter/blob/871715ba0b43c640257fb5ff6491b7420f23dcdd/docs/collector.netframework_clrjit.md) | .NET Framework JIT metrics |
[netframework_clrloading](https://github.com/grafana/windows_exporter/blob/871715ba0b43c640257fb5ff6491b7420f23dcdd/docs/collector.netframework_clrloading.md) | .NET Framework CLR Loading metrics |
[netframework_clrlocksandthreads](https://github.com/grafana/windows_exporter/blob/871715ba0b43c640257fb5ff6491b7420f23dcdd/docs/collector.netframework_clrlocksandthreads.md) | .NET Framework locks and metrics threads |
[netframework_clrmemory](https://github.com/grafana/windows_exporter/blob/871715ba0b43c640257fb5ff6491b7420f23dcdd/docs/collector.netframework_clrmemory.md) |  .NET Framework Memory metrics |
[netframework_clrremoting](https://github.com/grafana/windows_exporter/blob/871715ba0b43c640257fb5ff6491b7420f23dcdd/docs/collector.netframework_clrremoting.md) | .NET Framework Remoting metrics |
[netframework_clrsecurity](https://github.com/grafana/windows_exporter/blob/871715ba0b43c640257fb5ff6491b7420f23dcdd/docs/collector.netframework_clrsecurity.md) | .NET Framework Security Check metrics |
[net](https://github.com/grafana/windows_exporter/blob/871715ba0b43c640257fb5ff6491b7420f23dcdd/docs/collector.net.md) | Network interface I/O | &#10003;
[os](https://github.com/grafana/windows_exporter/blob/871715ba0b43c640257fb5ff6491b7420f23dcdd/docs/collector.os.md) | OS metrics (memory, processes, users) | &#10003;
[process](https://github.com/grafana/windows_exporter/blob/871715ba0b43c640257fb5ff6491b7420f23dcdd/docs/collector.process.md) | Per-process metrics |
[remote_fx](https://github.com/grafana/windows_exporter/blob/871715ba0b43c640257fb5ff6491b7420f23dcdd/docs/collector.remote_fx.md) | RemoteFX protocol (RDP) metrics |
[service](https://github.com/grafana/windows_exporter/blob/871715ba0b43c640257fb5ff6491b7420f23dcdd/docs/collector.service.md) | Service state metrics | &#10003;
[smtp](https://github.com/grafana/windows_exporter/blob/871715ba0b43c640257fb5ff6491b7420f23dcdd/docs/collector.smtp.md) | IIS SMTP Server |
[system](https://github.com/grafana/windows_exporter/blob/871715ba0b43c640257fb5ff6491b7420f23dcdd/docs/collector.system.md) | System calls | &#10003;
[tcp](https://github.com/grafana/windows_exporter/blob/871715ba0b43c640257fb5ff6491b7420f23dcdd/docs/collector.tcp.md) | TCP connections |
[time](https://github.com/grafana/windows_exporter/blob/871715ba0b43c640257fb5ff6491b7420f23dcdd/docs/collector.time.md) | Windows Time Service |
[thermalzone](https://github.com/grafana/windows_exporter/blob/871715ba0b43c640257fb5ff6491b7420f23dcdd/docs/collector.thermalzone.md) | Thermal information
[terminal_services](https://github.com/grafana/windows_exporter/blob/871715ba0b43c640257fb5ff6491b7420f23dcdd/docs/collector.terminal_services.md) | Terminal services (RDS)
[textfile](https://github.com/grafana/windows_exporter/blob/871715ba0b43c640257fb5ff6491b7420f23dcdd/docs/collector.textfile.md) | Read prometheus metrics from a text file |
[vmware](https://github.com/grafana/windows_exporter/blob/871715ba0b43c640257fb5ff6491b7420f23dcdd/docs/collector.vmware.md) | Performance counters installed by the Vmware Guest agent |

See the linked documentation on each collector for more information on reported metrics, configuration settings and usage examples.

## Example

This example uses a [`prometheus.scrape` component][scrape] to collect metrics
from `prometheus.exporter.windows`:

```river
prometheus.exporter.windows "default" {
}

// Configure a prometheus.scrape component to collect windows metrics.
prometheus.scrape "example" {
  targets    = prometheus.exporter.windows.this.targets
  forward_to = [ /* ... */ ]
}
```

[scrape]: {{< relref "./prometheus.scrape.md" >}}
