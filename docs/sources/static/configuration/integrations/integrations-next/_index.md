---
aliases:
- ../../../configuration/integrations/integrations-next/
- /docs/grafana-cloud/monitor-infrastructure/agent/static/configuration/integrations/integrations-next/
- /docs/grafana-cloud/send-data/agent/static/configuration/integrations/integrations-next/
canonical: https://grafana.com/docs/agent/latest/static/configuration/integrations/integrations-next/
description: Learn about integrations next
menuTitle: Integrations next
title: Integrations next (Experimental)
weight: 100
---

# Integrations next (Experimental)

Release v0.22.0 of Grafana Agent includes experimental support for a revamped
integrations subsystem. The integrations subsystem is the second oldest part of
Grafana Agent, and has started to feel out of place as we built out the
project.

The revamped integrations subsystem can be enabled by passing
`integrations-next` to the `-enable-features` command line flag. As an
experimental feature, there are no stability guarantees, and it may receive a
higher frequency of breaking changes than normal.

The revamped integrations subsystem has the following benefits over the
original subsystem:

* Integrations can opt in to supporting multiple instances. For example, you
  may now run any number of `redis_exporter` integrations, where before you
  could only have one per agent. Integrations such as `node_exporter` still
  only support a single instance, as it wouldn't make sense to have multiple
  instances of those.

* Autoscrape (previously called "self-scraping"), when enabled, now supports
  sending metrics for an integration directly to a running metrics instance.
  This allows you configuring an integration to send to a specific Prometheus
  remote_write endpoint.

* A new service discovery HTTP API is included. This can be used with
  Prometheus' [http_sd_config][http_sd_config]. The API returns extra labels
  for integrations that previously were only available when autoscraping, such
  as `agent_hostname`.

* Integrations that aren't Prometheus exporters may now be added, such as
  integrations that generate logs or traces.

* Autoscrape, when enabled, now works completely in-memory without using the
  network.

[http_sd_config]: https://prometheus.io/docs/prometheus/2.45/configuration/configuration/#http_sd_config

## Config changes

The revamp contains a number of breaking changes to the config. The schema of the
`integrations` key in the config file is now the following:

```yaml
integrations:
  # Controls settings for integrations that generate metrics.
  metrics:
    # Controls default settings for autoscrape. Individual instances of
    # integrations inherit the defaults and may override them.
    autoscrape:
      # Enables autoscrape of integrations.
      [enable: <boolean> | default = true]

      # Specifies the metrics instance name to send metrics to. Instance
      # names are located at metrics.configs[].name from the top-level config.
      # The instance must exist.
      #
      # As it is common to use the name "default" for your primary instance,
      # we assume the same here.
      [metrics_instance: <string> | default = "default"]

      # Autoscrape interval and timeout. Defaults are inherited from the global
      # section of the top-level metrics config.
      [scrape_interval: <duration> | default = <metrics.global.scrape_interval>]
      [scrape_timeout: <duration> | default = <metrics.global.scrape_timeout>]

  # Configs for integrations which do not support multiple instances.
  [agent: <agent_config>]
  [cadvisor: <cadvisor_config>]
  [node_exporter: <node_exporter_config>]
  [process: <process_exporter_config>]
  [statsd: <statsd_exporter_config>]
  [windows: <windows_exporter_config>]
  [eventhandler: <eventhandler_config>]
  [snmp: <snmp_exporter_config>]
  [blackbox: <blackbox_config>]

  # Configs for integrations that do support multiple instances. Note that
  # these must be arrays.
  consul_configs:
    [- <consul_exporter_config> ...]

  dnsmasq_configs:
    [- <dnsmasq_exporter_config> ...]

  elasticsearch_configs:
    [- <elasticsearch_exporter_config> ...]

  github_configs:
    [- <github_exporter_config> ...]

  kafka_configs:
    [- <kafka_exporter_config> ...]

  memcached_configs:
    [- <memcached_exporter_config> ...]

  mongodb_configs:
    [- <mongodb_exporter_config> ...]

  mssql_configs:
    [- <mssql_config> ...]

  mysql_configs:
    [- <mysqld_exporter_config> ...]

  oracledb_configs:
    [ - <oracledb_exporter_config> ...]

  postgres_configs:
    [- <postgres_exporter_config> ...]

  redis_configs:
    [- <redis_exporter_config> ...]

  snowflake_configs:
    [- <snowflake_config> ...]

  app_agent_receiver_configs:
    [- <app_agent_receiver_config>]

  apache_http_configs:
    [- <apache_http_config>]

  squid_configs:
    [- <squid_config> ...]

  vsphere_configs:
    [- <vsphere_config>]

  gcp_configs:
    [- <gcp_config>]
    
  azure_configs:
    [- <azure_config>]   
    
  cloudwatch_configs:
    [- <cloudwatch_config>]    
```

Note that most integrations are no longer configured with the `_exporter` name.
`node_exporter` is the only integration with `_exporter` name due to its
popularity in the Prometheus ecosystem.

## Integrations changes

Integrations no longer support an `enabled` field; they are enabled by being
defined in the YAML. To disable an integration, comment it out or remove it.

Metrics-based integrations now use this common set of options:

```yaml
# Provide an explicit value to uniquely identify this instance of the
# integration. If not provided, a reasonable default will be inferred based
# on the integration.
#
# The value here must be unique across all instances of the same integration.
[instance: <string>]

# Override autoscrape defaults for this integration.
autoscrape:
  # Enables autoscrape of integrations.
  [enable: <boolean> | default = <integrations.metrics.autoscrape.enable>]

  # Specifies the metrics instance name to send metrics to.
  [metrics_instance: <string> | default = <integrations.metrics.autoscrape.metrics_instance>]

  # Relabel the autoscrape job.
  relabel_configs:
    [- <relabel_config> ... ]

  # Relabel metrics coming from the integration.
  metric_relabel_configs:
    [ - <relabel_config> ... ]

  # Autoscrape interval and timeout.
  [scrape_interval: <duration> | default = <integrations.metrics.autoscrape.scrape_interval>]
  [scrape_timeout: <duration> | default = <integrations.metrics.autoscrape.scrape_timeout>]

# An optional extra set of labels to add to metrics from the integration target. These
# labels are only exposed via the integration service discovery HTTP API and
# added when autoscrape is used. They will not be found directly on the metrics
# page for an integration.
extra_labels:
  [ <labelname>: <labelvalue> ... ]
```

The old set of common options have been removed and do not work when the revamp
is being used:

```yaml
# OLD SCHEMA: NO LONGER SUPPORTED

[enabled: <boolean> | default = false]
[instance: <string>]
[scrape_integration: <boolean> | default = <integrations_config.scrape_integrations>]
[scrape_interval: <duration> | default = <global_config.scrape_interval>]
[scrape_timeout: <duration> | default = <global_config.scrape_timeout>]
[wal_truncate_frequency: <duration> | default = "60m"]
relabel_configs:
  [- <relabel_config> ...]
metric_relabel_configs:
  [ - <relabel_config> ...]
```
