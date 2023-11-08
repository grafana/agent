---
aliases:
- ../../../configuration/integrations/squid-config/
- /docs/grafana-cloud/monitor-infrastructure/agent/static/configuration/integrations/squid-config/
- /docs/grafana-cloud/send-data/agent/static/configuration/integrations/squid-config/
canonical: https://grafana.com/docs/agent/latest/static/configuration/integrations/squid-config/
description: Learn about squid_config
title: squid_config
---

# squid_config

The `squid_config` block configures the `squid` integration,
which is an embedded version of a forked version of the [`Squid_exporter`](https://github.com/boynux/squid-exporter). This integration allows you to collect third-party [Squid](http://www.squid-cache.org/) metrics.

Full reference of options:

```yaml
  # Enables the Squid integration, allowing the Agent to automatically
  # collect metrics for the specified Squid instance.
  [enabled: <boolean> | default = false]

  # Sets an explicit value for the instance label when the integration is
  # self-scraped. Overrides inferred values.
  #
  # The default value for this integration is the configured host:port of the connection string.
  [instance: <string>]

  # Automatically collect metrics from this integration. If disabled,
  # the Squid integration is run but not scraped and thus not
  # remote-written. Metrics for the integration are exposed at
  # /integrations/squid/metrics and can be scraped by an external
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

  # Relabel metrics coming from the integration, lets you drop series
  # that you don't care about from the integration.
  metric_relabel_configs:
    [ - <relabel_config> ... ]

  # How frequently the WAL is truncated for this integration.
  [wal_truncate_frequency: <duration> | default = "60m"]

  #
  # Exporter-specific configuration options
  #

  # The address used to connect to the Squid instance in the format
  # of <HOST>:<PORT>.
  # i.e. "localhost:3128"
  [address: <string>]

  # The username for squid instance.
  [username: <string>]

  # The password for username above.
  [password: <string>]
```

## Configuration example

```yaml
integrations:
  squid:
    enabled: true
    address: localhost:3128
    scrape_interval: 1m
    scrape_timeout: 1m
    scrape_integration: true
metrics:
  wal_directory: /tmp/grafana-agent-wal
server:
  log_level: debug
```
