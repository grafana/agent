+++
title = "ssl_exporter_config"
+++

# ssl config

Use the `ssl` block to configure `ssl` integration.
'ssl integration' is an embedded version of
[`ssl_exporter`](https://github.com/ribbybibby/ssl_exporter). This enables the collection of SSL certificate metrics from hosts and files.


## Quick configuration example

Define SSL targets in Grafana Agent's integration block:

```yaml
prometheus:
  wal_directory: /tmp/wal
  configs:
    - name: default

integrations:
  ssl:
    ssl_targets:
      - name: consul
        target: /etc/consul/ssl/certs/*.c*rt
        module: file
      - name: nomad
        target: /etc/nomad/ssl/certs/*.c*rt
        module: file
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

  # SSL configuration file with custom modules.
  # See https://github.com/ribbybibby/ssl_exporter#configuration-file for more details how to generate custom config.file.
  # If not defined, embedded ssl_exporter default set of modules is used.
  [config_file: <string> | default = ""]

  # List of SSL targets to poll
  ssl_targets:
    [- <ssl_target> ... ]

```

## ssl_target config

```yaml
  # Name of a ssl_target
  [name: <string>]

  # The address of SSL device
  [target: <string>]

  # SSL module to use for polling
  [module: <string> | default = ""]
```


## About ssl_exporter Modules

For more information on the supported modules, refer to [ribbybibby/ssl_exporter](https://github.com/ribbybibby/ssl_exporter#configuration)
