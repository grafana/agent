---
aliases:
- ../../../configuration/integrations/blackbox-config/
- /docs/grafana-cloud/monitor-infrastructure/agent/static/configuration/integrations/blackbox-config/
- /docs/grafana-cloud/send-data/agent/static/configuration/integrations/blackbox-config/
canonical: https://grafana.com/docs/agent/latest/static/configuration/integrations/blackbox-config/
description: Learn about blackbox_config
title: blackbox_config
---

# blackbox_config

The `blackbox_config` block configures the `blackbox_exporter`
integration, which is an embedded version of
[`blackbox_exporter`](https://github.com/prometheus/blackbox_exporter). This allows
for the collection of blackbox metrics (probes) and exposing them as Prometheus metrics.

## Quick configuration example

To get started, define Blackbox targets in Grafana Agent's integration block:

```yaml
metrics:
  wal_directory: /tmp/wal
integrations:
  blackbox:
    enabled: true
    blackbox_targets:
      - name: example
        address: http://example.com
        module: http_2xx
    blackbox_config:
      modules:
        http_2xx:
          prober: http
          timeout: 5s
          http:
            method: POST
            headers:
              Content-Type: application/json
            body: '{}'
            preferred_ip_protocol: "ip4"
```

Full reference of options:

```yaml
  # Enables the blackbox_exporter integration, allowing the Agent to automatically
  # collect system metrics from the configured statsd server address
  [enabled: <boolean> | default = false]

  # Sets an explicit value for the instance label when the integration is
  # self-scraped. Overrides inferred values.
  #
  # The default value for this integration is inferred from the agent hostname
  # and HTTP listen port, delimited by a colon.
  [instance: <string>]

  # Automatically collect metrics from this integration. If disabled,
  # the blackbox_exporter integration will be run but not scraped and thus not
  # remote-written. Metrics for the integration will be exposed at
  # /integrations/blackbox/metrics and can be scraped by an external
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

  # blackbox configuration file with custom modules.
  # This field has precedence to the config defined in the blackbox_config block.
  # See https://github.com/prometheus/blackbox_exporter/blob/master/example.yml for more details how to generate custom blackbox.yml file.
  [config_file: <string> | default = ""]

  # Embedded blackbox configuration. You can specify your modules here instead of an external config file.
  # config_file or blackbox_config must be specified.
  # See https://github.com/prometheus/blackbox_exporter/blob/master/CONFIGURATION.md for more details how to specify your blackbox modules.
  blackbox_config:
    [- <modules> ... ]

  # List of targets to probe
  blackbox_targets:
    [- <blackbox_target> ... ]

  # Option to configure blackbox_exporter.
  # Represents the offset to subtract from timeout in seconds when probing targets.
  [probe_timeout_offset: <float> | default = 0.5]
```
## blackbox_target config

```yaml
  # Name of a blackbox_target
  [name: <string>]

  # The address of the target to probe
  [address: <string>]

  # Blackbox module to use to probe
  [module: <string> | default = ""]
```
