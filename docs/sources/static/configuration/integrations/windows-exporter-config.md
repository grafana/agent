---
aliases:
- ../../../configuration/integrations/windows-exporter-config/
- /docs/grafana-cloud/monitor-infrastructure/agent/static/configuration/integrations/windows-exporter-config/
- /docs/grafana-cloud/send-data/agent/static/configuration/integrations/windows-exporter-config/
canonical: https://grafana.com/docs/agent/latest/static/configuration/integrations/windows-exporter-config/
description: Learn about windows_exporter_config
title: windows_exporter_config
---

# windows_exporter_config

The `windows_exporter_config` block configures the `windows_exporter`
integration, which is an embedded version of
[`windows_exporter`](https://github.com/grafana/windows_exporter). This allows
for the collection of Windows metrics and exposing them as Prometheus metrics.

Full reference of options:

```yaml
  # Enables the windows_exporter integration, allowing the Agent to automatically
  # collect system metrics from the local windows instance
  [enabled: <boolean> | default = false]

  # Sets an explicit value for the instance label when the integration is
  # self-scraped. Overrides inferred values.
  #
  # The default value for this integration is inferred from the agent hostname
  # and HTTP listen port, delimited by a colon.
  [instance: <string>]

  # Automatically collect metrics from this integration. If disabled,
  # the consul_exporter integration will be run but not scraped and thus not
  # remote-written. Metrics for the integration will be exposed at
  # /integrations/windows_exporter/metrics and can be scraped by an external
  # process.
  [scrape_integration: <boolean> | default = <integrations_config.scrape_integrations>]

  # How often should the metrics be collected? Defaults to
  # prometheus.global.scrape_interval.
  [scrape_interval: <duration> | default = <global_config.scrape_interval>]

  # The timeout before considering the scrape a failure. Defaults to
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

  #
  # Exporter-specific configuration options
  #

  # List of collectors to enable. Any non-experimental collector from the
  # embedded version of windows_exporter can be enabled here.
  [enabled_collectors: <string> | default = "cpu,cs,logical_disk,net,os,service,system"]

  # Settings for collectors which accept configuration. Settings specified here
  # are only used if the corresponding collector is enabled in
  # enabled_collectors.

  # Configuration for Exchange Mail Server
  exchange:
    # Comma-separated List of collectors to use. Defaults to all, if not specified.
    # Maps to collectors.exchange.enabled in windows_exporter
    [enabled_list: <string>]

  # Configuration for the IIS web server
  iis:
    # Regexp of sites to whitelist. Site name must both match whitelist and not match blacklist to be included.
    # Maps to collector.iis.site-whitelist in windows_exporter
    [site_whitelist: <string> | default = ".+"]

    # Regexp of sites to blacklist. Site name must both match whitelist and not match blacklist to be included.
    # Maps to collector.iis.site-blacklist in windows_exporter
    [site_blacklist: <string> | default = ""]

    # Regexp of apps to whitelist. App name must both match whitelist and not match blacklist to be included.
    # Maps to collector.iis.app-whitelist in windows_exporter
    [app_whitelist: <string> | default=".+"]

    # Regexp of apps to blacklist. App name must both match whitelist and not match blacklist to be included.
    # Maps to collector.iis.app-blacklist in windows_exporter
    [app_blacklist: <string> | default=".+"]

  # Configuration for reading metrics from a text files in a directory
  text_file:
    # Directory to read text files with metrics from.
    # Maps to collector.textfile.directory in windows_exporter
    [text_file_directory: <string> | default="C:\Program Files\windows_exporter\textfile_inputs"]

  # Configuration for SMTP metrics
  smtp:
    # Regexp of virtual servers to whitelist. Server name must both match whitelist and not match blacklist to be included.
    # Maps to collector.smtp.server-whitelist in windows_exporter
    [whitelist: <string> | default=".+"]

    # Regexp of virtual servers to blacklist. Server name must both match whitelist and not match blacklist to be included.
    # Maps to collector.smtp.server-blacklist in windows_exporter
    [blacklist: <string> | default=""]

  # Configuration for Windows Services
  service:
    # "WQL 'where' clause to use in WMI metrics query. Limits the response to the services you specify and reduces the size of the response.
    # Maps to collector.service.services-where in windows_exporter
    [where_clause: <string> | default=""]

  # Configuration for physical disk on Windows 
  physical_disk:
    # Regexp of volumes to include. Disk name must both match include and not match exclude to be included.
    # Maps to collector.logical_disk.disk-include in windows_exporter.
    [include: <string> | default=".+"]

    # Regexp of volumes to exclude. Disk name must both match include and not match exclude to be included.
    # Maps to collector.logical_disk.disk-exclude in windows_exporter.
    [exclude: <string> | default=".+"]

  # Configuration for Windows Processes
  process:
    # Regexp of processes to include. Process name must both match whitelist and not match blacklist to be included.
    # Maps to collector.process.whitelist in windows_exporter
    [whitelist: <string> | default=".+"]

    # Regexp of processes to exclude. Process name must both match whitelist and not match blacklist to be included.
    # Maps to collector.process.blacklist in windows_exporter
    [blacklist: <string> | default=""]

  # Configuration for NICs
  network:
    # Regexp of NIC's to whitelist. NIC name must both match whitelist and not match blacklist to be included.
    # Maps to collector.net.nic-whitelist in windows_exporter
    [whitelist: <string> | default=".+"]

    # Regexp of NIC's to blacklist. NIC name must both match whitelist and not match blacklist to be included.
    # Maps to collector.net.nic-blacklist in windows_exporter
    [blacklist: <string> | default=""]

  # Configuration for Microsoft SQL Server
  mssql:
    # Comma-separated list of mssql WMI classes to use.
    # Maps to collectors.mssql.classes-enabled in windows_exporter
    [enabled_classes: <string> | default="accessmethods,availreplica,bufman,databases,dbreplica,genstats,locks,memmgr,sqlstats,sqlerrors,transactions"]

  # Configuration for Microsoft Queue
  msqm:
    # WQL 'where' clause to use in WMI metrics query. Limits the response to the msmqs you specify and reduces the size of the response.
    # Maps to collector.msmq.msmq-where in windows_exporter
    [where_clause: <string> | default=""]

  # Configuration for disk information
  logical_disk:
    # Regexp of volumes to whitelist. Volume name must both match whitelist and not match blacklist to be included.
    # Maps to collector.logical_disk.volume-whitelist in windows_exporter
    [whitelist: <string> | default=".+"]

    # Regexp of volumes to blacklist. Volume name must both match whitelist and not match blacklist to be included.
    # Maps to collector.logical_disk.volume-blacklist in windows_exporter
    [blacklist: <string> | default=".+"]

  # Configuration for Windows Task Scheduler
  scheduled_task:
    # Regexp of tasks to include.
    [include: <string> | default ".+"]
    #Regexp of tasks to exclude.
    [exclude: <string> | default ""]
```
