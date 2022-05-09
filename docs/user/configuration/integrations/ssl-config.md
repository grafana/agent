+++
title = "ssl_exporter_config"
+++

# ssl config

The `ssl` block configures the `ssl` integration,
which is an embedded version of
[`ssl_exporter`](https://github.com/ribbybibby/ssl_exporter). This allows collection of SSL certificate metrics from hosts and files.


## Quick configuration example

To get started, define SSL targets in Grafana agent's integration block:

```yaml
metrics:
  wal_directory: /tmp/wal
integrations:
  ssl:
    enabled: true
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
  # Enables the ssl integration, allowing the Agent to automatically
  # collect metrics for the specified github objects.
  [enabled: <boolean> | default = false]

  # Sets an explicit value for the instance label when the integration is
  # self-scraped. Overrides inferred values.
  #
  # The default value for this integration is inferred from the hostname portion
  # of api_url.
  [instance: <string>]

  # Automatically collect metrics from this integration. If disabled,
  # the ssl integration will be run but not scraped and thus not
  # remote-written. Metrics for the integration will be exposed at
  # /integrations/ssl/metrics and can be scraped by an external
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

  # Either the filename or host to target
  [target: <string>]

  # SSL module (enum: tcp, https, file, kubernetes, kubeconfig)
  [module: <string> | default = "tcp"]
```

## About ssl_exporter Modules

For more information on the supported modules, refer to [ribbybibby/ssl_exporter](https://github.com/ribbybibby/ssl_exporter#configuration)
