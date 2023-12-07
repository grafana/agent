---
aliases:
- ../../../configuration/integrations/dnsmasq-exporter-config/
- /docs/grafana-cloud/monitor-infrastructure/agent/static/configuration/integrations/dnsmasq-exporter-config/
- /docs/grafana-cloud/send-data/agent/static/configuration/integrations/dnsmasq-exporter-config/
canonical: https://grafana.com/docs/agent/latest/static/configuration/integrations/dnsmasq-exporter-config/
description: Learn about dnsmasq_exporter_config
title: dnsmasq_exporter_config
---

# dnsmasq_exporter_config

The `dnsmasq_exporter_config` block configures the `dnsmasq_exporter` integration,
which is an embedded version of
[`dnsmasq_exporter`](https://github.com/google/dnsmasq_exporter). This allows for
the collection of metrics from dnsmasq servers.

Note that currently, an Agent can only collect metrics from a single dnsmasq
server. If you want to collect metrics from multiple servers, you can run
multiple Agents and add labels using `relabel_configs` to differentiate between
the servers:

```yaml
dnsmasq_exporter:
  enabled: true
  dnsmasq_address: dnsmasq-a:53
  relabel_configs:
  - source_labels: [__address__]
    target_label: instance
    replacement: dnsmasq-a
```

Full reference of options:

```yaml
  # Enables the dnsmasq_exporter integration, allowing the Agent to automatically
  # collect system metrics from the configured dnsmasq server address
  [enabled: <boolean> | default = false]

  # Sets an explicit value for the instance label when the integration is
  # self-scraped. Overrides inferred values.
  #
  # The default value for this integration is inferred from the dnsmasq_address
  # value.
  [instance: <string>]

  # Automatically collect metrics from this integration. If disabled,
  # the dnsmasq_exporter integration will be run but not scraped and thus not
  # remote-written. Metrics for the integration will be exposed at
  # /integrations/dnsmasq_exporter/metrics and can be scraped by an external
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

  # Address of the dnsmasq server in host:port form.
  [dnsmasq_address: <string> | default = "localhost:53"]

  # Path to the dnsmasq leases file. If this file doesn't exist, scraping
  # dnsmasq # will fail with an warning log message.
  [leases_path: <string> | default = "/var/lib/misc/dnsmasq.leases"]

  # Expose dnsmasq leases as metrics (high cardinality).
  [expose_leases: <boolean> | default = false]
```
