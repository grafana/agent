---
aliases:
- ../../../../configuration/integrations/integrations-next/blackbox-config/
- /docs/grafana-cloud/monitor-infrastructure/agent/static/configuration/integrations/integrations-next/blackbox-config/
- /docs/grafana-cloud/send-data/agent/static/configuration/integrations/integrations-next/blackbox-config/
canonical: https://grafana.com/docs/agent/latest/static/configuration/integrations/integrations-next/blackbox-config/
description: Learn about blackbox_config next
title: blackbox_config next
---

# blackbox_config next

The `blackbox_config` block configures the `blackbox_exporter`
integration, which is an embedded version of
[`blackbox_exporter`](https://github.com/prometheus/blackbox_exporter). This allows
for the collection of blackbox metrics (probes) and exposing them as Prometheus metrics.

## Quick configuration example

To get started, define Blackbox targets in Grafana Agent's integration block:

```yaml
metrics:
  wal_directory: /tmp/wal
  configs:
    - name: default
integrations:
  blackbox:
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

    # Autoscrape interval and timeout.
    [scrape_interval: <duration> | default = <integrations.metrics.autoscrape.scrape_interval>]
    [scrape_timeout: <duration> | default = <integrations.metrics.autoscrape.scrape_timeout>]

  # An optional extra set of labels to add to metrics from the integration target. These
  # labels are only exposed via the integration service discovery HTTP API and
  # added when autoscrape is used. They will not be found directly on the metrics
  # page for an integration.
  extra_labels:
    [ <labelname>: <labelvalue> ... ]

  #
  # Exporter-specific configuration options
  #

  # blackbox configuration file with custom modules.
  # This field has precedence to the config defined in the blackbox_config block.
  # See https://github.com/prometheus/blackbox_exporter/blob/master/example.yml for more details how to generate custom blackbox.yml file.
  [config_file: <string> | default = ""]

  # Embedded blackbox configuration. You can specify your modules here instead of an external config file.
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
